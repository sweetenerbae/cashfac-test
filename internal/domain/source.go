package domain

import "context"

type SourceItem struct {
	ExternalID   string
	Title        string
	Text         string
	SourceName   string
	SourceURL    string
	PublishedRaw string
}

type SourceClient interface {
	FetchLatest(ctx context.Context, limit int) ([]SourceItem, error)
}
