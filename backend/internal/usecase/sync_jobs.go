package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"cashfac-test/internal/domain"
)

type SyncJobsUseCase struct {
	jobs    domain.SyncJobRepository
	newsUse *NewsUseCase
}

func NewSyncJobsUseCase(jobs domain.SyncJobRepository, newsUse *NewsUseCase) *SyncJobsUseCase {
	return &SyncJobsUseCase{
		jobs:    jobs,
		newsUse: newsUse,
	}
}

func (uc *SyncJobsUseCase) Start(ctx context.Context, limit int) (domain.SyncJob, error) {
	now := time.Now().UTC()
	job := domain.SyncJob{
		ID:        uuid.NewString(),
		Status:    domain.JobStatusPending,
		Mood:      domain.MoodNeutral,
		Limit:     limit,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.jobs.Save(ctx, job); err != nil {
		return domain.SyncJob{}, fmt.Errorf("save sync job: %w", err)
	}

	go uc.run(job)

	return job, nil
}

func (uc *SyncJobsUseCase) Get(ctx context.Context, id string) (domain.SyncJob, error) {
	return uc.jobs.GetByID(ctx, id)
}

func (uc *SyncJobsUseCase) run(job domain.SyncJob) {
	ctx := context.Background()
	job.Status = domain.JobStatusRunning
	job.UpdatedAt = time.Now().UTC()
	_ = uc.jobs.Save(ctx, job)

	count, err := uc.newsUse.SyncWithProgress(ctx, job.Limit, func(processed, total int) {
		job.ProcessedCount = processed
		job.TotalCount = total
		job.UpdatedAt = time.Now().UTC()
		_ = uc.jobs.Save(ctx, job)
	})
	if err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = err.Error()
		job.ProcessedCount = count
		if job.TotalCount == 0 {
			job.TotalCount = job.Limit
		}
		job.UpdatedAt = time.Now().UTC()
		_ = uc.jobs.Save(ctx, job)
		return
	}

	job.Status = domain.JobStatusCompleted
	job.ProcessedCount = count
	if job.TotalCount == 0 {
		job.TotalCount = count
	}
	job.UpdatedAt = time.Now().UTC()
	_ = uc.jobs.Save(ctx, job)
}
