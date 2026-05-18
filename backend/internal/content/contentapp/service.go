package contentapp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/content/domain"
)

const DefaultCacheTTL = 5 * time.Minute

type Service struct {
	repo   BlockRepository
	cache  PublishedCache
	cat    Catalog
	audit  AuditSink
	admins SuperAdminGate
}

type Deps struct {
	Repo   BlockRepository
	Cache  PublishedCache
	Cat    Catalog
	Audit  AuditSink
	Admins SuperAdminGate
}

func New(d Deps) *Service {
	return &Service{
		repo:   d.Repo,
		cache:  d.Cache,
		cat:    d.Cat,
		audit:  d.Audit,
		admins: d.Admins,
	}
}

func (s *Service) GetPublished(ctx context.Context, slug, locale string, ifNoneMatch string) (*PublicView, error) {
	if s.cat != nil && !s.cat.IsKnownSlug(slug) {
		return nil, domain.ErrNotFound
	}
	if s.cat != nil && !s.cat.IsSupportedLocale(locale) {
		return nil, ErrInvalidLocale
	}
	if s.repo == nil {
		return nil, domain.ErrUnavailable
	}

	if s.cache != nil {
		if hit := s.tryCache(ctx, slug, locale, ifNoneMatch); hit != nil {
			return hit, nil
		}
	}

	block, effective, source, err := s.resolvePublished(ctx, slug, locale)
	if err != nil {
		return nil, err
	}
	etag := computeETag(slug, effective, block.Content, block.PublishedAt)
	if ifNoneMatch != "" && etagsMatch(ifNoneMatch, etag) {
		return &PublicView{
			Slug: slug, Locale: effective, ETag: etag, Source: source,
			UpdatedAt: block.UpdatedAt, NotModified: true,
		}, nil
	}
	if s.cache != nil {
		_ = s.cache.Store(ctx, slug, locale, CachedPublished{
			Slug: slug, Locale: effective, Content: block.Content,
			SchemaVersion: block.SchemaVersion, PublishedAt: block.PublishedAt,
			UpdatedAt: block.UpdatedAt, Source: source, ETag: etag,
		}, DefaultCacheTTL)
	}
	return &PublicView{
		Slug:          slug,
		Locale:        effective,
		Content:       block.Content,
		SchemaVersion: block.SchemaVersion,
		PublishedAt:   block.PublishedAt,
		UpdatedAt:     block.UpdatedAt,
		ETag:          etag,
		Source:        source,
	}, nil
}

func (s *Service) tryCache(ctx context.Context, slug, locale, ifNoneMatch string) *PublicView {
	entry, ok, err := s.cache.Lookup(ctx, slug, locale)
	if err != nil || !ok {
		return nil
	}
	if ifNoneMatch != "" && etagsMatch(ifNoneMatch, entry.ETag) {
		return &PublicView{
			Slug: entry.Slug, Locale: entry.Locale, ETag: entry.ETag, Source: entry.Source,
			UpdatedAt: entry.UpdatedAt, NotModified: true, CacheHit: true,
		}
	}
	return &PublicView{
		Slug:          entry.Slug,
		Locale:        entry.Locale,
		Content:       entry.Content,
		SchemaVersion: entry.SchemaVersion,
		PublishedAt:   entry.PublishedAt,
		UpdatedAt:     entry.UpdatedAt,
		ETag:          entry.ETag,
		Source:        entry.Source,
		CacheHit:      true,
	}
}

func (s *Service) GetPreview(ctx context.Context, slug, locale string, actorID uuid.UUID, isAdmin bool) (*PreviewView, error) {
	if s.cat != nil && !s.cat.IsKnownSlug(slug) {
		return nil, domain.ErrNotFound
	}
	if s.repo == nil {
		return nil, domain.ErrUnavailable
	}
	if !isAdmin {
		pub, err := s.GetPublished(ctx, slug, locale, "")
		if err != nil {
			return nil, err
		}
		pub.Source = "published_fallback"
		return &PreviewView{PublicView: *pub}, nil
	}
	b, err := s.repo.Lookup(ctx, slug, locale, true)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	body := b.Content
	hasDraft := len(b.DraftContent) > 0
	if hasDraft {
		body = b.DraftContent
	}
	if s.audit != nil && actorID != uuid.Nil {
		s.audit.LogPreview(ctx, actorID, slug+":"+locale, map[string]any{
			"slug": slug, "locale": locale, "draft_present": hasDraft, "preview_source": "url_param",
		})
	}
	etag := computeETag(slug, locale, body, b.PublishedAt)
	return &PreviewView{PublicView: PublicView{
		Slug: slug, Locale: locale, Content: body, SchemaVersion: b.SchemaVersion,
		PublishedAt: b.PublishedAt, UpdatedAt: b.UpdatedAt, ETag: etag, Source: "preview",
	}}, nil
}

func (s *Service) ListAdminMatrix(ctx context.Context) ([]MatrixEntryView, error) {
	if s.repo == nil {
		return nil, domain.ErrUnavailable
	}
	existing, err := s.repo.ListMatrix(ctx)
	if err != nil {
		return nil, err
	}
	index := map[string]domain.MatrixEntry{}
	for _, e := range existing {
		index[e.Slug+":"+e.Locale] = e
	}
	slugs := s.cat.AllowedSlugs()
	locales := s.cat.SupportedLocales()
	full := make([]MatrixEntryView, 0, len(slugs)*len(locales))
	for _, d := range slugs {
		for _, loc := range locales {
			if e, ok := index[d.Slug+":"+loc]; ok {
				full = append(full, matrixToView(e))
				continue
			}
			full = append(full, MatrixEntryView{Slug: d.Slug, Locale: loc})
		}
	}
	return full, nil
}

func (s *Service) GetAdminBlock(ctx context.Context, slug, locale string) (*AdminBlockView, error) {
	if s.repo == nil {
		return nil, domain.ErrUnavailable
	}
	b, err := s.repo.Lookup(ctx, slug, locale, true)
	if errors.Is(err, domain.ErrNotFound) {
		desc, _ := s.cat.LookupSlug(slug)
		return &AdminBlockView{Slug: slug, Locale: locale, SchemaVersion: desc.SchemaVersion, Empty: true}, nil
	}
	if err != nil {
		return nil, err
	}
	return blockToAdminView(b), nil
}

func (s *Service) UpsertAdmin(ctx context.Context, cmd UpsertCommand) (*UpsertResult, error) {
	if s.repo == nil {
		return nil, domain.ErrUnavailable
	}
	target := cmd.Slug + ":" + cmd.Locale
	desc, ok := s.cat.LookupSlug(cmd.Slug)
	if !ok {
		return nil, domain.ErrNotFound
	}
	if len(cmd.Content) == 0 {
		s.rejectSave(ctx, cmd, target, "schema_violation", "missing content field")
		return nil, &SchemaViolation{Code: "schema_violation", Detail: "content field required"}
	}
	if len(cmd.Content) > MaxContentBytes {
		s.rejectSave(ctx, cmd, target, "too_large", fmt.Sprintf("content exceeds %d bytes", MaxContentBytes))
		return nil, ErrContentTooLarge
	}
	if err := s.cat.ValidateContent(cmd.Slug, cmd.Content); err != nil {
		var se *domain.SchemaError
		if errors.As(err, &se) {
			meta := map[string]any{"error_code": se.Code, "error_detail": se.Detail}
			if se.Field != "" {
				meta["field"] = se.Field
			}
			if se.SchemaPath != "" {
				meta["schema_path"] = se.SchemaPath
			}
			s.rejectSave(ctx, cmd, target, se.Code, se.Detail)
			return nil, &SchemaViolation{Code: se.Code, Field: se.Field, Detail: se.Detail, SchemaPath: se.SchemaPath}
		}
		return nil, err
	}
	out, err := s.repo.Upsert(ctx, domain.UpsertParams{
		Slug: cmd.Slug, Locale: cmd.Locale, Content: cmd.Content, Publish: cmd.Publish,
		SchemaVersion: desc.SchemaVersion, UpdatedBy: cmd.ActorID, PriorUpdatedAt: cmd.PriorUpdatedAt,
	})
	if err != nil {
		var pe *domain.ErrPrecondition
		if errors.As(err, &pe) {
			s.rejectSave(ctx, cmd, target, "precondition_failed", pe.CurrentUpdatedAt.UTC().Format(time.RFC3339Nano))
			return nil, pe
		}
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.InvalidateSlug(ctx, cmd.Slug)
	}
	view := blockToAdminView(out)
	res := &UpsertResult{AdminBlockView: *view, Published: cmd.Publish}
	if cmd.Publish {
		s.logPublished(ctx, cmd, target, out)
	} else {
		s.logDraft(ctx, cmd, target, out)
	}
	return res, nil
}

func (s *Service) DeleteAdmin(ctx context.Context, slug, locale string, actorID uuid.UUID, actorEmail string) (*DeleteResult, error) {
	if s.repo == nil {
		return nil, domain.ErrUnavailable
	}
	prior, err := s.repo.Delete(ctx, slug, locale)
	if err != nil {
		return nil, err
	}
	if prior == nil {
		return nil, domain.ErrNotFound
	}
	if s.cache != nil {
		_ = s.cache.InvalidateSlug(ctx, slug)
	}
	if s.audit != nil {
		meta := map[string]any{"prior_published_hash": sha256Hex(prior.Content)}
		if len(prior.DraftContent) > 0 {
			meta["prior_draft_hash"] = sha256Hex(prior.DraftContent)
		}
		if len(prior.Content) <= 4*1024 {
			meta["prior_content_inline"] = prior.Content
		}
		s.audit.LogReverted(ctx, actorID, actorEmail, slug+":"+locale, meta)
	}
	return &DeleteResult{Reverted: true, Prior: *blockToAdminView(prior)}, nil
}

func (s *Service) CheckSuperAdmin(ctx context.Context, userID uuid.UUID) bool {
	if s.admins == nil {
		return false
	}
	return s.admins.IsSuperAdmin(ctx, userID)
}

func (s *Service) LogAccessDenied(ctx context.Context, userID uuid.UUID, method, path string) {
	if s.audit != nil {
		s.audit.LogAccessDenied(ctx, userID, method, path)
	}
}

func (s *Service) CurrentUpdatedAt(ctx context.Context, slug, locale string) (time.Time, error) {
	if s.repo == nil {
		return time.Time{}, domain.ErrUnavailable
	}
	b, err := s.repo.Lookup(ctx, slug, locale, false)
	if err != nil {
		return time.Time{}, err
	}
	return b.UpdatedAt, nil
}

func (s *Service) resolvePublished(ctx context.Context, slug, requestedLocale string) (*domain.Block, string, string, error) {
	block, err := s.repo.Lookup(ctx, slug, requestedLocale, false)
	if err == nil && block.PublishedAt != nil {
		return block, requestedLocale, "direct", nil
	}
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, "", "", err
	}
	for _, fb := range s.cat.FallbackLocales(requestedLocale) {
		fbBlock, ferr := s.repo.Lookup(ctx, slug, fb, false)
		if ferr == nil && fbBlock.PublishedAt != nil {
			return fbBlock, fb, "locale_fallback:" + fb, nil
		}
		if ferr != nil && !errors.Is(ferr, domain.ErrNotFound) {
			return nil, "", "", ferr
		}
	}
	return nil, "", "", domain.ErrNotFound
}

func (s *Service) rejectSave(ctx context.Context, cmd UpsertCommand, target, code, detail string) {
	if s.audit == nil {
		return
	}
	s.audit.LogSaveRejected(ctx, cmd.ActorID, cmd.ActorEmail, target, map[string]any{
		"error_code": code, "error_detail": detail,
	})
}

func (s *Service) logPublished(ctx context.Context, cmd UpsertCommand, target string, out *domain.Block) {
	if s.audit == nil {
		return
	}
	s.audit.LogPublished(ctx, cmd.ActorID, cmd.ActorEmail, target, map[string]any{
		"schema_version": out.SchemaVersion, "content_hash": sha256Hex(out.Content),
		"content_size_bytes": len(out.Content), "source": "direct_publish",
	})
}

func (s *Service) logDraft(ctx context.Context, cmd UpsertCommand, target string, out *domain.Block) {
	if s.audit == nil {
		return
	}
	s.audit.LogDraftSaved(ctx, cmd.ActorID, cmd.ActorEmail, target, map[string]any{
		"schema_version": out.SchemaVersion, "content_hash": sha256Hex(out.DraftContent),
		"content_size_bytes": len(out.DraftContent),
	})
}

func matrixToView(e domain.MatrixEntry) MatrixEntryView {
	return MatrixEntryView{
		Slug: e.Slug, Locale: e.Locale, HasPublished: e.HasPublished, HasDraft: e.HasDraft,
		UpdatedAt: e.UpdatedAt, UpdatedBy: e.UpdatedBy,
	}
}

func blockToAdminView(b *domain.Block) *AdminBlockView {
	v := &AdminBlockView{
		Slug: b.Slug, Locale: b.Locale, SchemaVersion: b.SchemaVersion, PublishedAt: b.PublishedAt,
	}
	if b.PublishedAt != nil {
		v.Content = b.Content
	}
	if len(b.DraftContent) > 0 {
		v.DraftContent = b.DraftContent
	}
	if !b.UpdatedAt.IsZero() {
		t := b.UpdatedAt
		v.UpdatedAt = &t
		v.ETag = b.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	v.UpdatedBy = b.UpdatedBy
	return v
}

var ErrInvalidLocale = errors.New("invalid locale")
var ErrContentTooLarge = errors.New("content too large")

func computeETag(slug, locale string, content json.RawMessage, publishedAt *time.Time) string {
	h := sha256.New()
	h.Write([]byte(slug))
	h.Write([]byte{0})
	h.Write([]byte(locale))
	h.Write([]byte{0})
	h.Write(content)
	if publishedAt != nil {
		h.Write([]byte{0})
		h.Write([]byte(publishedAt.UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func etagsMatch(clientHeader, serverETag string) bool {
	parts := strings.Split(clientHeader, ",")
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == "*" {
			return true
		}
		v = strings.TrimPrefix(v, "W/")
		v = strings.Trim(v, `"`)
		if v == serverETag {
			return true
		}
	}
	return false
}

func sha256Hex(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func ParseIfMatchTimestamp(raw string) (time.Time, bool) {
	v := strings.TrimSpace(raw)
	v = strings.TrimPrefix(v, "W/")
	v = strings.Trim(v, `"`)
	if ts, err := time.Parse(time.RFC3339Nano, v); err == nil {
		return ts, true
	}
	if ts, err := time.Parse(time.RFC3339, v); err == nil {
		return ts, true
	}
	return time.Time{}, false
}
