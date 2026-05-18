package contentapp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/content/domain"
)

type BlockRepository interface {
	Lookup(ctx context.Context, slug, locale string, includeDraft bool) (*domain.Block, error)
	Upsert(ctx context.Context, p domain.UpsertParams) (*domain.Block, error)
	Delete(ctx context.Context, slug, locale string) (*domain.Block, error)
	ListMatrix(ctx context.Context) ([]domain.MatrixEntry, error)
}

type CachedPublished struct {
	Slug          string
	Locale        string
	Content       json.RawMessage
	SchemaVersion int
	PublishedAt   *time.Time
	UpdatedAt     time.Time
	Source        string
	ETag          string
}

type PublishedCache interface {
	Lookup(ctx context.Context, slug, locale string) (*CachedPublished, bool, error)
	Store(ctx context.Context, slug, locale string, entry CachedPublished, ttl time.Duration) error
	InvalidateSlug(ctx context.Context, slug string) error
}

type Catalog interface {
	IsKnownSlug(slug string) bool
	IsSupportedLocale(locale string) bool
	LookupSlug(slug string) (domain.SlugDescriptor, bool)
	FallbackLocales(requested string) []string
	AllowedSlugs() []domain.SlugDescriptor
	SupportedLocales() []string
	ValidateContent(slug string, raw []byte) error
}

type AuditSink interface {
	LogPublished(ctx context.Context, actorID uuid.UUID, actorEmail, target string, meta map[string]any)
	LogDraftSaved(ctx context.Context, actorID uuid.UUID, actorEmail, target string, meta map[string]any)
	LogReverted(ctx context.Context, actorID uuid.UUID, actorEmail, target string, meta map[string]any)
	LogSaveRejected(ctx context.Context, actorID uuid.UUID, actorEmail, target string, meta map[string]any)
	LogPreview(ctx context.Context, actorID uuid.UUID, target string, meta map[string]any)
	LogAccessDenied(ctx context.Context, actorID uuid.UUID, method, path string)
}

type SuperAdminGate interface {
	IsSuperAdmin(ctx context.Context, userID uuid.UUID) bool
}
