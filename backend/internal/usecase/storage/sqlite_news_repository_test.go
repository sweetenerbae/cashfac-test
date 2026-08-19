package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"cashfac-test/internal/domain"
)

func TestSQLiteNewsRepositoryAddsImageURLToExistingSchema(t *testing.T) {
	t.Parallel()

	databasePath := filepath.Join(t.TempDir(), "news.db")
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}

	const legacySchema = `
		CREATE TABLE news (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			original_text TEXT NOT NULL,
			rewritten_text TEXT NOT NULL,
			mood TEXT NOT NULL,
			source_name TEXT NOT NULL,
			source_url TEXT NOT NULL,
			published_at TEXT NOT NULL,
			created_at TEXT NOT NULL,
			external_id TEXT NOT NULL,
			fact_checksum TEXT NOT NULL,
			original_digest TEXT NOT NULL
		)
	`
	if _, err := db.Exec(legacySchema); err != nil {
		_ = db.Close()
		t.Fatalf("create legacy schema: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close legacy database: %v", err)
	}

	repo, err := NewSQLiteNewsRepository(databasePath)
	if err != nil {
		t.Fatalf("open migrated repository: %v", err)
	}
	defer repo.Close()

	now := time.Now().UTC()
	item := domain.News{
		ID:             "news-1-neutral",
		Title:          "News title",
		OriginalText:   "News body",
		Mood:           domain.MoodNeutral,
		SourceName:     "The Guardian",
		SourceURL:      "https://example.com/news-1",
		ImageURL:       "https://media.example.com/news-1.jpg",
		PublishedAt:    now,
		CreatedAt:      now,
		ExternalID:     "news-1",
		OriginalDigest: "digest",
	}
	if err := repo.Save(context.Background(), item); err != nil {
		t.Fatalf("save news with image: %v", err)
	}

	saved, err := repo.GetByID(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("get saved news: %v", err)
	}
	if saved.ImageURL != item.ImageURL {
		t.Fatalf("expected image URL %q, got %q", item.ImageURL, saved.ImageURL)
	}
}
