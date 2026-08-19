package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"cashfac-test/internal/domain"
	"cashfac-test/internal/platform/logger"
)

type NewsUseCase struct {
	repo         domain.NewsRepository
	source       domain.SourceClient
	rewriter     domain.Rewriter
	rewriteMu    sync.Mutex
	rewriteCalls map[string]*rewriteCall
}

type rewriteCall struct {
	done   chan struct{}
	result domain.RewriteResult
	err    error
}

type SyncProgressFunc func(processed, total int)

func NewNewsUseCase(repo domain.NewsRepository, source domain.SourceClient, rewriter domain.Rewriter) *NewsUseCase {
	return &NewsUseCase{
		repo:         repo,
		source:       source,
		rewriter:     rewriter,
		rewriteCalls: make(map[string]*rewriteCall),
	}
}

func (uc *NewsUseCase) Sync(ctx context.Context, limit int) (int, error) {
	return uc.sync(ctx, limit, nil)
}

func (uc *NewsUseCase) SyncWithProgress(ctx context.Context, limit int, progressFn SyncProgressFunc) (int, error) {
	return uc.sync(ctx, limit, progressFn)
}

func (uc *NewsUseCase) sync(ctx context.Context, limit int, progressFn SyncProgressFunc) (int, error) {
	startedAt := time.Now()
	logger.Info("SOURCE", "fetching latest news", logger.F("limit", limit))
	items, err := uc.source.FetchLatest(ctx, limit)
	if err != nil {
		logger.Error("SOURCE", "failed to fetch latest news", logger.F("duration", logger.Duration(time.Since(startedAt))), logger.F("error", err))
		return 0, fmt.Errorf("fetch latest news: %w", err)
	}
	logger.Success("SOURCE", "latest news received",
		logger.F("items", len(items)),
		logger.F("duration", logger.Duration(time.Since(startedAt))),
	)

	now := time.Now().UTC()
	savedCount := 0
	externalIDs := make([]string, 0, len(items))

	for _, item := range items {
		externalIDs = append(externalIDs, item.ExternalID)

		publishedAt, _ := time.Parse(time.RFC3339, item.PublishedRaw)
		originalDigest := checksum(item.Text)
		newsItem := domain.News{
			ID:             checksum(item.ExternalID + string(domain.MoodNeutral)),
			Title:          item.Title,
			OriginalText:   item.Text,
			Mood:           domain.MoodNeutral,
			SourceName:     item.SourceName,
			SourceURL:      item.SourceURL,
			ImageURL:       item.ImageURL,
			PublishedAt:    publishedAt,
			CreatedAt:      now,
			ExternalID:     item.ExternalID,
			OriginalDigest: originalDigest,
		}

		existing, getErr := uc.repo.GetByExternalIDAndMood(ctx, item.ExternalID, domain.MoodNeutral)
		if getErr == nil && existing.OriginalDigest == originalDigest {
			newsItem.RewrittenText = existing.RewrittenText
			newsItem.FactChecksum = existing.FactChecksum
		} else if getErr != nil && !errors.Is(getErr, domain.ErrNewsNotFound) {
			return savedCount, fmt.Errorf("get existing news %s: %w", item.ExternalID, getErr)
		}

		if err := uc.repo.Save(ctx, newsItem); err != nil {
			return savedCount, fmt.Errorf("save news %s: %w", item.ExternalID, err)
		}

		savedCount++
		if progressFn != nil {
			progressFn(savedCount, len(items))
		}
	}

	if err := uc.repo.PruneByExternalIDs(ctx, externalIDs); err != nil {
		return savedCount, fmt.Errorf("prune stale news: %w", err)
	}

	return savedCount, nil
}

func (uc *NewsUseCase) List(ctx context.Context, mood domain.Mood) ([]domain.News, error) {
	return uc.repo.List(ctx, mood)
}

func (uc *NewsUseCase) Get(ctx context.Context, id string) (domain.News, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *NewsUseCase) GetByExternalID(ctx context.Context, externalID string, mood domain.Mood) (domain.News, error) {
	if mood != "" {
		item, err := uc.repo.GetByExternalIDAndMood(ctx, externalID, mood)
		if err == nil {
			return item, nil
		}
		if !errors.Is(err, domain.ErrNewsNotFound) {
			return domain.News{}, fmt.Errorf("get news by external id and mood: %w", err)
		}
	}

	item, err := uc.repo.GetByExternalID(ctx, externalID)
	if err != nil {
		return domain.News{}, fmt.Errorf("get news by external id: %w", err)
	}

	return item, nil
}

func (uc *NewsUseCase) RewriteByExternalID(ctx context.Context, externalID string, mood domain.Mood) (domain.RewriteResult, error) {
	startedAt := time.Now()

	if mood == "" {
		mood = domain.MoodNeutral
	}
	if !mood.IsValid() {
		return domain.RewriteResult{}, fmt.Errorf("unsupported mood %q", mood)
	}

	key := externalID + ":" + string(mood)
	uc.rewriteMu.Lock()
	if call, ok := uc.rewriteCalls[key]; ok {
		uc.rewriteMu.Unlock()
		select {
		case <-ctx.Done():
			logger.Warn("LLM", "rewrite wait cancelled", logger.F("article", logArticleID(externalID)), logger.F("mood", mood))
			return domain.RewriteResult{}, ctx.Err()
		case <-call.done:
			result := call.result
			result.Meta.Source = domain.RewriteSourceShared
			result.Meta.LLMRequests = 0
			result.Meta.SavedLLMRequests = 1
			result.Meta.DurationMs = time.Since(startedAt).Milliseconds()
			logRewriteResult(externalID, mood, result, call.err)
			return result, call.err
		}
	}

	call := &rewriteCall{done: make(chan struct{})}
	uc.rewriteCalls[key] = call
	uc.rewriteMu.Unlock()

	call.result, call.err = uc.rewriteByExternalID(ctx, externalID, mood)
	call.result.Meta.DurationMs = time.Since(startedAt).Milliseconds()

	uc.rewriteMu.Lock()
	close(call.done)
	delete(uc.rewriteCalls, key)
	uc.rewriteMu.Unlock()

	logRewriteResult(externalID, mood, call.result, call.err)
	return call.result, call.err
}

func logRewriteResult(externalID string, mood domain.Mood, result domain.RewriteResult, err error) {
	fields := []logger.Field{
		logger.F("article", logArticleID(externalID)),
		logger.F("mood", mood),
		logger.F("source", result.Meta.Source),
		logger.F("llm_requests", result.Meta.LLMRequests),
		logger.F("saved_requests", result.Meta.SavedLLMRequests),
		logger.F("duration", logger.Duration(time.Duration(result.Meta.DurationMs)*time.Millisecond)),
	}
	if err != nil {
		logger.Error("LLM", "rewrite failed", append(fields, logger.F("error", err))...)
		return
	}

	message := "rewrite generated"
	if result.Meta.Source == domain.RewriteSourceCache {
		message = "rewrite served from cache"
	} else if result.Meta.Source == domain.RewriteSourceShared {
		message = "in-flight rewrite reused"
	}
	logger.Success("LLM", message, fields...)
}

func logArticleID(externalID string) string {
	return checksum(externalID)[:10]
}

func (uc *NewsUseCase) rewriteByExternalID(ctx context.Context, externalID string, mood domain.Mood) (domain.RewriteResult, error) {
	baseItem, err := uc.repo.GetByExternalIDAndMood(ctx, externalID, domain.MoodNeutral)
	if errors.Is(err, domain.ErrNewsNotFound) {
		baseItem, err = uc.repo.GetByExternalID(ctx, externalID)
	}
	if err != nil {
		return domain.RewriteResult{}, fmt.Errorf("get base news by external id: %w", err)
	}

	originalDigest := checksum(baseItem.OriginalText)

	item, err := uc.repo.GetByExternalIDAndMood(ctx, externalID, mood)
	if err == nil && shouldReuseRewrite(item, mood, originalDigest) {
		return domain.RewriteResult{
			News: item,
			Meta: domain.RewriteMeta{
				Source:           domain.RewriteSourceCache,
				SavedLLMRequests: 1,
			},
		}, nil
	}
	if err != nil && !errors.Is(err, domain.ErrNewsNotFound) {
		return domain.RewriteResult{}, fmt.Errorf("get news by external id and mood: %w", err)
	}

	rewritten, err := uc.rewriter.Rewrite(ctx, domain.RewriteRequest{
		Title: baseItem.Title,
		Text:  baseItem.OriginalText,
		Mood:  mood,
	})
	if err != nil {
		return domain.RewriteResult{}, fmt.Errorf("rewrite news by external id %s: %w", externalID, err)
	}

	item = domain.News{
		ID:             checksum(baseItem.ExternalID + string(mood)),
		Title:          baseItem.Title,
		OriginalText:   baseItem.OriginalText,
		RewrittenText:  rewritten.Text,
		Mood:           mood,
		SourceName:     baseItem.SourceName,
		SourceURL:      baseItem.SourceURL,
		ImageURL:       baseItem.ImageURL,
		PublishedAt:    baseItem.PublishedAt,
		CreatedAt:      time.Now().UTC(),
		ExternalID:     baseItem.ExternalID,
		FactChecksum:   rewritten.FactChecksum,
		OriginalDigest: originalDigest,
	}

	if err := uc.repo.Save(ctx, item); err != nil {
		return domain.RewriteResult{}, fmt.Errorf("save rewritten news by external id %s: %w", externalID, err)
	}

	return domain.RewriteResult{
		News: item,
		Meta: domain.RewriteMeta{
			Source:      domain.RewriteSourceGenerated,
			LLMRequests: 1,
		},
	}, nil
}

func checksum(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func shouldReuseRewrite(item domain.News, mood domain.Mood, originalDigest string) bool {
	trimmed := strings.TrimSpace(item.RewrittenText)
	if trimmed == "" {
		return false
	}
	if item.OriginalDigest != originalDigest {
		return false
	}

	if strings.HasPrefix(strings.ToLower(trimmed), "["+string(mood)+"]") {
		return false
	}

	normalizedRewrite := normalizeRewriteText(trimmed)
	normalizedOriginal := normalizeRewriteText(item.OriginalText)
	if normalizedRewrite == normalizedOriginal {
		return false
	}

	return true
}

func normalizeRewriteText(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(value)), " ")
}
