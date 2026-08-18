package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"cashfac-test/internal/domain"
)

type NewsUseCase struct {
	repo     domain.NewsRepository
	source   domain.SourceClient
	rewriter domain.Rewriter
}

func NewNewsUseCase(repo domain.NewsRepository, source domain.SourceClient, rewriter domain.Rewriter) *NewsUseCase {
	return &NewsUseCase{
		repo:     repo,
		source:   source,
		rewriter: rewriter,
	}
}

func (uc *NewsUseCase) Sync(ctx context.Context, limit int, mood domain.Mood) error {
	items, err := uc.source.FetchLatest(ctx, limit)
	if err != nil {
		return fmt.Errorf("fetch latest news: %w", err)
	}

	result := make([]domain.News, 0, len(items))
	now := time.Now().UTC()

	for _, item := range items {
		rewritten, err := uc.rewriter.Rewrite(ctx, domain.RewriteRequest{
			Title: item.Title,
			Text:  item.Text,
			Mood:  mood,
		})
		if err != nil {
			return fmt.Errorf("rewrite news %s: %w", item.ExternalID, err)
		}

		publishedAt, _ := time.Parse(time.RFC3339, item.PublishedRaw)
		originalDigest := checksum(item.Text)

		result = append(result, domain.News{
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
		})
	}

	if err := uc.repo.SaveBatch(ctx, result); err != nil {
		return fmt.Errorf("save batch: %w", err)
	}

	return nil
}

func (uc *NewsUseCase) List(ctx context.Context, mood domain.Mood) ([]domain.News, error) {
	return uc.repo.List(ctx, mood)
}

func (uc *NewsUseCase) Get(ctx context.Context, id string) (domain.News, error) {
	return uc.repo.GetByID(ctx, id)
}

func checksum(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
