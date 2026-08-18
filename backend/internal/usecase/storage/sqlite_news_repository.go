package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"cashfac-test/internal/domain"
)

type SQLiteNewsRepository struct {
	db *sql.DB
}

func NewSQLiteNewsRepository(databasePath string) (*SQLiteNewsRepository, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, fmt.Errorf("create sqlite directory: %w", err)
	}

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	repo := &SQLiteNewsRepository{db: db}
	if err := repo.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return repo, nil
}

func (r *SQLiteNewsRepository) Save(ctx context.Context, item domain.News) error {
	const query = `
		INSERT INTO news (
			id, title, original_text, rewritten_text, mood, source_name, source_url,
			published_at, created_at, external_id, fact_checksum, original_digest
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			original_text = excluded.original_text,
			rewritten_text = excluded.rewritten_text,
			mood = excluded.mood,
			source_name = excluded.source_name,
			source_url = excluded.source_url,
			published_at = excluded.published_at,
			created_at = excluded.created_at,
			external_id = excluded.external_id,
			fact_checksum = excluded.fact_checksum,
			original_digest = excluded.original_digest
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID,
		item.Title,
		item.OriginalText,
		item.RewrittenText,
		string(item.Mood),
		item.SourceName,
		item.SourceURL,
		item.PublishedAt.UTC().Format(sqliteTimeLayout),
		item.CreatedAt.UTC().Format(sqliteTimeLayout),
		item.ExternalID,
		item.FactChecksum,
		item.OriginalDigest,
	)
	if err != nil {
		return fmt.Errorf("save news in sqlite: %w", err)
	}

	return nil
}

func (r *SQLiteNewsRepository) SaveBatch(ctx context.Context, items []domain.News) error {
	for _, item := range items {
		if err := r.Save(ctx, item); err != nil {
			return err
		}
	}

	return nil
}

func (r *SQLiteNewsRepository) List(ctx context.Context, mood domain.Mood) ([]domain.News, error) {
	query := `
		SELECT
			id, title, original_text, rewritten_text, mood, source_name, source_url,
			published_at, created_at, external_id, fact_checksum, original_digest
		FROM news
	`
	args := []any{}

	if mood != "" {
		query += ` WHERE mood = ?`
		args = append(args, string(mood))
	}

	query += ` ORDER BY published_at DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list news from sqlite: %w", err)
	}
	defer rows.Close()

	var result []domain.News
	for rows.Next() {
		item, err := scanNews(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate news rows: %w", err)
	}

	return result, nil
}

func (r *SQLiteNewsRepository) GetByID(ctx context.Context, id string) (domain.News, error) {
	const query = `
		SELECT
			id, title, original_text, rewritten_text, mood, source_name, source_url,
			published_at, created_at, external_id, fact_checksum, original_digest
		FROM news
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanNews(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.News{}, domain.ErrNewsNotFound
		}
		return domain.News{}, err
	}

	return item, nil
}

func (r *SQLiteNewsRepository) GetByExternalID(ctx context.Context, externalID string) (domain.News, error) {
	const query = `
		SELECT
			id, title, original_text, rewritten_text, mood, source_name, source_url,
			published_at, created_at, external_id, fact_checksum, original_digest
		FROM news
		WHERE external_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, externalID)
	item, err := scanNews(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.News{}, domain.ErrNewsNotFound
		}
		return domain.News{}, err
	}

	return item, nil
}

func (r *SQLiteNewsRepository) GetByExternalIDAndMood(ctx context.Context, externalID string, mood domain.Mood) (domain.News, error) {
	const query = `
		SELECT
			id, title, original_text, rewritten_text, mood, source_name, source_url,
			published_at, created_at, external_id, fact_checksum, original_digest
		FROM news
		WHERE external_id = ? AND mood = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, externalID, string(mood))
	item, err := scanNews(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.News{}, domain.ErrNewsNotFound
		}
		return domain.News{}, err
	}

	return item, nil
}

func (r *SQLiteNewsRepository) Close() error {
	return r.db.Close()
}

func (r *SQLiteNewsRepository) migrate() error {
	const query = `
		CREATE TABLE IF NOT EXISTS news (
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
		);

		CREATE INDEX IF NOT EXISTS idx_news_mood_published_at
		ON news (mood, published_at DESC);

		CREATE INDEX IF NOT EXISTS idx_news_external_id_mood_created_at
		ON news (external_id, mood, created_at DESC);
	`

	if _, err := r.db.Exec(query); err != nil {
		return fmt.Errorf("migrate sqlite schema: %w", err)
	}

	return nil
}

const sqliteTimeLayout = "2006-01-02T15:04:05.999999999Z07:00"

type scanner interface {
	Scan(dest ...any) error
}

func scanNews(row scanner) (domain.News, error) {
	var (
		item        domain.News
		mood        string
		publishedAt string
		createdAt   string
	)

	err := row.Scan(
		&item.ID,
		&item.Title,
		&item.OriginalText,
		&item.RewrittenText,
		&mood,
		&item.SourceName,
		&item.SourceURL,
		&publishedAt,
		&createdAt,
		&item.ExternalID,
		&item.FactChecksum,
		&item.OriginalDigest,
	)
	if err != nil {
		return domain.News{}, err
	}

	item.Mood = domain.Mood(mood)

	if item.PublishedAt, err = parseSQLiteTime(publishedAt); err != nil {
		return domain.News{}, fmt.Errorf("parse published_at: %w", err)
	}

	if item.CreatedAt, err = parseSQLiteTime(createdAt); err != nil {
		return domain.News{}, fmt.Errorf("parse created_at: %w", err)
	}

	return item, nil
}

func parseSQLiteTime(value string) (time.Time, error) {
	return time.Parse(sqliteTimeLayout, value)
}
