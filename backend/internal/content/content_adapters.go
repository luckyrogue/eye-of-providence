package content

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/content/contentapp"
	"github.com/eye-of-providence/backend/internal/content/domain"
)

type domainCatalog struct{}

func (domainCatalog) IsKnownSlug(slug string) bool       { return domain.IsKnownSlug(slug) }
func (domainCatalog) IsSupportedLocale(loc string) bool  { return domain.IsSupportedLocale(loc) }
func (domainCatalog) LookupSlug(slug string) (domain.SlugDescriptor, bool) {
	return domain.LookupSlug(slug)
}
func (domainCatalog) FallbackLocales(r string) []string { return domain.FallbackLocales(r) }
func (domainCatalog) AllowedSlugs() []domain.SlugDescriptor {
	return domain.AllowedSlugs
}
func (domainCatalog) SupportedLocales() []string { return domain.SupportedLocales }
func (domainCatalog) ValidateContent(slug string, raw []byte) error {
	return domain.Validate(slug, json.RawMessage(raw))
}

type cachePort struct{ c *Cache }

func (a cachePort) Lookup(ctx context.Context, slug, locale string) (*contentapp.CachedPublished, bool, error) {
	if a.c == nil {
		return nil, false, nil
	}
	e, hit, err := a.c.Lookup(ctx, slug, locale)
	if err != nil || !hit || e == nil {
		return nil, hit, err
	}
	return &contentapp.CachedPublished{
		Slug: e.Slug, Locale: e.Locale, Content: e.Content, SchemaVersion: e.SchemaVersion,
		PublishedAt: e.PublishedAt, UpdatedAt: e.UpdatedAt, Source: e.Source, ETag: e.ETag,
	}, true, nil
}

func (a cachePort) Store(ctx context.Context, slug, locale string, entry contentapp.CachedPublished, ttl time.Duration) error {
	if a.c == nil {
		return nil
	}
	return a.c.Store(ctx, slug, locale, &Entry{
		Slug: entry.Slug, Locale: entry.Locale, Content: entry.Content,
		SchemaVersion: entry.SchemaVersion, PublishedAt: entry.PublishedAt,
		UpdatedAt: entry.UpdatedAt, Source: entry.Source, ETag: entry.ETag,
	}, ttl)
}

func (a cachePort) InvalidateSlug(ctx context.Context, slug string) error {
	if a.c == nil {
		return nil
	}
	return a.c.InvalidateSlug(ctx, slug)
}

type auditPort struct {
	s audit.Service
}

func (a auditPort) LogPublished(ctx context.Context, actorID uuid.UUID, email, target string, meta map[string]any) {
	a.s.Log(ctx, audit.Entry{ActorID: actorID, ActorEmail: email, Action: audit.ActionContentPublished, TargetType: "content", TargetID: target, Metadata: meta})
}
func (a auditPort) LogDraftSaved(ctx context.Context, actorID uuid.UUID, email, target string, meta map[string]any) {
	a.s.Log(ctx, audit.Entry{ActorID: actorID, ActorEmail: email, Action: audit.ActionContentDraftSaved, TargetType: "content", TargetID: target, Metadata: meta})
}
func (a auditPort) LogReverted(ctx context.Context, actorID uuid.UUID, email, target string, meta map[string]any) {
	a.s.Log(ctx, audit.Entry{ActorID: actorID, ActorEmail: email, Action: audit.ActionContentRevertedDefault, TargetType: "content", TargetID: target, Metadata: meta})
}
func (a auditPort) LogSaveRejected(ctx context.Context, actorID uuid.UUID, email, target string, meta map[string]any) {
	a.s.Log(ctx, audit.Entry{ActorID: actorID, ActorEmail: email, Action: audit.ActionContentSaveRejected, TargetType: "content", TargetID: target, Metadata: meta})
}
func (a auditPort) LogPreview(ctx context.Context, actorID uuid.UUID, target string, meta map[string]any) {
	a.s.Log(ctx, audit.Entry{ActorID: actorID, Action: audit.ActionContentPreviewAccessed, TargetType: "content", TargetID: target, Metadata: meta})
}
func (a auditPort) LogAccessDenied(ctx context.Context, actorID uuid.UUID, method, path string) {
	a.s.Log(ctx, audit.Entry{ActorID: actorID, Action: audit.ActionContentAccessDenied, TargetType: "content", Metadata: map[string]any{
		"method": method, "path": path, "status": 403,
	}})
}

type superAdminPort struct {
	pool *pgxpool.Pool
	fn   func(ctx context.Context, userID uuid.UUID) bool
}

func (p superAdminPort) IsSuperAdmin(ctx context.Context, uid uuid.UUID) bool {
	if p.fn != nil {
		return p.fn(ctx, uid)
	}
	if p.pool == nil {
		return false
	}
	var role string
	err := p.pool.QueryRow(ctx, "SELECT global_role FROM users WHERE id=$1", uid).Scan(&role)
	return err == nil && role == "super_admin"
}

func newContentApp(repo *PGStore, cache *Cache, auditSvc audit.Service, pool *pgxpool.Pool, superFn func(context.Context, uuid.UUID) bool) *contentapp.Service {
	return contentapp.New(contentapp.Deps{
		Repo:   repo,
		Cache:  cachePort{c: cache},
		Cat:    domainCatalog{},
		Audit:  auditPort{s: auditSvc},
		Admins: superAdminPort{pool: pool, fn: superFn},
	})
}
