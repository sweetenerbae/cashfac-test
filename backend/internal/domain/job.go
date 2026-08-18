package domain

import "time"

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type SyncJob struct {
	ID             string
	Status         JobStatus
	Mood           Mood
	Limit          int
	ProcessedCount int
	TotalCount     int
	Error          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
