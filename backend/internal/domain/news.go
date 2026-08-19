package domain

import "time"

type Mood string

const (
	MoodNeutral Mood = "neutral"
	MoodHappy   Mood = "happy"
	MoodSad     Mood = "sad"
	MoodIronic  Mood = "ironic"
)

func (m Mood) IsValid() bool {
	switch m {
	case MoodNeutral, MoodHappy, MoodSad, MoodIronic:
		return true
	default:
		return false
	}
}

type News struct {
	ID             string
	Title          string
	OriginalText   string
	RewrittenText  string
	Mood           Mood
	SourceName     string
	SourceURL      string
	ImageURL       string
	PublishedAt    time.Time
	CreatedAt      time.Time
	ExternalID     string
	FactChecksum   string
	OriginalDigest string
}
