package content

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUnavailable = errors.New("content store unavailable: pool nil")

var ErrNotFound = errors.New("content block not found")

type Block struct {
	Slug          string          `json:"slug"`
	Locale        string          `json:"locale"`
	Content       json.RawMessage `json:"content"`
	DraftContent  json.RawMessage `json:"draft_content,omitempty"`
	SchemaVersion int             `json:"schema_version"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
	UpdatedAt     time.Time       `json:"updated_at"`
	UpdatedBy     *uuid.UUID      `json:"updated_by,omitempty"`
}

type MatrixEntry struct {
	Slug         string     `json:"slug"`
	Locale       string     `json:"locale"`
	HasPublished bool       `json:"has_published"`
	HasDraft     bool       `json:"has_draft"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
	UpdatedBy    *uuid.UUID `json:"updated_by,omitempty"`
}

type UpsertParams struct {
	Slug           string
	Locale         string
	Content        json.RawMessage
	Publish        bool
	SchemaVersion  int
	UpdatedBy      uuid.UUID
	PriorUpdatedAt *time.Time
}

type ErrPrecondition struct {
	CurrentUpdatedAt time.Time
}

func (e *ErrPrecondition) Error() string {
	return fmt.Sprintf("if-match precondition failed: current updated_at=%s",
		e.CurrentUpdatedAt.UTC().Format(time.RFC3339Nano))
}

type Store interface {
	Lookup(ctx context.Context, slug, locale string, includeDraft bool) (*Block, error)
	Upsert(ctx context.Context, p UpsertParams) (*Block, error)
	Delete(ctx context.Context, slug, locale string) (*Block, error)
	ListMatrix(ctx context.Context) ([]MatrixEntry, error)
}

type PGStore struct {
	Pool *pgxpool.Pool
}

func NewPGStore(pool *pgxpool.Pool) *PGStore {
	return &PGStore{Pool: pool}
}

func (s *PGStore) Lookup(ctx context.Context, slug, locale string, includeDraft bool) (*Block, error) {
	if s == nil || s.Pool == nil {
		return nil, ErrUnavailable
	}
	row := s.Pool.QueryRow(ctx, `
		SELECT slug, locale, content, draft_content, schema_version,
		       published_at, updated_at, updated_by
		FROM content_blocks
		WHERE slug = $1 AND locale = $2`, slug, locale)
	var b Block
	var rawContent, rawDraft []byte
	if err := row.Scan(&b.Slug, &b.Locale, &rawContent, &rawDraft,
		&b.SchemaVersion, &b.PublishedAt, &b.UpdatedAt, &b.UpdatedBy); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if len(rawContent) > 0 {
		b.Content = json.RawMessage(rawContent)
	}
	if includeDraft && len(rawDraft) > 0 {
		b.DraftContent = json.RawMessage(rawDraft)
	}
	return &b, nil
}

func (s *PGStore) currentUpdatedAt(ctx context.Context, slug, locale string) (time.Time, error) {
	var t time.Time
	err := s.Pool.QueryRow(ctx,
		`SELECT updated_at FROM content_blocks WHERE slug=$1 AND locale=$2`,
		slug, locale).Scan(&t)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrNotFound
	}
	return t, err
}

func (s *PGStore) Upsert(ctx context.Context, p UpsertParams) (*Block, error) {
	if s == nil || s.Pool == nil {
		return nil, ErrUnavailable
	}

	if p.PriorUpdatedAt != nil {
		cur, err := s.currentUpdatedAt(ctx, p.Slug, p.Locale)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, err
		}

		if err == nil && !cur.Equal(*p.PriorUpdatedAt) {
			return nil, &ErrPrecondition{CurrentUpdatedAt: cur}
		}
	}

	var updatedBy *uuid.UUID
	if p.UpdatedBy != uuid.Nil {
		updatedBy = &p.UpdatedBy
	}

	var (
		rawContent, rawDraft []byte
		out                  Block
	)
	if p.Publish {
		err := s.Pool.QueryRow(ctx, `
			INSERT INTO content_blocks (slug, locale, content, schema_version,
				published_at, draft_content, updated_at, updated_by)
			VALUES ($1, $2, $3, $4, now(), NULL, now(), $5)
			ON CONFLICT (slug, locale) DO UPDATE
				SET content        = EXCLUDED.content,
				    schema_version = EXCLUDED.schema_version,
				    published_at   = now(),
				    draft_content  = NULL,
				    updated_at     = now(),
				    updated_by     = EXCLUDED.updated_by
			RETURNING slug, locale, content, draft_content, schema_version,
			          published_at, updated_at, updated_by`,
			p.Slug, p.Locale, []byte(p.Content), p.SchemaVersion, updatedBy,
		).Scan(&out.Slug, &out.Locale, &rawContent, &rawDraft,
			&out.SchemaVersion, &out.PublishedAt, &out.UpdatedAt, &out.UpdatedBy)
		if err != nil {
			return nil, err
		}
	} else {

		err := s.Pool.QueryRow(ctx, `
			INSERT INTO content_blocks (slug, locale, content, schema_version,
				published_at, draft_content, updated_at, updated_by)
			VALUES ($1, $2, '{}'::jsonb, $3, NULL, $4, now(), $5)
			ON CONFLICT (slug, locale) DO UPDATE
				SET draft_content  = EXCLUDED.draft_content,
				    schema_version = GREATEST(content_blocks.schema_version, EXCLUDED.schema_version),
				    updated_at     = now(),
				    updated_by     = EXCLUDED.updated_by
			RETURNING slug, locale, content, draft_content, schema_version,
			          published_at, updated_at, updated_by`,
			p.Slug, p.Locale, p.SchemaVersion, []byte(p.Content), updatedBy,
		).Scan(&out.Slug, &out.Locale, &rawContent, &rawDraft,
			&out.SchemaVersion, &out.PublishedAt, &out.UpdatedAt, &out.UpdatedBy)
		if err != nil {
			return nil, err
		}
	}
	if len(rawContent) > 0 {
		out.Content = json.RawMessage(rawContent)
	}
	if len(rawDraft) > 0 {
		out.DraftContent = json.RawMessage(rawDraft)
	}
	return &out, nil
}

func (s *PGStore) Delete(ctx context.Context, slug, locale string) (*Block, error) {
	if s == nil || s.Pool == nil {
		return nil, ErrUnavailable
	}
	row := s.Pool.QueryRow(ctx, `
		DELETE FROM content_blocks
		WHERE slug = $1 AND locale = $2
		RETURNING slug, locale, content, draft_content, schema_version,
		          published_at, updated_at, updated_by`, slug, locale)
	var b Block
	var rawContent, rawDraft []byte
	err := row.Scan(&b.Slug, &b.Locale, &rawContent, &rawDraft,
		&b.SchemaVersion, &b.PublishedAt, &b.UpdatedAt, &b.UpdatedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(rawContent) > 0 {
		b.Content = json.RawMessage(rawContent)
	}
	if len(rawDraft) > 0 {
		b.DraftContent = json.RawMessage(rawDraft)
	}
	return &b, nil
}

func (s *PGStore) ListMatrix(ctx context.Context) ([]MatrixEntry, error) {
	if s == nil || s.Pool == nil {
		return nil, ErrUnavailable
	}
	rows, err := s.Pool.Query(ctx, `
		SELECT slug, locale,
		       (published_at IS NOT NULL) AS has_published,
		       (draft_content IS NOT NULL) AS has_draft,
		       updated_at, updated_by
		FROM content_blocks
		ORDER BY slug, locale`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MatrixEntry{}
	for rows.Next() {
		var e MatrixEntry
		var ts time.Time
		if err := rows.Scan(&e.Slug, &e.Locale, &e.HasPublished, &e.HasDraft, &ts, &e.UpdatedBy); err != nil {
			return nil, err
		}
		t := ts
		e.UpdatedAt = &t
		out = append(out, e)
	}
	return out, rows.Err()
}
