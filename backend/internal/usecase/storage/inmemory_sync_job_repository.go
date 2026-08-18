package storage

import (
	"context"
	"sync"

	"cashfac-test/internal/domain"
)

type InMemorySyncJobRepository struct {
	mu    sync.RWMutex
	items map[string]domain.SyncJob
}

func NewInMemorySyncJobRepository() *InMemorySyncJobRepository {
	return &InMemorySyncJobRepository{
		items: make(map[string]domain.SyncJob),
	}
}

func (r *InMemorySyncJobRepository) Save(_ context.Context, job domain.SyncJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items[job.ID] = job
	return nil
}

func (r *InMemorySyncJobRepository) GetByID(_ context.Context, id string) (domain.SyncJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	job, ok := r.items[id]
	if !ok {
		return domain.SyncJob{}, domain.ErrJobNotFound
	}

	return job, nil
}
