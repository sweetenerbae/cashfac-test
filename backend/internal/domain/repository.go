package domain

import "context"

type NewsRepository interface {
	Save(ctx context.Context, item News) error
	SaveBatch(ctx context.Context, items []News) error
	PruneByExternalIDs(ctx context.Context, externalIDs []string) error
	List(ctx context.Context, mood Mood) ([]News, error)
	GetByID(ctx context.Context, id string) (News, error)
	GetByExternalID(ctx context.Context, externalID string) (News, error)
	GetByExternalIDAndMood(ctx context.Context, externalID string, mood Mood) (News, error)
}
