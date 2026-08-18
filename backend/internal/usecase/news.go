package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"cashfac-test/internal/domain"
)

type NewsUseCase struct {
	repo     domain.NewsRepository
	source   domain.SourceClient
	rewriter domain.Rewriter
}

type SyncProgressFunc func(processed, total int)

func NewNewsUseCase(repo domain.NewsRepository, source domain.SourceClient, rewriter domain.Rewriter) *NewsUseCase {
	return &NewsUseCase{
		repo:     repo,
		source:   source,
		rewriter: rewriter,
	}
}

func (uc *NewsUseCase) Sync(ctx context.Context, limit int, mood domain.Mood) (int, error) {
	return uc.sync(ctx, limit, mood, nil)
}

func (uc *NewsUseCase) SyncWithProgress(ctx context.Context, limit int, mood domain.Mood, progressFn SyncProgressFunc) (int, error) {
	return uc.sync(ctx, limit, mood, progressFn)
}

func (uc *NewsUseCase) sync(ctx context.Context, limit int, mood domain.Mood, progressFn SyncProgressFunc) (int, error) {
	items, err := uc.source.FetchLatest(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("fetch latest news: %w", err)
	}

	now := time.Now().UTC()
	savedCount := 0

	for _, item := range items {
		rewritten, err := uc.rewriter.Rewrite(ctx, domain.RewriteRequest{
			Title: item.Title,
			Text:  item.Text,
			Mood:  mood,
		})
		if err != nil {
			return savedCount, fmt.Errorf("rewrite news %s: %w", item.ExternalID, err)
		}

		publishedAt, _ := time.Parse(time.RFC3339, item.PublishedRaw)
		originalDigest := checksum(item.Text)
		newsItem := domain.News{
			ID:             checksum(item.ExternalID + string(mood)),
			Title:          item.Title,
			OriginalText:   item.Text,
			RewrittenText:  rewritten.Text,
			Mood:           mood,
			SourceName:     item.SourceName,
			SourceURL:      item.SourceURL,
			PublishedAt:    publishedAt,
			CreatedAt:      now,
			ExternalID:     item.ExternalID,
			FactChecksum:   rewritten.FactChecksum,
			OriginalDigest: originalDigest,
		}

		if err := uc.repo.Save(ctx, newsItem); err != nil {
			return savedCount, fmt.Errorf("save news %s: %w", item.ExternalID, err)
		}

		savedCount++
		if progressFn != nil {
			progressFn(savedCount, len(items))
		}
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

func (uc *NewsUseCase) RewriteByExternalID(ctx context.Context, externalID string, mood domain.Mood) (domain.News, error) {
	if mood == "" {
		mood = domain.MoodNeutral
	}

	item, err := uc.repo.GetByExternalIDAndMood(ctx, externalID, mood)
	if err == nil {
		if shouldReuseRewrite(item, mood) {
			return item, nil
		}
	}
	if !errors.Is(err, domain.ErrNewsNotFound) {
		return domain.News{}, fmt.Errorf("get news by external id and mood: %w", err)
	}

	baseItem, err := uc.repo.GetByExternalID(ctx, externalID)
	if err != nil {
		return domain.News{}, fmt.Errorf("get base news by external id: %w", err)
	}

	rewritten, err := uc.rewriter.Rewrite(ctx, domain.RewriteRequest{
		Title: baseItem.Title,
		Text:  baseItem.OriginalText,
		Mood:  mood,
	})
	if err != nil {
		return domain.News{}, fmt.Errorf("rewrite news by external id %s: %w", externalID, err)
	}

	item = domain.News{
		ID:             checksum(baseItem.ExternalID + string(mood)),
		Title:          baseItem.Title,
		OriginalText:   baseItem.OriginalText,
		RewrittenText:  rewritten.Text,
		Mood:           mood,
		SourceName:     baseItem.SourceName,
		SourceURL:      baseItem.SourceURL,
		PublishedAt:    baseItem.PublishedAt,
		CreatedAt:      time.Now().UTC(),
		ExternalID:     baseItem.ExternalID,
		FactChecksum:   rewritten.FactChecksum,
		OriginalDigest: baseItem.OriginalDigest,
	}

	if err := uc.repo.Save(ctx, item); err != nil {
		return domain.News{}, fmt.Errorf("save rewritten news by external id %s: %w", externalID, err)
	}

	return item, nil
}

func checksum(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func shouldReuseRewrite(item domain.News, mood domain.Mood) bool {
	trimmed := strings.TrimSpace(item.RewrittenText)
	if trimmed == "" {
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

	if tokenOverlapRatio(normalizedOriginal, normalizedRewrite) > 0.78 {
		return false
	}

	return true
}

func normalizeRewriteText(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(value)), " ")
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
