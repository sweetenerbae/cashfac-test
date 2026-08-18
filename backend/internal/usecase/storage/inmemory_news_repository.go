package storage

import (
	"context"
	"slices"
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

func (r *InMemoryNewsRepository) Save(_ context.Context, item domain.News) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items[item.ID] = item

	return nil
}

func (r *InMemoryNewsRepository) SaveBatch(_ context.Context, items []domain.News) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, item := range items {
		r.items[item.ID] = item
	}

	return nil
}

func (r *InMemoryNewsRepository) PruneByExternalIDs(_ context.Context, externalIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	allowed := make(map[string]struct{}, len(externalIDs))
	for _, externalID := range externalIDs {
		allowed[externalID] = struct{}{}
	}

	for id, item := range r.items {
		if _, ok := allowed[item.ExternalID]; ok {
			continue
		}
		delete(r.items, id)
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

	slices.SortFunc(result, func(a, b domain.News) int {
		if a.PublishedAt.Equal(b.PublishedAt) {
			return b.CreatedAt.Compare(a.CreatedAt)
		}
		return b.PublishedAt.Compare(a.PublishedAt)
	})

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

func (r *InMemoryNewsRepository) GetByExternalID(_ context.Context, externalID string) (domain.News, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var (
		found bool
		match domain.News
	)

	for _, item := range r.items {
		if item.ExternalID != externalID {
			continue
		}
		if !found || item.CreatedAt.After(match.CreatedAt) {
			match = item
			found = true
		}
	}

	if !found {
		return domain.News{}, domain.ErrNewsNotFound
	}

	return match, nil
}

func (r *InMemoryNewsRepository) GetByExternalIDAndMood(_ context.Context, externalID string, mood domain.Mood) (domain.News, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var (
		found bool
		match domain.News
	)

	for _, item := range r.items {
		if item.ExternalID != externalID || item.Mood != mood {
			continue
		}
		if !found || item.CreatedAt.After(match.CreatedAt) {
			match = item
			found = true
		}
	}

	if !found {
		return domain.News{}, domain.ErrNewsNotFound
	}

	return match, nil
}
