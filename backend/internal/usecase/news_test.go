package usecase

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cashfac-test/internal/domain"
	"cashfac-test/internal/usecase/storage"
)

func TestNewsUseCaseSyncStoresOriginalsWithoutCallingRewriter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := storage.NewInMemoryNewsRepository()
	rewriter := &fakeRewriter{responseText: "this should not be used"}
	source := &fakeSource{
		batches: [][]domain.SourceItem{{makeSourceItem("news-1", "First news")}},
	}
	uc := NewNewsUseCase(repo, source, rewriter)

	count, err := uc.Sync(ctx, 1)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one saved item, got %d", count)
	}
	if rewriter.calls != 0 {
		t.Fatalf("expected sync not to call rewriter, got %d calls", rewriter.calls)
	}

	items, err := uc.List(ctx, domain.MoodNeutral)
	if err != nil {
		t.Fatalf("list neutral items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one neutral item, got %d", len(items))
	}
	if items[0].RewrittenText != "" {
		t.Fatalf("expected lazy rewrite to be empty, got %q", items[0].RewrittenText)
	}
	if items[0].ImageURL != source.batches[0][0].ImageURL {
		t.Fatalf("expected image URL %q, got %q", source.batches[0][0].ImageURL, items[0].ImageURL)
	}
}

func TestNewsUseCaseSyncKeepsCachedRewriteForUnchangedSource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := storage.NewInMemoryNewsRepository()
	rewriter := &fakeRewriter{responseText: "this should not be used"}
	sourceItem := makeSourceItem("news-1", "First news")
	cachedItem := makeNewsItem("news-1", domain.MoodNeutral, sourceItem.Text, "A cached neutral rewrite.")
	if err := repo.Save(ctx, cachedItem); err != nil {
		t.Fatalf("save cached item: %v", err)
	}

	uc := NewNewsUseCase(repo, &fakeSource{batches: [][]domain.SourceItem{{sourceItem}}}, rewriter)
	if _, err := uc.Sync(ctx, 1); err != nil {
		t.Fatalf("sync: %v", err)
	}

	item, err := uc.GetByExternalID(ctx, "news-1", domain.MoodNeutral)
	if err != nil {
		t.Fatalf("get neutral item: %v", err)
	}
	if item.RewrittenText != cachedItem.RewrittenText {
		t.Fatalf("expected cached rewrite %q, got %q", cachedItem.RewrittenText, item.RewrittenText)
	}
	if rewriter.calls != 0 {
		t.Fatalf("expected no rewriter calls, got %d", rewriter.calls)
	}
}

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

	if _, err := uc.Sync(ctx, 2); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	if _, err := uc.RewriteByExternalID(ctx, "news-1", domain.MoodHappy); err != nil {
		t.Fatalf("rewrite by external id: %v", err)
	}

	if _, err := uc.Sync(ctx, 1); err != nil {
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

	result, err := uc.RewriteByExternalID(ctx, "news-1", domain.MoodSad)
	if err != nil {
		t.Fatalf("rewrite by external id: %v", err)
	}

	if rewriter.calls != 1 {
		t.Fatalf("expected rewriter to be called once, got %d", rewriter.calls)
	}
	if result.News.RewrittenText != rewriter.responseText {
		t.Fatalf("expected rewritten text %q, got %q", rewriter.responseText, result.News.RewrittenText)
	}
	if result.Meta.Source != domain.RewriteSourceGenerated || result.Meta.LLMRequests != 1 {
		t.Fatalf("expected generated result with one LLM request, got %#v", result.Meta)
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

	result, err := uc.RewriteByExternalID(ctx, "news-1", domain.MoodIronic)
	if err != nil {
		t.Fatalf("rewrite by external id: %v", err)
	}

	if rewriter.calls != 0 {
		t.Fatalf("expected cached rewrite to be reused, got %d rewriter calls", rewriter.calls)
	}
	if result.News.RewrittenText != cachedItem.RewrittenText {
		t.Fatalf("expected cached rewritten text %q, got %q", cachedItem.RewrittenText, result.News.RewrittenText)
	}
	if result.Meta.Source != domain.RewriteSourceCache || result.Meta.LLMRequests != 0 || result.Meta.SavedLLMRequests != 1 {
		t.Fatalf("expected cache metrics, got %#v", result.Meta)
	}
}

func TestNewsUseCaseRewriteByExternalIDDeduplicatesConcurrentRequests(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := storage.NewInMemoryNewsRepository()
	rewriter := newBlockingRewriter()
	uc := NewNewsUseCase(repo, &fakeSource{}, rewriter)

	baseItem := makeNewsItem("news-1", domain.MoodNeutral, "Original story text.", "")
	if err := repo.Save(ctx, baseItem); err != nil {
		t.Fatalf("save base item: %v", err)
	}

	type concurrentResult struct {
		result domain.RewriteResult
		err    error
	}
	results := make(chan concurrentResult, 2)
	go func() {
		result, err := uc.RewriteByExternalID(ctx, "news-1", domain.MoodHappy)
		results <- concurrentResult{result: result, err: err}
	}()

	<-rewriter.started
	go func() {
		result, err := uc.RewriteByExternalID(ctx, "news-1", domain.MoodHappy)
		results <- concurrentResult{result: result, err: err}
	}()

	time.Sleep(10 * time.Millisecond)
	close(rewriter.release)

	llmRequests := 0
	savedRequests := 0
	sources := map[domain.RewriteResultSource]int{}
	for range 2 {
		outcome := <-results
		if outcome.err != nil {
			t.Fatalf("rewrite by external id: %v", outcome.err)
		}
		llmRequests += outcome.result.Meta.LLMRequests
		savedRequests += outcome.result.Meta.SavedLLMRequests
		sources[outcome.result.Meta.Source]++
	}
	if calls := rewriter.calls.Load(); calls != 1 {
		t.Fatalf("expected one shared rewriter call, got %d", calls)
	}
	if llmRequests != 1 || savedRequests != 1 {
		t.Fatalf("expected one LLM request and one saved request, got llm=%d saved=%d", llmRequests, savedRequests)
	}
	if sources[domain.RewriteSourceGenerated] != 1 || sources[domain.RewriteSourceShared] != 1 {
		t.Fatalf("expected generated and shared results, got %#v", sources)
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

type blockingRewriter struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
	calls   atomic.Int32
}

func newBlockingRewriter() *blockingRewriter {
	return &blockingRewriter{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (r *blockingRewriter) Rewrite(_ context.Context, request domain.RewriteRequest) (domain.RewriteResponse, error) {
	r.calls.Add(1)
	r.once.Do(func() {
		close(r.started)
	})
	<-r.release

	return domain.RewriteResponse{
		Text:         "A shared cached rewrite for concurrent callers.",
		FactChecksum: request.Title + ":" + request.Text,
	}, nil
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
		ImageURL:     "https://media.example.com/" + externalID + ".jpg",
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
		ImageURL:       "https://media.example.com/" + externalID + ".jpg",
		PublishedAt:    publishedAt,
		CreatedAt:      publishedAt,
		ExternalID:     externalID,
		FactChecksum:   "checksum",
		OriginalDigest: checksum(originalText),
	}
}
