package rewriter

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"cashfac-test/internal/domain"
)

type ZAIClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewZAIClient(apiKey, baseURL, model string) *ZAIClient {
	return &ZAIClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *ZAIClient) Rewrite(ctx context.Context, request domain.RewriteRequest) (domain.RewriteResponse, error) {
	text, err := c.completeRewrite(ctx, request, false)
	if err != nil {
		return domain.RewriteResponse{}, err
	}

	if isWeakRewrite(request, text) {
		log.Printf("rewriter: weak rewrite detected for mood=%s title=%q, retry with stricter prompt", request.Mood, request.Title)
		text, err = c.completeRewrite(ctx, request, true)
		if err != nil {
			return domain.RewriteResponse{}, err
		}
	}

	if isWeakRewrite(request, text) {
		log.Printf("rewriter: second weak rewrite detected for mood=%s title=%q, retry with final prompt", request.Mood, request.Title)
		text, err = c.completeRewrite(ctx, request, true)
		if err != nil {
			return domain.RewriteResponse{}, err
		}
	}

	if isWeakRewrite(request, text) {
		return domain.RewriteResponse{}, fmt.Errorf("rewriter could not prepare a distinct enough version for this mood")
	}

	return domain.RewriteResponse{
		Text:         text,
		FactChecksum: checksum(request.Title + ":" + request.Text),
	}, nil
}

func buildSystemPrompt() string {
	return strings.Join([]string{
		"You rewrite news articles into a requested emotional tone.",
		"Do not change any facts.",
		"Keep all names, dates, locations, numbers, organizations, quotes, and sequence of events intact.",
		"Do not invent details or omit important factual details.",
		"You must substantially change wording, sentence flow, and narrative voice.",
		"Do not return the original article with tiny edits.",
		"Do not add labels like [sad], [happy], [ironic], or [neutral].",
		"Return only the rewritten text with no explanations, no markdown, and no preface.",
	}, " ")
}

func buildUserPrompt(request domain.RewriteRequest) string {
	styleInstruction := moodInstruction(request.Mood)
	structureInstruction := articleStructureInstruction(request)

	return fmt.Sprintf(
		"Rewrite this article in %q tone while preserving all facts exactly.\n"+
			"Required style: %s\n"+
			"Structure guidance: %s\n"+
			"Important: keep every factual detail, but noticeably rewrite the phrasing so the tone is obvious to a reader.\n"+
			"Important: the wording must feel materially different from the source, not just lightly edited.\n"+
			"Important: do not use bracket labels, headings, bullet points, or explanations.\n\n"+
			"Title: %s\n\nArticle:\n%s",
		request.Mood,
		styleInstruction,
		structureInstruction,
		request.Title,
		request.Text,
	)
}

func buildStrictUserPrompt(request domain.RewriteRequest) string {
	return fmt.Sprintf(
		"Your previous rewrite was too close to the source. Try again.\n"+
			"Rewrite the article so that the emotional tone is immediately noticeable, but every fact remains unchanged.\n"+
			"Change sentence structure, transitions, emphasis, diction, and cadence.\n"+
			"Keep all names, dates, quotes, numbers, locations, and the order of events.\n"+
			"Do not shorten the article drastically.\n"+
			"Do not preserve the original opening sentence or the original paragraph rhythm.\n"+
			"Do not mirror the source wording except for unavoidable factual phrases, names, and quotes.\n"+
			"Each paragraph should feel freshly written, not lightly edited.\n"+
			"Do not add labels such as [sad] or any intro.\n"+
			"Tone requirement: %s\n"+
			"Structure guidance: %s\n\n"+
			"Title: %s\n\nArticle:\n%s",
		moodInstruction(request.Mood),
		articleStructureInstruction(request),
		request.Title,
		request.Text,
	)
}

func checksum(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

type zaiChatCompletionRequest struct {
	Model       string       `json:"model"`
	Messages    []zaiMessage `json:"messages"`
	Temperature float64      `json:"temperature"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Thinking    zaiThinking  `json:"thinking,omitempty"`
}

type zaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type zaiThinking struct {
	Type string `json:"type"`
}

type zaiChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *ZAIClient) completeRewrite(ctx context.Context, request domain.RewriteRequest, strict bool) (string, error) {
	userPrompt := buildUserPrompt(request)
	temperature := 0.55
	if strict {
		userPrompt = buildStrictUserPrompt(request)
		temperature = 0.8
	}

	body, err := json.Marshal(zaiChatCompletionRequest{
		Model: c.model,
		Messages: []zaiMessage{
			{
				Role:    "system",
				Content: buildSystemPrompt(),
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: temperature,
		MaxTokens:   estimateMaxTokens(request.Text),
		Thinking: zaiThinking{
			Type: "disabled",
		},
	})
	if err != nil {
		return "", fmt.Errorf("marshal zai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build zai request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("perform zai request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return "", fmt.Errorf("zai api status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var completion zaiChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
		return "", fmt.Errorf("decode zai response: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("zai response has no choices")
	}

	text := strings.TrimSpace(completion.Choices[0].Message.Content)
	if text == "" {
		return "", fmt.Errorf("zai response is empty")
	}

	return text, nil
}

func moodInstruction(mood domain.Mood) string {
	switch mood {
	case domain.MoodHappy:
		return "sound lighter, warmer, and more hopeful without joking or changing the seriousness of the facts"
	case domain.MoodSad:
		return "sound heavier, more somber, and more mournful while remaining factual and restrained"
	case domain.MoodIronic:
		return "sound dry, sharp, and mildly ironic, but never alter the facts or fabricate sarcasm around quotes"
	case domain.MoodNeutral:
		return "sound calm, clear, and polished, like a clean newsroom rewrite"
	default:
		return "adapt the tone clearly while preserving every fact"
	}
}

func isWeakRewrite(request domain.RewriteRequest, rewritten string) bool {
	trimmed := strings.TrimSpace(rewritten)
	if trimmed == "" {
		return true
	}

	lowerTrimmed := strings.ToLower(trimmed)
	if strings.HasPrefix(lowerTrimmed, "["+string(request.Mood)+"]") {
		return true
	}

	normalizedOriginal := normalizeForComparison(request.Text)
	normalizedRewritten := normalizeForComparison(trimmed)
	if normalizedOriginal == normalizedRewritten {
		return true
	}

	if sharedPrefixRatio(normalizedOriginal, normalizedRewritten) > 0.82 {
		return true
	}

	if tokenOverlapRatio(normalizedOriginal, normalizedRewritten) > 0.74 {
		return true
	}

	return false
}

var nonLetterNumberPattern = regexp.MustCompile(`[^\p{L}\p{N}\s]+`)

func normalizeForComparison(value string) string {
	normalized := strings.ToLower(value)
	normalized = nonLetterNumberPattern.ReplaceAllString(normalized, " ")
	normalized = strings.Join(strings.Fields(normalized), " ")
	return normalized
}

func sharedPrefixRatio(a, b string) float64 {
	if a == "" || b == "" {
		return 0
	}

	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}

	matched := 0
	for matched < limit && a[matched] == b[matched] {
		matched++
	}

	base := len(a)
	if len(b) < base {
		base = len(b)
	}
	if base == 0 {
		return 0
	}

	return float64(matched) / float64(base)
}

func articleStructureInstruction(request domain.RewriteRequest) string {
	lowerTitle := strings.ToLower(request.Title)
	if strings.Contains(lowerTitle, "review") {
		return "rewrite it like a fresh review in the requested mood, with a clearly different critical voice and sentence rhythm"
	}

	return "rewrite it like a fresh article in the requested mood, with a clearly different narrative flow"
}

func tokenOverlapRatio(a, b string) float64 {
	if a == "" || b == "" {
		return 0
	}

	aTokens := strings.Fields(a)
	bTokens := strings.Fields(b)
	if len(aTokens) == 0 || len(bTokens) == 0 {
		return 0
	}

	bagA := make(map[string]int, len(aTokens))
	bagB := make(map[string]int, len(bTokens))
	for _, token := range aTokens {
		bagA[token]++
	}
	for _, token := range bTokens {
		bagB[token]++
	}

	intersection := 0
	union := 0
	seen := make(map[string]struct{}, len(bagA)+len(bagB))

	for token, countA := range bagA {
		countB := bagB[token]
		if countA < countB {
			intersection += countA
		} else {
			intersection += countB
		}
		if countA > countB {
			union += countA
		} else {
			union += countB
		}
		seen[token] = struct{}{}
	}

	for token, countB := range bagB {
		if _, ok := seen[token]; ok {
			continue
		}
		union += countB
	}

	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

func estimateMaxTokens(text string) int {
	runeCount := len([]rune(text))
	if runeCount == 0 {
		return 1024
	}

	estimated := runeCount/3 + 256
	if estimated < 512 {
		return 512
	}

	if estimated > 8192 {
		return 8192
	}

	return estimated
}
