package usecase

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cashfac-test/internal/domain"
	"cashfac-test/internal/usecase/storage"
)

func TestSyncJobsUseCaseReusesActiveJob(t *testing.T) {
	t.Parallel()

	source := &blockingSyncSource{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	newsUse := NewNewsUseCase(storage.NewInMemoryNewsRepository(), source, &fakeRewriter{})
	jobs := storage.NewInMemorySyncJobRepository()
	uc := NewSyncJobsUseCase(jobs, newsUse)

	first, err := uc.Start(context.Background(), 10)
	if err != nil {
		t.Fatalf("start first sync: %v", err)
	}
	second, err := uc.Start(context.Background(), 10)
	if err != nil {
		t.Fatalf("start second sync: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected active job %q to be reused, got %q", first.ID, second.ID)
	}

	select {
	case <-source.started:
	case <-time.After(time.Second):
		t.Fatal("sync did not start")
	}
	close(source.release)

	deadline := time.Now().Add(time.Second)
	for {
		job, getErr := uc.Get(context.Background(), first.ID)
		if getErr != nil {
			t.Fatalf("get sync job: %v", getErr)
		}
		if job.Status == domain.JobStatusCompleted {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("sync did not complete, last status %q", job.Status)
		}
		time.Sleep(5 * time.Millisecond)
	}

	if calls := source.calls.Load(); calls != 1 {
		t.Fatalf("expected one source request, got %d", calls)
	}
}

type blockingSyncSource struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
	calls   atomic.Int32
}

func (s *blockingSyncSource) FetchLatest(_ context.Context, _ int) ([]domain.SourceItem, error) {
	s.calls.Add(1)
	s.once.Do(func() {
		close(s.started)
	})
	<-s.release
	return []domain.SourceItem{makeSourceItem("news-1", "First news")}, nil
}
