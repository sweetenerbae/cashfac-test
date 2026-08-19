package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"cashfac-test/internal/domain"
	"cashfac-test/internal/platform/logger"
)

type SyncJobsUseCase struct {
	jobs        domain.SyncJobRepository
	newsUse     *NewsUseCase
	mu          sync.Mutex
	activeJobID string
}

func NewSyncJobsUseCase(jobs domain.SyncJobRepository, newsUse *NewsUseCase) *SyncJobsUseCase {
	return &SyncJobsUseCase{
		jobs:    jobs,
		newsUse: newsUse,
	}
}

func (uc *SyncJobsUseCase) Start(ctx context.Context, limit int) (domain.SyncJob, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if uc.activeJobID != "" {
		activeJob, err := uc.jobs.GetByID(ctx, uc.activeJobID)
		if err == nil && (activeJob.Status == domain.JobStatusPending || activeJob.Status == domain.JobStatusRunning) {
			logger.Info("SYNC", "active news sync reused", logger.F("job_id", activeJob.ID))
			return activeJob, nil
		}
		uc.activeJobID = ""
	}

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
	uc.activeJobID = job.ID

	logger.Info("SYNC", "news sync queued",
		logger.F("job_id", job.ID),
		logger.F("limit", limit),
	)
	go uc.run(job)

	return job, nil
}

func (uc *SyncJobsUseCase) Get(ctx context.Context, id string) (domain.SyncJob, error) {
	return uc.jobs.GetByID(ctx, id)
}

func (uc *SyncJobsUseCase) run(job domain.SyncJob) {
	defer func() {
		uc.mu.Lock()
		if uc.activeJobID == job.ID {
			uc.activeJobID = ""
		}
		uc.mu.Unlock()
	}()

	ctx := context.Background()
	startedAt := time.Now()
	job.Status = domain.JobStatusRunning
	job.UpdatedAt = time.Now().UTC()
	if err := uc.jobs.Save(ctx, job); err != nil {
		logger.Error("SYNC", "failed to mark job as running", logger.F("job_id", job.ID), logger.F("error", err))
		return
	}
	logger.Info("SYNC", "news sync started", logger.F("job_id", job.ID), logger.F("limit", job.Limit))

	count, err := uc.newsUse.SyncWithProgress(ctx, job.Limit, func(processed, total int) {
		job.ProcessedCount = processed
		job.TotalCount = total
		job.UpdatedAt = time.Now().UTC()
		if saveErr := uc.jobs.Save(ctx, job); saveErr != nil {
			logger.Warn("SYNC", "failed to save job progress", logger.F("job_id", job.ID), logger.F("error", saveErr))
		}
	})
	if err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = err.Error()
		job.ProcessedCount = count
		if job.TotalCount == 0 {
			job.TotalCount = job.Limit
		}
		job.UpdatedAt = time.Now().UTC()
		if saveErr := uc.jobs.Save(ctx, job); saveErr != nil {
			logger.Error("SYNC", "failed to save job failure", logger.F("job_id", job.ID), logger.F("error", saveErr))
		}
		logger.Error("SYNC", "news sync failed",
			logger.F("job_id", job.ID),
			logger.F("processed", count),
			logger.F("duration", logger.Duration(time.Since(startedAt))),
			logger.F("error", err),
		)
		return
	}

	job.Status = domain.JobStatusCompleted
	job.ProcessedCount = count
	if job.TotalCount == 0 {
		job.TotalCount = count
	}
	job.UpdatedAt = time.Now().UTC()
	if err := uc.jobs.Save(ctx, job); err != nil {
		logger.Error("SYNC", "failed to save completed job", logger.F("job_id", job.ID), logger.F("error", err))
		return
	}
	logger.Success("SYNC", "news sync completed",
		logger.F("job_id", job.ID),
		logger.F("saved", count),
		logger.F("duration", logger.Duration(time.Since(startedAt))),
	)
}
