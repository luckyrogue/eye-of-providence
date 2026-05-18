package content

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/content/domain"
)

type PGStore struct {
	Pool *pgxpool.Pool
}

func NewPGStore(pool *pgxpool.Pool) *PGStore {
	return &PGStore{Pool: pool}
}

func (s *PGStore) Lookup(ctx context.Context, slug, locale string, includeDraft bool) (*domain.Block, error) {
	if s == nil || s.Pool == nil {
		return nil, domain.ErrUnavailable
	}
	row := s.Pool.QueryRow(ctx, `
		SELECT slug, locale, content, draft_content, schema_version,
		       published_at, updated_at, updated_by
		FROM content_blocks
		WHERE slug = $1 AND locale = $2`, slug, locale)
	var b domain.Block
	var rawContent, rawDraft []byte
	if err := row.Scan(&b.Slug, &b.Locale, &rawContent, &rawDraft,
		&b.SchemaVersion, &b.PublishedAt, &b.UpdatedAt, &b.UpdatedBy); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
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
		return time.Time{}, domain.ErrNotFound
	}
	return t, err
}

func (s *PGStore) Upsert(ctx context.Context, p domain.UpsertParams) (*domain.Block, error) {
	if s == nil || s.Pool == nil {
		return nil, domain.ErrUnavailable
	}
	if p.PriorUpdatedAt != nil {
		cur, err := s.currentUpdatedAt(ctx, p.Slug, p.Locale)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
		if err == nil && !cur.Equal(*p.PriorUpdatedAt) {
			return nil, &domain.ErrPrecondition{CurrentUpdatedAt: cur}
		}
	}
	var updatedBy *uuid.UUID
	if p.UpdatedBy != uuid.Nil {
		updatedBy = &p.UpdatedBy
	}
	var rawContent, rawDraft []byte
	var out domain.Block
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

func (s *PGStore) Delete(ctx context.Context, slug, locale string) (*domain.Block, error) {
	if s == nil || s.Pool == nil {
		return nil, domain.ErrUnavailable
	}
	row := s.Pool.QueryRow(ctx, `
		DELETE FROM content_blocks
		WHERE slug = $1 AND locale = $2
		RETURNING slug, locale, content, draft_content, schema_version,
		          published_at, updated_at, updated_by`, slug, locale)
	var b domain.Block
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

func (s *PGStore) ListMatrix(ctx context.Context) ([]domain.MatrixEntry, error) {
	if s == nil || s.Pool == nil {
		return nil, domain.ErrUnavailable
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
	out := []domain.MatrixEntry{}
	for rows.Next() {
		var e domain.MatrixEntry
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
