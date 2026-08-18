package rewriter

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	body, err := json.Marshal(zaiChatCompletionRequest{
		Model: c.model,
		Messages: []zaiMessage{
			{
				Role:    "system",
				Content: buildSystemPrompt(),
			},
			{
				Role: "user",
				Content: buildUserPrompt(request),
			},
		},
		Temperature: 0.4,
		MaxTokens:   estimateMaxTokens(request.Text),
		Thinking: zaiThinking{
			Type: "disabled",
		},
	})
	if err != nil {
		return domain.RewriteResponse{}, fmt.Errorf("marshal zai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return domain.RewriteResponse{}, fmt.Errorf("build zai request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.RewriteResponse{}, fmt.Errorf("perform zai request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return domain.RewriteResponse{}, fmt.Errorf("zai api status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var completion zaiChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
		return domain.RewriteResponse{}, fmt.Errorf("decode zai response: %w", err)
	}

	if len(completion.Choices) == 0 {
		return domain.RewriteResponse{}, fmt.Errorf("zai response has no choices")
	}

	text := strings.TrimSpace(completion.Choices[0].Message.Content)
	if text == "" {
		return domain.RewriteResponse{}, fmt.Errorf("zai response is empty")
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
		"Return only the rewritten text with no explanations, no markdown, and no preface.",
	}, " ")
}

func buildUserPrompt(request domain.RewriteRequest) string {
	return fmt.Sprintf(
		"Rewrite the news article in %q tone while preserving all facts exactly.\n\nTitle: %s\n\nArticle:\n%s",
		request.Mood,
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
