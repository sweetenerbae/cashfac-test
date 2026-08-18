package storage

import (
	"context"
	"sync"

	"cashfac-test/internal/domain"
)

type InMemoryNewsRepository struct {
	mu    sync.RWMutex
	items map[string]domain.News
}

func NewInMemoryNewsRepository() *InMemoryNewsRepository {
	return &InMemoryNewsRepository{
		items: make(map[string]domain.News),
	}
}

func (r *InMemoryNewsRepository) SaveBatch(_ context.Context, items []domain.News) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, item := range items {
		r.items[item.ID] = item
	}

	return nil
}

func (r *InMemoryNewsRepository) List(_ context.Context, mood domain.Mood) ([]domain.News, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.News, 0, len(r.items))
	for _, item := range r.items {
		if mood != "" && item.Mood != mood {
			continue
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *InMemoryNewsRepository) GetByID(_ context.Context, id string) (domain.News, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.items[id]
	if !ok {
		return domain.News{}, domain.ErrNewsNotFound
	}

	return item, nil
}
