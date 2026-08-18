package domain

import "time"

type Mood string

const (
	MoodNeutral Mood = "neutral"
	MoodHappy   Mood = "happy"
	MoodSad     Mood = "sad"
	MoodIronic  Mood = "ironic"
)

type News struct {
	ID             string
	Title          string
	OriginalText   string
	RewrittenText  string
	Mood           Mood
	SourceName     string
	SourceURL      string
	PublishedAt    time.Time
	CreatedAt      time.Time
	ExternalID     string
	FactChecksum   string
	OriginalDigest string
}
