package content

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
)

const (
	ActionContentPublished       = audit.ActionContentPublished
	ActionContentDraftSaved      = audit.ActionContentDraftSaved
	ActionContentRevertedDefault = audit.ActionContentRevertedDefault
	ActionContentSaveRejected    = audit.ActionContentSaveRejected
	ActionContentAccessDenied    = audit.ActionContentAccessDenied
	ActionContentPreviewAccessed = audit.ActionContentPreviewAccessed
)

const MaxContentBytes = 256 * 1024

type Service struct {
	Store     Store
	Cache     *Cache
	Audit     audit.Service
	Logger    *zap.Logger
	JWTSecret string

	SuperAdminCheck func(ctx context.Context, userID uuid.UUID) bool

	Pool *pgxpool.Pool
}

func (s *Service) RegisterPublicRoute(app *fiber.App) {
	app.Get("/v1/content/:slug", s.handlePublicGet)
}

func (s *Service) RegisterAdminRoutes(router fiber.Router) {
	router.Get("/admin/content", s.handleAdminMatrix)
	router.Get("/admin/content/:slug/:locale", s.handleAdminGet)
	router.Put("/admin/content/:slug/:locale", s.handleAdminUpsert)
	router.Delete("/admin/content/:slug/:locale", s.handleAdminDelete)
}

type publicResponse struct {
	Slug          string          `json:"slug"`
	Locale        string          `json:"locale"`
	Content       json.RawMessage `json:"content"`
	SchemaVersion int             `json:"schema_version"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func (s *Service) handlePublicGet(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if !IsKnownSlug(slug) {
		c.Set(fiber.HeaderCacheControl, "no-store")
		return httperr.NotFound(c, "unknown_slug", "unknown content slug: "+slug)
	}
	locale := c.Query("locale")
	if locale == "" {

		locale = "ru"
	}
	if !IsSupportedLocale(locale) {
		c.Set(fiber.HeaderCacheControl, "no-store")
		return httperr.BadRequest(c, "invalid_locale",
			"locale must be one of: en, ru, kk, es")
	}

	if c.Query("preview") == "1" {
		return s.handlePublicPreview(c, slug, locale)
	}

	if s.Store == nil {
		c.Set(fiber.HeaderCacheControl, "no-store")
		return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
	}

	if entry, hit, err := s.Cache.Lookup(c.Context(), slug, locale); hit && err == nil {
		s.writePublicHeaders(c, entry.Slug, entry.Locale, entry.ETag, entry.Source, entry.UpdatedAt)
		if reqETag := c.Get("If-None-Match"); reqETag != "" && etagsMatch(reqETag, entry.ETag) {
			return c.SendStatus(http.StatusNotModified)
		}
		return c.JSON(publicResponse{
			Slug:          entry.Slug,
			Locale:        entry.Locale,
			Content:       entry.Content,
			SchemaVersion: entry.SchemaVersion,
			PublishedAt:   entry.PublishedAt,
			UpdatedAt:     entry.UpdatedAt,
		})
	} else if err != nil {
		s.warn("cache lookup failed", zap.String("slug", slug), zap.Error(err))
	}

	block, effectiveLocale, source, err := s.resolvePublished(c.Context(), slug, locale)
	if err != nil {
		if errors.Is(err, ErrUnavailable) {
			c.Set(fiber.HeaderCacheControl, "no-store")
			return httperr.Unavailable(c, "content_store_unavailable",
				"content service unavailable")
		}
		if errors.Is(err, ErrNotFound) {
			c.Set(fiber.HeaderCacheControl, "no-store")
			return httperr.NotFound(c, "not_found",
				"no published content for slug and locale (or fallbacks)")
		}
		s.warn("public lookup failed", zap.String("slug", slug), zap.String("locale", locale), zap.Error(err))
		return httperr.Internal(c)
	}

	etag := computeETag(slug, effectiveLocale, block.Content, block.PublishedAt)

	_ = s.Cache.Store(c.Context(), slug, locale, &Entry{
		Slug:          slug,
		Locale:        effectiveLocale,
		Content:       block.Content,
		SchemaVersion: block.SchemaVersion,
		PublishedAt:   block.PublishedAt,
		UpdatedAt:     block.UpdatedAt,
		Source:        source,
		ETag:          etag,
	}, DefaultCacheTTL)

	s.writePublicHeaders(c, slug, effectiveLocale, etag, source, block.UpdatedAt)

	if reqETag := c.Get("If-None-Match"); reqETag != "" && etagsMatch(reqETag, etag) {
		return c.SendStatus(http.StatusNotModified)
	}

	return c.JSON(publicResponse{
		Slug:          slug,
		Locale:        effectiveLocale,
		Content:       block.Content,
		SchemaVersion: block.SchemaVersion,
		PublishedAt:   block.PublishedAt,
		UpdatedAt:     block.UpdatedAt,
	})
}

func (s *Service) resolvePublished(ctx context.Context, slug, requestedLocale string) (*Block, string, string, error) {
	if s.Store == nil {
		return nil, "", "", ErrUnavailable
	}

	block, err := s.Store.Lookup(ctx, slug, requestedLocale, false)
	if err == nil && block.PublishedAt != nil {
		return block, requestedLocale, "direct", nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, "", "", err
	}

	for _, fb := range FallbackLocales(requestedLocale) {
		fbBlock, ferr := s.Store.Lookup(ctx, slug, fb, false)
		if ferr == nil && fbBlock.PublishedAt != nil {
			return fbBlock, fb, "locale_fallback:" + fb, nil
		}
		if ferr != nil && !errors.Is(ferr, ErrNotFound) {
			return nil, "", "", ferr
		}
	}
	return nil, "", "", ErrNotFound
}

func (s *Service) writePublicHeaders(c *fiber.Ctx, _slug, effective, etag, source string, updatedAt time.Time) {
	c.Set(fiber.HeaderCacheControl, "public, max-age=300, s-maxage=600")
	c.Set("ETag", `"`+etag+`"`)
	c.Set(fiber.HeaderContentLanguage, effective)
	c.Set(fiber.HeaderVary, "Accept-Language")
	c.Set("X-Eop-Content-Source", source)
	c.Set("Surrogate-Key", "content:"+_slug)
	if !updatedAt.IsZero() {
		c.Set(fiber.HeaderLastModified, updatedAt.UTC().Format(http.TimeFormat))
	}
}

func (s *Service) handlePublicPreview(c *fiber.Ctx, slug, locale string) error {
	c.Set(fiber.HeaderCacheControl, "no-store")
	userID, isAdmin := s.previewActor(c)
	if !isAdmin {

		if s.Store == nil {
			return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
		}
		b, effective, source, err := s.resolvePublished(c.Context(), slug, locale)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return httperr.NotFound(c, "not_found", "no published content")
			}
			return httperr.Internal(c)
		}
		etag := computeETag(slug, effective, b.Content, b.PublishedAt)
		s.writePublicHeaders(c, slug, effective, etag, source, b.UpdatedAt)
		c.Set(fiber.HeaderCacheControl, "no-store")
		return c.JSON(publicResponse{
			Slug: slug, Locale: effective, Content: b.Content,
			SchemaVersion: b.SchemaVersion, PublishedAt: b.PublishedAt, UpdatedAt: b.UpdatedAt,
		})
	}

	if s.Store == nil {
		return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
	}
	b, err := s.Store.Lookup(c.Context(), slug, locale, true)
	if errors.Is(err, ErrNotFound) {
		return httperr.NotFound(c, "not_found", "no row for slug/locale")
	}
	if err != nil {
		s.warn("preview lookup failed", zap.String("slug", slug), zap.Error(err))
		return httperr.Internal(c)
	}
	body := b.Content
	hasDraft := len(b.DraftContent) > 0
	if hasDraft {
		body = b.DraftContent
	}
	s.Audit.LogFromCtx(c, userID, "", ActionContentPreviewAccessed,
		"content", slug+":"+locale, map[string]any{
			"slug":           slug,
			"locale":         locale,
			"draft_present":  hasDraft,
			"preview_source": "url_param",
		})
	etag := computeETag(slug, locale, body, b.PublishedAt)
	c.Set("ETag", `"`+etag+`"`)
	c.Set(fiber.HeaderContentLanguage, locale)
	c.Set("X-Eop-Content-Source", "preview")
	return c.JSON(publicResponse{
		Slug: slug, Locale: locale, Content: body,
		SchemaVersion: b.SchemaVersion, PublishedAt: b.PublishedAt, UpdatedAt: b.UpdatedAt,
	})
}

func (s *Service) previewActor(c *fiber.Ctx) (uuid.UUID, bool) {
	uid, ok := s.parseBearerJWT(c)
	if !ok {

		if claims := auth.ClaimsFromCtx(c); claims != nil {
			parsed, perr := uuid.Parse(claims.UserID)
			if perr == nil {
				uid = parsed
				ok = true
			}
		}
	}
	if !ok {
		return uuid.Nil, false
	}
	if !s.checkSuperAdmin(c.Context(), uid) {
		return uuid.Nil, false
	}
	return uid, true
}

func (s *Service) parseBearerJWT(c *fiber.Ctx) (uuid.UUID, bool) {
	h := c.Get("Authorization")
	if len(h) < 8 || !strings.EqualFold(h[:7], "Bearer ") {
		return uuid.Nil, false
	}
	tok := strings.TrimSpace(h[7:])
	if tok == "" {
		return uuid.Nil, false
	}
	claims, err := auth.ParseJWT(s.JWTSecret, tok)
	if err != nil {
		return uuid.Nil, false
	}
	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, false
	}
	return uid, true
}

func (s *Service) checkSuperAdmin(ctx context.Context, uid uuid.UUID) bool {
	if s.SuperAdminCheck != nil {
		return s.SuperAdminCheck(ctx, uid)
	}
	if s.Pool == nil {
		return false
	}
	var role string
	err := s.Pool.QueryRow(ctx, "SELECT global_role FROM users WHERE id=$1", uid).Scan(&role)
	return err == nil && role == "super_admin"
}

type adminMatrixResponse struct {
	Entries []MatrixEntry `json:"entries"`
}

func (s *Service) handleAdminMatrix(c *fiber.Ctx) error {
	if !s.guardSuperAdmin(c, "GET", "/v1/admin/content") {
		return nil
	}
	if s.Store == nil {
		return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
	}
	existing, err := s.Store.ListMatrix(c.Context())
	if err != nil {
		s.warn("admin matrix list failed", zap.Error(err))
		return httperr.Internal(c)
	}
	index := map[string]MatrixEntry{}
	for _, e := range existing {
		index[e.Slug+":"+e.Locale] = e
	}
	full := make([]MatrixEntry, 0, len(AllowedSlugs)*len(SupportedLocales))
	for _, d := range AllowedSlugs {
		for _, loc := range SupportedLocales {
			if e, ok := index[d.Slug+":"+loc]; ok {
				full = append(full, e)
				continue
			}
			full = append(full, MatrixEntry{Slug: d.Slug, Locale: loc})
		}
	}
	return c.JSON(adminMatrixResponse{Entries: full})
}

type adminBlockResponse struct {
	Slug          string          `json:"slug"`
	Locale        string          `json:"locale"`
	Content       json.RawMessage `json:"content,omitempty"`
	DraftContent  json.RawMessage `json:"draft_content,omitempty"`
	SchemaVersion int             `json:"schema_version"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
	UpdatedAt     *time.Time      `json:"updated_at,omitempty"`
	UpdatedBy     *uuid.UUID      `json:"updated_by,omitempty"`
	ETag          string          `json:"etag,omitempty"`
}

func (s *Service) handleAdminGet(c *fiber.Ctx) error {
	slug := c.Params("slug")
	locale := c.Params("locale")
	if !IsKnownSlug(slug) {
		return httperr.NotFound(c, "unknown_slug", "unknown content slug: "+slug)
	}
	if !IsSupportedLocale(locale) {
		return httperr.BadRequest(c, "invalid_locale",
			"locale must be one of: en, ru, kk, es")
	}
	if !s.guardSuperAdmin(c, "GET", "/v1/admin/content/"+slug+"/"+locale) {
		return nil
	}
	if s.Store == nil {
		return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
	}
	b, err := s.Store.Lookup(c.Context(), slug, locale, true)
	if errors.Is(err, ErrNotFound) {

		desc, _ := LookupSlug(slug)
		return c.JSON(adminBlockResponse{
			Slug:          slug,
			Locale:        locale,
			SchemaVersion: desc.SchemaVersion,
		})
	}
	if err != nil {
		s.warn("admin get failed", zap.String("slug", slug), zap.String("locale", locale), zap.Error(err))
		return httperr.Internal(c)
	}
	resp := adminBlockResponse{
		Slug:          b.Slug,
		Locale:        b.Locale,
		SchemaVersion: b.SchemaVersion,
		PublishedAt:   b.PublishedAt,
	}
	if b.PublishedAt != nil {
		resp.Content = b.Content
	}
	if len(b.DraftContent) > 0 {
		resp.DraftContent = b.DraftContent
	}
	t := b.UpdatedAt
	resp.UpdatedAt = &t
	resp.UpdatedBy = b.UpdatedBy
	etag := b.UpdatedAt.UTC().Format(time.RFC3339Nano)
	resp.ETag = etag
	c.Set("ETag", `"`+etag+`"`)
	return c.JSON(resp)
}

type adminUpsertReq struct {
	Content json.RawMessage `json:"content"`
	Publish bool            `json:"publish"`
}

func (s *Service) handleAdminUpsert(c *fiber.Ctx) error {
	slug := c.Params("slug")
	locale := c.Params("locale")
	if !IsKnownSlug(slug) {
		return httperr.NotFound(c, "unknown_slug", "unknown content slug: "+slug)
	}
	if !IsSupportedLocale(locale) {
		return httperr.BadRequest(c, "invalid_locale",
			"locale must be one of: en, ru, kk, es")
	}
	if !s.guardSuperAdmin(c, "PUT", "/v1/admin/content/"+slug+"/"+locale) {
		return nil
	}
	if s.Store == nil {
		return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
	}
	desc, _ := LookupSlug(slug)

	actorID, actorEmail := s.actorInfo(c)
	targetID := slug + ":" + locale

	var req adminUpsertReq
	if err := c.BodyParser(&req); err != nil {
		s.Audit.LogFromCtx(c, actorID, actorEmail, ActionContentSaveRejected,
			"content", targetID, map[string]any{
				"error_code":   "invalid_json",
				"error_detail": err.Error(),
			})
		return httperr.BadRequest(c, "invalid_json", "invalid JSON body")
	}
	if len(req.Content) == 0 {
		s.Audit.LogFromCtx(c, actorID, actorEmail, ActionContentSaveRejected,
			"content", targetID, map[string]any{
				"error_code":   "schema_violation",
				"error_detail": "missing content field",
			})
		return httperr.BadRequest(c, "schema_violation", "content field required")
	}
	if len(req.Content) > MaxContentBytes {
		s.Audit.LogFromCtx(c, actorID, actorEmail, ActionContentSaveRejected,
			"content", targetID, map[string]any{
				"error_code":   "too_large",
				"error_detail": fmt.Sprintf("content exceeds %d bytes", MaxContentBytes),
			})
		return httperr.TooLarge(c, "too_large",
			fmt.Sprintf("content exceeds %d KB", MaxContentBytes/1024))
	}

	if verr := desc.Validate(req.Content); verr != nil {
		meta := map[string]any{
			"error_code":   verr.Code,
			"error_detail": verr.Detail,
		}
		if verr.Field != "" {
			meta["field"] = verr.Field
		}
		if verr.SchemaPath != "" {
			meta["schema_path"] = verr.SchemaPath
		}
		s.Audit.LogFromCtx(c, actorID, actorEmail, ActionContentSaveRejected,
			"content", targetID, meta)
		pd := httperr.ProblemDetails{
			Status: fiber.StatusUnprocessableEntity,
			Code:   verr.Code,
			Detail: verr.Detail,
			Extensions: map[string]any{
				"field":       verr.Field,
				"schema_path": verr.SchemaPath,
			},
		}
		return httperr.Send(c, pd)
	}

	var priorTS *time.Time
	if rawIM := c.Get("If-Match"); rawIM != "" {
		ts, ok := parseIfMatchTimestamp(rawIM)
		if ok {
			priorTS = &ts
		}

		if !ok {
			cur, err := s.currentUpdatedAt(c.Context(), slug, locale)
			if err == nil {
				s.Audit.LogFromCtx(c, actorID, actorEmail, ActionContentSaveRejected,
					"content", targetID, map[string]any{
						"error_code":   "precondition_failed",
						"current_etag": cur.UTC().Format(time.RFC3339Nano),
					})
				return s.precondFailedResp(c, cur)
			}

		}
	}

	out, err := s.Store.Upsert(c.Context(), UpsertParams{
		Slug:           slug,
		Locale:         locale,
		Content:        req.Content,
		Publish:        req.Publish,
		SchemaVersion:  desc.SchemaVersion,
		UpdatedBy:      actorID,
		PriorUpdatedAt: priorTS,
	})
	if err != nil {
		var pe *ErrPrecondition
		if errors.As(err, &pe) {
			s.Audit.LogFromCtx(c, actorID, actorEmail, ActionContentSaveRejected,
				"content", targetID, map[string]any{
					"error_code":   "precondition_failed",
					"current_etag": pe.CurrentUpdatedAt.UTC().Format(time.RFC3339Nano),
				})
			return s.precondFailedResp(c, pe.CurrentUpdatedAt)
		}
		if errors.Is(err, ErrUnavailable) {
			return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
		}
		s.warn("admin upsert failed", zap.String("slug", slug), zap.String("locale", locale), zap.Error(err))
		return httperr.Internal(c)
	}

	_ = s.Cache.InvalidateSlug(c.Context(), slug)

	if req.Publish {
		s.Audit.LogFromCtx(c, actorID, actorEmail, ActionContentPublished,
			"content", targetID, map[string]any{
				"schema_version":     out.SchemaVersion,
				"content_hash":       sha256Hex(out.Content),
				"content_size_bytes": len(out.Content),
				"source":             "direct_publish",
			})
	} else {
		s.Audit.LogFromCtx(c, actorID, actorEmail, ActionContentDraftSaved,
			"content", targetID, map[string]any{
				"schema_version":     out.SchemaVersion,
				"content_hash":       sha256Hex(out.DraftContent),
				"content_size_bytes": len(out.DraftContent),
			})
	}

	resp := adminBlockResponse{
		Slug:          out.Slug,
		Locale:        out.Locale,
		SchemaVersion: out.SchemaVersion,
		PublishedAt:   out.PublishedAt,
		UpdatedAt:     timePtr(out.UpdatedAt),
		UpdatedBy:     out.UpdatedBy,
		ETag:          out.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if out.PublishedAt != nil {
		resp.Content = out.Content
	}
	if len(out.DraftContent) > 0 {
		resp.DraftContent = out.DraftContent
	}
	return c.JSON(resp)
}

func (s *Service) handleAdminDelete(c *fiber.Ctx) error {
	slug := c.Params("slug")
	locale := c.Params("locale")
	if !IsKnownSlug(slug) {
		return httperr.NotFound(c, "unknown_slug", "unknown content slug: "+slug)
	}
	if !IsSupportedLocale(locale) {
		return httperr.BadRequest(c, "invalid_locale",
			"locale must be one of: en, ru, kk, es")
	}
	if !s.guardSuperAdmin(c, "DELETE", "/v1/admin/content/"+slug+"/"+locale) {
		return nil
	}
	if s.Store == nil {
		return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
	}
	prior, err := s.Store.Delete(c.Context(), slug, locale)
	if err != nil {
		if errors.Is(err, ErrUnavailable) {
			return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
		}
		s.warn("admin delete failed", zap.String("slug", slug), zap.String("locale", locale), zap.Error(err))
		return httperr.Internal(c)
	}
	if prior == nil {
		return httperr.NotFound(c, "row_absent", "no row exists for slug and locale")
	}

	_ = s.Cache.InvalidateSlug(c.Context(), slug)

	actorID, actorEmail := s.actorInfo(c)
	meta := map[string]any{
		"prior_published_hash": sha256Hex(prior.Content),
	}
	if len(prior.DraftContent) > 0 {
		meta["prior_draft_hash"] = sha256Hex(prior.DraftContent)
	}
	if len(prior.Content) <= 4*1024 {
		meta["prior_content_inline"] = prior.Content
	}
	s.Audit.LogFromCtx(c, actorID, actorEmail, ActionContentRevertedDefault,
		"content", slug+":"+locale, meta)

	return c.JSON(fiber.Map{"reverted": true})
}

func (s *Service) guardSuperAdmin(c *fiber.Ctx, method, path string) bool {
	uid, ok := s.parseBearerJWT(c)
	if !ok {
		if claims := auth.ClaimsFromCtx(c); claims != nil {
			parsed, perr := uuid.Parse(claims.UserID)
			if perr == nil {
				uid = parsed
				ok = true
			}
		}
	}
	if !ok {
		_ = httperr.Unauthorized(c, "auth_required", "authentication required")
		return false
	}
	if !s.checkSuperAdmin(c.Context(), uid) {
		s.Audit.LogFromCtx(c, uid, "", ActionContentAccessDenied,
			"content", "", map[string]any{
				"method": method,
				"path":   path,
				"status": 403,
			})
		_ = httperr.Forbidden(c, "super_admin_required", "super_admin only")
		return false
	}
	return true
}

func (s *Service) actorInfo(c *fiber.Ctx) (uuid.UUID, string) {
	if claims := auth.ClaimsFromCtx(c); claims != nil {
		uid, _ := uuid.Parse(claims.UserID)
		return uid, claims.Email
	}
	uid, _ := s.parseBearerJWT(c)
	return uid, ""
}

func (s *Service) warn(msg string, fields ...zap.Field) {
	if s.Logger == nil {
		return
	}
	s.Logger.Warn(msg, fields...)
}

func (s *Service) currentUpdatedAt(ctx context.Context, slug, locale string) (time.Time, error) {
	if s.Store == nil {
		return time.Time{}, ErrUnavailable
	}
	b, err := s.Store.Lookup(ctx, slug, locale, false)
	if err != nil {
		return time.Time{}, err
	}
	return b.UpdatedAt, nil
}

func (s *Service) precondFailedResp(c *fiber.Ctx, cur time.Time) error {
	pd := httperr.ProblemDetails{
		Status: fiber.StatusPreconditionFailed,
		Code:   "precondition_failed",
		Detail: "content was updated by another admin; reload to see latest",
		Extensions: map[string]any{
			"current_etag": cur.UTC().Format(time.RFC3339Nano),
		},
	}
	return httperr.Send(c, pd)
}

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

func parseIfMatchTimestamp(raw string) (time.Time, bool) {
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

func sha256Hex(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
