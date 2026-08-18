package usecase

import (
	"context"
	"testing"
	"time"

	"cashfac-test/internal/domain"
	"cashfac-test/internal/usecase/storage"
)

func TestNewsUseCaseSyncPrunesStaleNews(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := storage.NewInMemoryNewsRepository()
	rewriter := &fakeRewriter{
		responseText: "rewritten text",
	}
	source := &fakeSource{
		batches: [][]domain.SourceItem{
			{
				makeSourceItem("news-1", "First news"),
				makeSourceItem("news-2", "Second news"),
			},
			{
				makeSourceItem("news-2", "Second news"),
			},
		},
	}

	uc := NewNewsUseCase(repo, source, rewriter)

	if _, err := uc.Sync(ctx, 2, domain.MoodNeutral); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	if _, err := uc.RewriteByExternalID(ctx, "news-1", domain.MoodHappy); err != nil {
		t.Fatalf("rewrite by external id: %v", err)
	}

	if _, err := uc.Sync(ctx, 1, domain.MoodNeutral); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	neutralItems, err := uc.List(ctx, domain.MoodNeutral)
	if err != nil {
		t.Fatalf("list neutral items: %v", err)
	}
	if len(neutralItems) != 1 || neutralItems[0].ExternalID != "news-2" {
		t.Fatalf("expected only news-2 in neutral list, got %#v", neutralItems)
	}

	happyItems, err := uc.List(ctx, domain.MoodHappy)
	if err != nil {
		t.Fatalf("list happy items: %v", err)
	}
	if len(happyItems) != 0 {
		t.Fatalf("expected stale mood variants to be pruned, got %#v", happyItems)
	}
}

func TestNewsUseCaseRewriteByExternalIDRewritesWeakCachedVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := storage.NewInMemoryNewsRepository()
	rewriter := &fakeRewriter{
		responseText: "A much sadder rewrite with a different voice.",
	}
	uc := NewNewsUseCase(repo, &fakeSource{}, rewriter)

	baseItem := makeNewsItem("news-1", domain.MoodNeutral, "Original story text.", "Original story text.")
	weakCachedItem := makeNewsItem("news-1", domain.MoodSad, "Original story text.", "Original story text.")

	if err := repo.Save(ctx, baseItem); err != nil {
		t.Fatalf("save base item: %v", err)
	}
	if err := repo.Save(ctx, weakCachedItem); err != nil {
		t.Fatalf("save weak cached item: %v", err)
	}

	item, err := uc.RewriteByExternalID(ctx, "news-1", domain.MoodSad)
	if err != nil {
		t.Fatalf("rewrite by external id: %v", err)
	}

	if rewriter.calls != 1 {
		t.Fatalf("expected rewriter to be called once, got %d", rewriter.calls)
	}
	if item.RewrittenText != rewriter.responseText {
		t.Fatalf("expected rewritten text %q, got %q", rewriter.responseText, item.RewrittenText)
	}
}

func TestNewsUseCaseRewriteByExternalIDReusesStrongCachedVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := storage.NewInMemoryNewsRepository()
	rewriter := &fakeRewriter{
		responseText: "this should not be used",
	}
	uc := NewNewsUseCase(repo, &fakeSource{}, rewriter)

	baseItem := makeNewsItem("news-1", domain.MoodNeutral, "Original story text.", "Original story text.")
	cachedItem := makeNewsItem("news-1", domain.MoodIronic, "Original story text.", "A dry and noticeably different rewrite.")

	if err := repo.Save(ctx, baseItem); err != nil {
		t.Fatalf("save base item: %v", err)
	}
	if err := repo.Save(ctx, cachedItem); err != nil {
		t.Fatalf("save cached item: %v", err)
	}

	item, err := uc.RewriteByExternalID(ctx, "news-1", domain.MoodIronic)
	if err != nil {
		t.Fatalf("rewrite by external id: %v", err)
	}

	if rewriter.calls != 0 {
		t.Fatalf("expected cached rewrite to be reused, got %d rewriter calls", rewriter.calls)
	}
	if item.RewrittenText != cachedItem.RewrittenText {
		t.Fatalf("expected cached rewritten text %q, got %q", cachedItem.RewrittenText, item.RewrittenText)
	}
}

type fakeSource struct {
	batches [][]domain.SourceItem
	calls   int
}

func (s *fakeSource) FetchLatest(_ context.Context, _ int) ([]domain.SourceItem, error) {
	if s.calls >= len(s.batches) {
		return nil, nil
	}

	items := s.batches[s.calls]
	s.calls++
	return items, nil
}

type fakeRewriter struct {
	responseText string
	calls        int
}

func (r *fakeRewriter) Rewrite(_ context.Context, request domain.RewriteRequest) (domain.RewriteResponse, error) {
	r.calls++

	text := r.responseText
	if text == "" {
		text = string(request.Mood) + ": " + request.Text
	}

	return domain.RewriteResponse{
		Text:         text,
		FactChecksum: request.Title + ":" + request.Text,
	}, nil
}

func makeSourceItem(externalID, title string) domain.SourceItem {
	return domain.SourceItem{
		ExternalID:   externalID,
		Title:        title,
		Text:         title + " body",
		SourceName:   "The Guardian",
		SourceURL:    "https://example.com/" + externalID,
		PublishedRaw: "2026-08-18T10:00:00Z",
	}
}

func makeNewsItem(externalID string, mood domain.Mood, originalText, rewrittenText string) domain.News {
	publishedAt := time.Date(2026, time.August, 18, 10, 0, 0, 0, time.UTC)

	return domain.News{
		ID:             checksum(externalID + string(mood)),
		Title:          "Sample title",
		OriginalText:   originalText,
		RewrittenText:  rewrittenText,
		Mood:           mood,
		SourceName:     "The Guardian",
		SourceURL:      "https://example.com/" + externalID,
		PublishedAt:    publishedAt,
		CreatedAt:      publishedAt,
		ExternalID:     externalID,
		FactChecksum:   "checksum",
		OriginalDigest: checksum(originalText),
	}
}
