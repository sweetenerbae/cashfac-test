package domain

import "context"

type SyncJobRepository interface {
	Save(ctx context.Context, job SyncJob) error
	GetByID(ctx context.Context, id string) (SyncJob, error)
}
