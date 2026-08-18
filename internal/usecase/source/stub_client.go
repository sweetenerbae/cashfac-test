package source

import (
	"context"
	"time"

	"cashfac-test/internal/domain"
)

type StubClient struct{}

func NewStubClient() *StubClient {
	return &StubClient{}
}

func (c *StubClient) FetchLatest(_ context.Context, limit int) ([]domain.SourceItem, error) {
	if limit <= 0 {
		limit = 10
	}

	now := time.Now().UTC()
	items := make([]domain.SourceItem, 0, limit)
	for i := 0; i < limit; i++ {
		items = append(items, domain.SourceItem{
			ExternalID:   now.Format("20060102") + "-" + string(rune('a'+i)),
			Title:        "Stub news title",
			Text:         "Stub news text. Replace this source client with RSS/API integration.",
			SourceName:   "stub-source",
			SourceURL:    "https://example.com/news",
			PublishedRaw: now.Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
		})
	}

	return items, nil
}
