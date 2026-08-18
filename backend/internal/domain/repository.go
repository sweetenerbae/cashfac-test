package domain

import "context"

type NewsRepository interface {
	SaveBatch(ctx context.Context, items []News) error
	List(ctx context.Context, mood Mood) ([]News, error)
	GetByID(ctx context.Context, id string) (News, error)
}
