package content

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/content/contentapp"
	"github.com/eye-of-providence/backend/internal/content/domain"
	"github.com/eye-of-providence/backend/internal/httperr"
)

type Handler struct {
	app       *contentapp.Service
	logger    *zap.Logger
	jwtSecret string
}

func (h *Handler) RegisterPublicRoute(app *fiber.App) {
	app.Get("/v1/content/:slug", h.handlePublicGet)
}

func (h *Handler) RegisterAdminRoutes(router fiber.Router) {
	router.Get("/admin/content", h.handleAdminMatrix)
	router.Get("/admin/content/:slug/:locale", h.handleAdminGet)
	router.Put("/admin/content/:slug/:locale", h.handleAdminUpsert)
	router.Delete("/admin/content/:slug/:locale", h.handleAdminDelete)
}

type publicResponse struct {
	Slug          string          `json:"slug"`
	Locale        string          `json:"locale"`
	Content       json.RawMessage `json:"content"`
	SchemaVersion int             `json:"schema_version"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func (h *Handler) handlePublicGet(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if !domain.IsKnownSlug(slug) {
		c.Set(fiber.HeaderCacheControl, "no-store")
		return httperr.NotFound(c, "unknown_slug", "unknown content slug: "+slug)
	}
	locale := c.Query("locale")
	if locale == "" {
		locale = "ru"
	}
	if !domain.IsSupportedLocale(locale) {
		c.Set(fiber.HeaderCacheControl, "no-store")
		return httperr.BadRequest(c, "invalid_locale", "locale must be one of: en, ru, kk, es")
	}
	if c.Query("preview") == "1" {
		return h.handlePublicPreview(c, slug, locale)
	}
	view, err := h.app.GetPublished(c.Context(), slug, locale, c.Get("If-None-Match"))
	if err != nil {
		return h.mapPublicErr(c, slug, locale, err)
	}
	if view.NotModified {
		h.writePublicHeaders(c, view.Slug, view.Locale, view.ETag, view.Source, view.UpdatedAt)
		return c.SendStatus(http.StatusNotModified)
	}
	h.writePublicHeaders(c, view.Slug, view.Locale, view.ETag, view.Source, view.UpdatedAt)
	return c.JSON(publicResponse{
		Slug: view.Slug, Locale: view.Locale, Content: view.Content,
		SchemaVersion: view.SchemaVersion, PublishedAt: view.PublishedAt, UpdatedAt: view.UpdatedAt,
	})
}

func (h *Handler) handlePublicPreview(c *fiber.Ctx, slug, locale string) error {
	c.Set(fiber.HeaderCacheControl, "no-store")
	uid, isAdmin := h.previewActor(c)
	view, err := h.app.GetPreview(c.Context(), slug, locale, uid, isAdmin)
	if err != nil {
		return h.mapPublicErr(c, slug, locale, err)
	}
	if view.Source == "preview" {
		c.Set("X-Eop-Content-Source", "preview")
	} else {
		h.writePublicHeaders(c, view.Slug, view.Locale, view.ETag, view.Source, view.UpdatedAt)
		c.Set(fiber.HeaderCacheControl, "no-store")
	}
	c.Set("ETag", `"`+view.ETag+`"`)
	c.Set(fiber.HeaderContentLanguage, view.Locale)
	return c.JSON(publicResponse{
		Slug: view.Slug, Locale: view.Locale, Content: view.Content,
		SchemaVersion: view.SchemaVersion, PublishedAt: view.PublishedAt, UpdatedAt: view.UpdatedAt,
	})
}

func (h *Handler) mapPublicErr(c *fiber.Ctx, slug, locale string, err error) error {
	c.Set(fiber.HeaderCacheControl, "no-store")
	if errors.Is(err, domain.ErrUnavailable) {
		return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
	}
	if errors.Is(err, domain.ErrNotFound) {
		return httperr.NotFound(c, "not_found", "no published content for slug and locale (or fallbacks)")
	}
	if errors.Is(err, contentapp.ErrInvalidLocale) {
		return httperr.BadRequest(c, "invalid_locale", "locale must be one of: en, ru, kk, es")
	}
	h.warn("public lookup failed", zap.String("slug", slug), zap.String("locale", locale), zap.Error(err))
	return httperr.Internal(c)
}

func (h *Handler) writePublicHeaders(c *fiber.Ctx, _slug, effective, etag, source string, updatedAt time.Time) {
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

type adminMatrixResponse struct {
	Entries []contentapp.MatrixEntryView `json:"entries"`
}

func (h *Handler) handleAdminMatrix(c *fiber.Ctx) error {
	if !h.guardSuperAdmin(c, "GET", "/v1/admin/content") {
		return nil
	}
	entries, err := h.app.ListAdminMatrix(c.Context())
	if err != nil {
		if errors.Is(err, domain.ErrUnavailable) {
			return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
		}
		h.warn("admin matrix list failed", zap.Error(err))
		return httperr.Internal(c)
	}
	return c.JSON(adminMatrixResponse{Entries: entries})
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

func (h *Handler) handleAdminGet(c *fiber.Ctx) error {
	slug := c.Params("slug")
	locale := c.Params("locale")
	if !domain.IsKnownSlug(slug) {
		return httperr.NotFound(c, "unknown_slug", "unknown content slug: "+slug)
	}
	if !domain.IsSupportedLocale(locale) {
		return httperr.BadRequest(c, "invalid_locale", "locale must be one of: en, ru, kk, es")
	}
	if !h.guardSuperAdmin(c, "GET", "/v1/admin/content/"+slug+"/"+locale) {
		return nil
	}
	view, err := h.app.GetAdminBlock(c.Context(), slug, locale)
	if err != nil {
		if errors.Is(err, domain.ErrUnavailable) {
			return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
		}
		h.warn("admin get failed", zap.String("slug", slug), zap.String("locale", locale), zap.Error(err))
		return httperr.Internal(c)
	}
	resp := adminBlockResponse{
		Slug: view.Slug, Locale: view.Locale, SchemaVersion: view.SchemaVersion, PublishedAt: view.PublishedAt,
		Content: view.Content, DraftContent: view.DraftContent, UpdatedAt: view.UpdatedAt, UpdatedBy: view.UpdatedBy,
		ETag: view.ETag,
	}
	if view.ETag != "" {
		c.Set("ETag", `"`+view.ETag+`"`)
	}
	return c.JSON(resp)
}

type adminUpsertReq struct {
	Content json.RawMessage `json:"content"`
	Publish bool            `json:"publish"`
}

func (h *Handler) handleAdminUpsert(c *fiber.Ctx) error {
	slug := c.Params("slug")
	locale := c.Params("locale")
	if !domain.IsKnownSlug(slug) {
		return httperr.NotFound(c, "unknown_slug", "unknown content slug: "+slug)
	}
	if !domain.IsSupportedLocale(locale) {
		return httperr.BadRequest(c, "invalid_locale", "locale must be one of: en, ru, kk, es")
	}
	if !h.guardSuperAdmin(c, "PUT", "/v1/admin/content/"+slug+"/"+locale) {
		return nil
	}
	actorID, actorEmail := h.actorInfo(c)
	var req adminUpsertReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_json", "invalid JSON body")
	}
	var priorTS *time.Time
	if rawIM := c.Get("If-Match"); rawIM != "" {
		if ts, ok := contentapp.ParseIfMatchTimestamp(rawIM); ok {
			priorTS = &ts
		} else if cur, err := h.app.CurrentUpdatedAt(c.Context(), slug, locale); err == nil {
			return h.precondFailedResp(c, cur)
		}
	}
	res, err := h.app.UpsertAdmin(c.Context(), contentapp.UpsertCommand{
		Slug: slug, Locale: locale, Content: req.Content, Publish: req.Publish,
		ActorID: actorID, ActorEmail: actorEmail, PriorUpdatedAt: priorTS,
	})
	if err != nil {
		return h.mapAdminUpsertErr(c, err)
	}
	return c.JSON(adminBlockResponse{
		Slug: res.Slug, Locale: res.Locale, SchemaVersion: res.SchemaVersion, PublishedAt: res.PublishedAt,
		Content: res.Content, DraftContent: res.DraftContent, UpdatedAt: res.UpdatedAt, UpdatedBy: res.UpdatedBy,
		ETag: res.ETag,
	})
}

func (h *Handler) mapAdminUpsertErr(c *fiber.Ctx, err error) error {
	var sv *contentapp.SchemaViolation
	if errors.As(err, &sv) {
		pd := httperr.ProblemDetails{
			Status: fiber.StatusUnprocessableEntity, Code: sv.Code, Detail: sv.Detail,
			Extensions: map[string]any{"field": sv.Field, "schema_path": sv.SchemaPath},
		}
		return httperr.Send(c, pd)
	}
	if errors.Is(err, contentapp.ErrContentTooLarge) {
		return httperr.TooLarge(c, "too_large", "content exceeds 256 KB")
	}
	var pe *domain.ErrPrecondition
	if errors.As(err, &pe) {
		return h.precondFailedResp(c, pe.CurrentUpdatedAt)
	}
	if errors.Is(err, domain.ErrUnavailable) {
		return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
	}
	h.warn("admin upsert failed", zap.Error(err))
	return httperr.Internal(c)
}

func (h *Handler) handleAdminDelete(c *fiber.Ctx) error {
	slug := c.Params("slug")
	locale := c.Params("locale")
	if !domain.IsKnownSlug(slug) {
		return httperr.NotFound(c, "unknown_slug", "unknown content slug: "+slug)
	}
	if !domain.IsSupportedLocale(locale) {
		return httperr.BadRequest(c, "invalid_locale", "locale must be one of: en, ru, kk, es")
	}
	if !h.guardSuperAdmin(c, "DELETE", "/v1/admin/content/"+slug+"/"+locale) {
		return nil
	}
	actorID, actorEmail := h.actorInfo(c)
	_, err := h.app.DeleteAdmin(c.Context(), slug, locale, actorID, actorEmail)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return httperr.NotFound(c, "row_absent", "no row exists for slug and locale")
		}
		if errors.Is(err, domain.ErrUnavailable) {
			return httperr.Unavailable(c, "content_store_unavailable", "content service unavailable")
		}
		h.warn("admin delete failed", zap.String("slug", slug), zap.String("locale", locale), zap.Error(err))
		return httperr.Internal(c)
	}
	return c.JSON(fiber.Map{"reverted": true})
}

func (h *Handler) guardSuperAdmin(c *fiber.Ctx, method, path string) bool {
	uid, ok := h.parseActor(c)
	if !ok {
		_ = httperr.Unauthorized(c, "auth_required", "authentication required")
		return false
	}
	if !h.app.CheckSuperAdmin(c.Context(), uid) {
		h.app.LogAccessDenied(c.Context(), uid, method, path)
		_ = httperr.Forbidden(c, "super_admin_required", "super_admin only")
		return false
	}
	return true
}

func (h *Handler) previewActor(c *fiber.Ctx) (uuid.UUID, bool) {
	uid, ok := h.parseActor(c)
	if !ok {
		return uuid.Nil, false
	}
	if !h.app.CheckSuperAdmin(c.Context(), uid) {
		return uuid.Nil, false
	}
	return uid, true
}

func (h *Handler) parseActor(c *fiber.Ctx) (uuid.UUID, bool) {
	if claims := auth.ClaimsFromCtx(c); claims != nil {
		uid, err := uuid.Parse(claims.UserID)
		if err == nil {
			return uid, true
		}
	}
	authz := c.Get("Authorization")
	if len(authz) < 8 || !strings.EqualFold(authz[:7], "Bearer ") {
		return uuid.Nil, false
	}
	claims, err := auth.ParseJWT(h.jwtSecret, strings.TrimSpace(authz[7:]))
	if err != nil {
		return uuid.Nil, false
	}
	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, false
	}
	return uid, true
}

func (h *Handler) actorInfo(c *fiber.Ctx) (uuid.UUID, string) {
	if claims := auth.ClaimsFromCtx(c); claims != nil {
		uid, _ := uuid.Parse(claims.UserID)
		return uid, claims.Email
	}
	uid, _ := h.parseActor(c)
	return uid, ""
}

func (h *Handler) precondFailedResp(c *fiber.Ctx, cur time.Time) error {
	return httperr.Send(c, httperr.ProblemDetails{
		Status: fiber.StatusPreconditionFailed, Code: "precondition_failed",
		Detail: "content was updated by another admin; reload to see latest",
		Extensions: map[string]any{"current_etag": cur.UTC().Format(time.RFC3339Nano)},
	})
}

func (h *Handler) warn(msg string, fields ...zap.Field) {
	if h.logger == nil {
		return
	}
	h.logger.Warn(msg, fields...)
}
