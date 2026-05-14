// admin_phase3.go — super_admin endpoints для Phase 3 admin workstream.
//
// Покрывает:
//   * /v1/admin/email-templates       — matrix + read/upsert/delete (override
//                                       layer над embedded baseline)
//   * /v1/admin/teams/:id/flags       — JSONB feature flags per-team
//   * /v1/admin/teams/:id/plan-limits — JSONB numeric overrides per-team
//   * /v1/admin/webhooks              — cross-team webhooks view (read-only)
//   * /v1/admin/api-tokens            — cross-user tokens view (prefix only)
//
// Все mutation endpoints пишут audit log через s.Audit (action names per
// .team/product-audit-taxonomy.md). Все защищены requireSuperAdmin.

package teams

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/mailer"
	"github.com/eye-of-providence/backend/internal/plans"
)

// --- Email templates ---

// emailTemplateMatrixEntry — один cell матрицы (key × locale).
type emailTemplateMatrixEntry struct {
	Key         string     `json:"key"`
	Locale      string     `json:"locale"`
	HasOverride bool       `json:"has_override"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty"`
}

func (s Service) handleAdminListEmailTemplates(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	overrides := map[string]mailer.Template{}
	if s.TemplateStore != nil {
		rows, err := s.TemplateStore.ListOverrides(c.Context())
		if err != nil {
			return s.internalErr(c, err)
		}
		for _, t := range rows {
			overrides[t.Key+":"+t.Locale] = t
		}
	}
	matrix := []emailTemplateMatrixEntry{}
	for _, key := range mailer.SupportedTemplateKeys {
		for _, loc := range mailer.SupportedLocales {
			entry := emailTemplateMatrixEntry{Key: key, Locale: loc}
			if row, ok := overrides[key+":"+loc]; ok {
				entry.HasOverride = true
				ts := row.UpdatedAt
				entry.UpdatedAt = &ts
				entry.UpdatedBy = row.UpdatedBy
			}
			matrix = append(matrix, entry)
		}
	}
	return c.JSON(fiber.Map{"entries": matrix})
}

// emailTemplateView — payload для GET single template.
type emailTemplateView struct {
	Key       string     `json:"key"`
	Locale    string     `json:"locale"`
	Subject   string     `json:"subject"`
	BodyHTML  string     `json:"body_html"`
	BodyText  string     `json:"body_text"`
	IsDefault bool       `json:"is_default"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

func (s Service) handleAdminGetEmailTemplate(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	key, locale, err := parseTemplatePath(c)
	if err != nil {
		return wrapParseTemplatePathError(c, err)
	}
	view := emailTemplateView{Key: key, Locale: string(locale)}
	if s.TemplateStore != nil {
		row, err := s.TemplateStore.Lookup(c.Context(), key, locale)
		if err != nil {
			return s.internalErr(c, err)
		}
		if row != nil {
			view.Subject = row.Subject
			view.BodyHTML = row.BodyHTML
			view.BodyText = row.BodyText
			view.IsDefault = false
			ts := row.UpdatedAt
			view.UpdatedAt = &ts
			view.UpdatedBy = row.UpdatedBy
			return c.JSON(view)
		}
	}
	// Fallback: embedded baseline.
	baseline := mailer.BaselineTemplate(key, locale)
	if baseline == nil {
		return httperr.NotFound(c, "template_not_found", "no baseline for "+key+":"+string(locale))
	}
	view.Subject = baseline.Subject
	view.BodyHTML = baseline.BodyHTML
	view.BodyText = baseline.BodyText
	view.IsDefault = true
	return c.JSON(view)
}

type adminUpsertTemplateReq struct {
	Subject  string `json:"subject"`
	BodyHTML string `json:"body_html"`
	BodyText string `json:"body_text"`
}

func (s Service) handleAdminUpsertEmailTemplate(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	if s.TemplateStore == nil {
		return httperr.Unavailable(c, "template_store_unavailable", "template store not configured")
	}
	key, locale, err := parseTemplatePath(c)
	if err != nil {
		return wrapParseTemplatePathError(c, err)
	}
	var req adminUpsertTemplateReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	if req.Subject == "" || req.BodyHTML == "" {
		return httperr.BadRequest(c, "missing_field", "subject and body_html are required")
	}
	// Size cap (256 KB), per admin-email-templates.md edge case.
	if len(req.BodyHTML) > 256*1024 || len(req.BodyText) > 256*1024 {
		return httperr.TooLarge(c, "body_too_large", "body exceeds 256 KB")
	}
	// Sanity-check render against empty vars to surface broken template
	// syntax at save time, not send time.
	if _, err := mailer.Render(mailer.Template{
		Key:      key,
		Locale:   string(locale),
		Subject:  req.Subject,
		BodyHTML: req.BodyHTML,
		BodyText: req.BodyText,
	}, nil); err != nil {
		actorID, actorEmail := s.actorInfo(c)
		s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionEmailTemplateUpdateRejected,
			"email_template", key+":"+string(locale), map[string]any{
				"error_code":   "invalid_template_syntax",
				"error_detail": err.Error(),
			})
		return httperr.BadRequest(c, "invalid_template_syntax", err.Error())
	}

	actorID, actorEmail := s.actorInfo(c)
	out, err := s.TemplateStore.Upsert(c.Context(), mailer.Template{
		Key:      key,
		Locale:   string(locale),
		Subject:  req.Subject,
		BodyHTML: req.BodyHTML,
		BodyText: req.BodyText,
	}, actorID)
	if err != nil {
		switch {
		case errors.Is(err, mailer.ErrInvalidEmailTemplateKey):
			return httperr.BadRequest(c, "invalid_template_key",
				"key must be one of "+joinSupported(mailer.SupportedTemplateKeys))
		case errors.Is(err, mailer.ErrInvalidEmailTemplateLocale):
			return httperr.BadRequest(c, "invalid_template_locale",
				"locale must be one of "+joinSupported(mailer.SupportedLocales))
		default:
			return s.internalErr(c, err)
		}
	}
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionEmailTemplateUpdated,
		"email_template", key+":"+string(locale), map[string]any{
			"subject":         req.Subject,
			"body_html_bytes": len(req.BodyHTML),
			"body_text_bytes": len(req.BodyText),
		})
	return c.JSON(fiber.Map{
		"key":        out.Key,
		"locale":     out.Locale,
		"subject":    out.Subject,
		"body_html":  out.BodyHTML,
		"body_text":  out.BodyText,
		"updated_at": out.UpdatedAt,
		"updated_by": out.UpdatedBy,
		"is_default": false,
	})
}

func (s Service) handleAdminDeleteEmailTemplate(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	if s.TemplateStore == nil {
		return httperr.Unavailable(c, "template_store_unavailable", "template store not configured")
	}
	key, locale, err := parseTemplatePath(c)
	if err != nil {
		return wrapParseTemplatePathError(c, err)
	}
	prev, derr := s.TemplateStore.Delete(c.Context(), key, string(locale))
	if derr != nil {
		return s.internalErr(c, derr)
	}
	if prev == nil {
		return httperr.NotFound(c, "template_not_overridden",
			"no override exists for "+key+":"+string(locale))
	}
	actorID, actorEmail := s.actorInfo(c)
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionEmailTemplateReverted,
		"email_template", key+":"+string(locale), map[string]any{
			"previous_subject":    prev.Subject,
			"previous_body_html":  prev.BodyHTML,
			"previous_body_text":  prev.BodyText,
			"previous_updated_at": prev.UpdatedAt,
			"previous_updated_by": prev.UpdatedBy,
		})
	return c.Status(http.StatusNoContent).Send(nil)
}

// Sentinel errors for parseTemplatePath — httperr.Send returns nil after writing
// the response, so returning that value as `error` would let the caller fall
// through with empty key/locale (DB CHECK failures, wrong status codes).
var (
	errInvalidAdminTemplateKey    = errors.New("invalid admin template key")
	errInvalidAdminTemplateLocale = errors.New("invalid admin template locale")
)

// parseTemplatePath — общий валидатор для :key/:locale path params.
// On failure returns a sentinel error; caller must map via wrapParseTemplatePathError.
func parseTemplatePath(c *fiber.Ctx) (string, mailer.Locale, error) {
	key := c.Params("key")
	locale := c.Params("locale")
	if !mailer.IsSupportedTemplateKey(key) {
		return "", "", errInvalidAdminTemplateKey
	}
	if !mailer.IsSupportedLocale(locale) {
		return "", "", errInvalidAdminTemplateLocale
	}
	return key, mailer.Locale(locale), nil
}

func wrapParseTemplatePathError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, errInvalidAdminTemplateKey):
		return httperr.BadRequest(c, "invalid_template_key",
			"key must be one of "+joinSupported(mailer.SupportedTemplateKeys))
	case errors.Is(err, errInvalidAdminTemplateLocale):
		return httperr.BadRequest(c, "invalid_template_locale",
			"locale must be one of "+joinSupported(mailer.SupportedLocales))
	default:
		return err
	}
}

func joinSupported(items []string) string {
	out := ""
	for i, s := range items {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}

// --- Team flags ---

func (s Service) handleAdminGetTeamFlags(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	flags, ferr := s.readTeamFlags(c, teamID)
	if ferr != nil {
		return ferr
	}
	return c.JSON(fiber.Map{
		"team_id": teamID,
		"flags":   flags,
	})
}

type adminFlagsPatchReq struct {
	Flags map[string]any `json:"flags"`
}

func (s Service) handleAdminPatchTeamFlags(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	var req adminFlagsPatchReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	if req.Flags == nil {
		return httperr.BadRequest(c, "missing_flags", "flags object required")
	}
	// Strict allowlist validation BEFORE any DB mutation (all-or-nothing).
	patch, verr := plans.ValidateFlags(req.Flags)
	if verr != nil {
		actorID, actorEmail := s.actorInfo(c)
		s.logFlagRejection(c, actorID, actorEmail, teamID, verr)
		return mapFlagError(c, verr)
	}

	existing, ferr := s.readTeamFlags(c, teamID)
	if ferr != nil {
		return ferr
	}
	merged := plans.MergeFlags(existing, patch)
	merged = pruneNullsMap(merged)
	mergedBytes, err := json.Marshal(merged)
	if err != nil {
		return s.internalErr(c, err)
	}
	tag, err := s.Pool.Exec(c.Context(),
		"UPDATE teams SET flags = $1 WHERE id = $2", mergedBytes, teamID)
	if err != nil {
		return s.internalErr(c, err)
	}
	if tag.RowsAffected() == 0 {
		return httperr.NotFound(c, "team_not_found", "team not found")
	}

	diff := plans.FlagsDiff(existing, merged)
	actorID, actorEmail := s.actorInfo(c)
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionTeamFlagsUpdated,
		"team", teamID.String(), map[string]any{
			"diff": diff,
		})
	return c.JSON(fiber.Map{
		"team_id": teamID,
		"flags":   merged,
	})
}

func (s Service) logFlagRejection(c *fiber.Ctx, actorID uuid.UUID, actorEmail string, teamID uuid.UUID, err error) {
	meta := map[string]any{}
	var fe *plans.FlagError
	if errors.As(err, &fe) {
		meta["error_code"] = fe.Code
		meta["field"] = fe.Field
	} else {
		meta["error_code"] = "validation_failed"
		meta["error_detail"] = err.Error()
	}
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionTeamFlagsUpdateRejected,
		"team", teamID.String(), meta)
}

// readTeamFlags — JSONB → map. Empty object на NULL / missing.
func (s Service) readTeamFlags(c *fiber.Ctx, teamID uuid.UUID) (map[string]any, error) {
	var raw []byte
	err := s.Pool.QueryRow(c.Context(),
		"SELECT flags FROM teams WHERE id = $1", teamID).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httperr.NotFound(c, "team_not_found", "team not found")
	}
	if err != nil {
		return nil, s.internalErr(c, err)
	}
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}, nil
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

// pruneNullsMap — JSONB-friendly clean-up. nil-значения уже удалены через
// MergeFlags, но если DB вернула row с {"x": null} (явная клиентская
// очистка), нормализуем чтобы Diff не ловил artifact'ы.
func pruneNullsMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		if v == nil {
			continue
		}
		out[k] = v
	}
	return out
}

// mapFlagError — переводит plans.FlagError → httperr response с правильным
// status code (400 для unknown/type, 400 для range; QA scenario может
// ожидать 422 для range — оставляем 400 как существующий контракт).
func mapFlagError(c *fiber.Ctx, err error) error {
	var fe *plans.FlagError
	if !errors.As(err, &fe) {
		return httperr.BadRequest(c, "invalid_flags", err.Error())
	}
	pd := httperr.ProblemDetails{
		Status: fiber.StatusBadRequest,
		Code:   fe.Code,
		Detail: fe.Detail,
		Extensions: map[string]any{
			"field": fe.Field,
		},
	}
	if fe.Expected != "" {
		pd.Extensions["expected"] = fe.Expected
		pd.Extensions["got"] = fe.Got
	}
	if fe.Min != 0 {
		pd.Extensions["minimum"] = fe.Min
	}
	if fe.Max != 0 {
		pd.Extensions["maximum"] = fe.Max
	}
	return httperr.Send(c, pd)
}

// --- Plan-limit overrides ---

func (s Service) handleAdminGetTeamPlanLimits(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	override, plan, ferr := s.readTeamOverrides(c, teamID)
	if ferr != nil {
		return ferr
	}
	defaults := s.Plans.Limits(plan)
	return c.JSON(fiber.Map{
		"team_id":            teamID,
		"plan":               plan,
		"overrides":          override,
		"effective_defaults": planLimitsAsMap(defaults),
	})
}

// adminPlanLimitsPatchReq shape (documented в коде, не используется как
// struct — ниже парсится вручную через raw JSON, чтобы различить
// `{}` (отсутствует) от `{"limits": null}` (full reset)):
//
//	{ "limits": { "<key>": <int>|null } }   — partial patch / single key reset
//	{ "limits": null }                       — full reset to plan defaults

func (s Service) handleAdminPatchTeamPlanLimits(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	// Парсим вручную через raw JSON чтобы отличить
	// `{}` (отсутствует) от `{"limits": null}` (reset all).
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &raw); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	limitsRaw, present := raw["limits"]
	if !present {
		return httperr.BadRequest(c, "missing_limits", "limits field required (use null to reset)")
	}

	actorID, actorEmail := s.actorInfo(c)

	if string(limitsRaw) == "null" {
		// Full reset — drop column.
		existing, _, ferr := s.readTeamOverrides(c, teamID)
		if ferr != nil {
			return ferr
		}
		if _, err := s.Pool.Exec(c.Context(),
			"UPDATE teams SET plan_limits_override = NULL WHERE id = $1", teamID); err != nil {
			return s.internalErr(c, err)
		}
		s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionTeamPlanOverridesCleared,
			"team", teamID.String(), map[string]any{
				"cleared_keys": keysOf(existing),
			})
		return c.JSON(fiber.Map{
			"team_id":   teamID,
			"overrides": nil,
		})
	}

	var patch map[string]any
	if err := json.Unmarshal(limitsRaw, &patch); err != nil {
		return httperr.BadRequest(c, "invalid_limits", "limits must be object or null")
	}
	normalized, verr := plans.ValidateOverrides(patch)
	if verr != nil {
		s.logOverrideRejection(c, actorID, actorEmail, teamID, verr)
		return mapFlagError(c, verr)
	}

	existing, _, ferr := s.readTeamOverrides(c, teamID)
	if ferr != nil {
		return ferr
	}
	merged := plans.MergeOverrides(existing, normalized)
	merged = pruneNullsMap(merged)
	var newValue []byte
	if len(merged) == 0 {
		// Если после merge ничего не осталось — store NULL (= clear).
		if _, err := s.Pool.Exec(c.Context(),
			"UPDATE teams SET plan_limits_override = NULL WHERE id = $1", teamID); err != nil {
			return s.internalErr(c, err)
		}
	} else {
		bts, err := json.Marshal(merged)
		if err != nil {
			return s.internalErr(c, err)
		}
		newValue = bts
		if _, err := s.Pool.Exec(c.Context(),
			"UPDATE teams SET plan_limits_override = $1 WHERE id = $2", bts, teamID); err != nil {
			return s.internalErr(c, err)
		}
	}

	diff := plans.FlagsDiff(existing, merged) // та же diff-shape применима
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionTeamPlanOverridesUpdated,
		"team", teamID.String(), map[string]any{
			"diff": diff,
		})
	_ = newValue // оставляем переменную для clarity (раньше использовалась в return body)
	return c.JSON(fiber.Map{
		"team_id":   teamID,
		"overrides": merged,
	})
}

func (s Service) logOverrideRejection(c *fiber.Ctx, actorID uuid.UUID, actorEmail string, teamID uuid.UUID, err error) {
	meta := map[string]any{}
	var fe *plans.FlagError
	if errors.As(err, &fe) {
		meta["error_code"] = fe.Code
		meta["field"] = fe.Field
	} else {
		meta["error_code"] = "validation_failed"
		meta["error_detail"] = err.Error()
	}
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionTeamPlanOverridesRejected,
		"team", teamID.String(), meta)
}

// readTeamOverrides — joint read для plan + overrides.
func (s Service) readTeamOverrides(c *fiber.Ctx, teamID uuid.UUID) (map[string]any, string, error) {
	var raw []byte
	var plan string
	err := s.Pool.QueryRow(c.Context(),
		"SELECT plan_limits_override, COALESCE(subscription_plan, 'free') FROM teams WHERE id = $1",
		teamID).Scan(&raw, &plan)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", httperr.NotFound(c, "team_not_found", "team not found")
	}
	if err != nil {
		return nil, "", s.internalErr(c, err)
	}
	out := map[string]any{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, plan, nil
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func planLimitsAsMap(l plans.Limits) map[string]any {
	return map[string]any{
		"plan":               l.Plan,
		"max_users_per_team": l.MaxUsersPerTeam,
		"max_webhooks":       l.MaxWebhooks,
		"retention_days":     l.RetentionDays,
		"sso":                l.SSO,
		"audit_log":          l.AuditLog,
		"custom_roles":       l.CustomRoles,
		"webhook_signing":    l.WebhookSigning,
	}
}

// --- Cross-team webhooks list ---

type adminWebhookRow struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	UserEmail      string     `json:"user_email"`
	URL            string     `json:"url"`
	Events         []string   `json:"events"`
	Format         string     `json:"format"`
	Active         bool       `json:"active"`
	LastDeliveryAt *time.Time `json:"last_delivery_at,omitempty"`
	LastStatus     *int       `json:"last_status,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (s Service) handleAdminListAllWebhooks(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	limit, offset := adminListPagination(c)
	if limit > 100 {
		limit = 100
	}
	rows, err := s.Pool.Query(c.Context(), `
		SELECT w.id, w.user_id, COALESCE(u.email, '') AS user_email, w.url, w.events,
		       w.format, w.active, w.last_delivery_at, w.last_status, w.created_at
		FROM webhooks w
		LEFT JOIN users u ON u.id = w.user_id
		ORDER BY w.created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	out := []adminWebhookRow{}
	for rows.Next() {
		var r adminWebhookRow
		if err := rows.Scan(&r.ID, &r.UserID, &r.UserEmail, &r.URL, &r.Events,
			&r.Format, &r.Active, &r.LastDeliveryAt, &r.LastStatus, &r.CreatedAt); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, r)
	}
	return c.JSON(fiber.Map{
		"webhooks": out,
		"limit":    limit,
		"offset":   offset,
	})
}

// --- Cross-user API tokens list ---

type adminAPITokenRow struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	UserEmail  string     `json:"user_email"`
	Name       string     `json:"name"`
	Scope      string     `json:"scope"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

func (s Service) handleAdminListAllAPITokens(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	limit, offset := adminListPagination(c)
	if limit > 100 {
		limit = 100
	}
	// include_revoked=1 для отображения всех (default — только active).
	includeRevoked := c.Query("include_revoked") == "1"
	q := `
		SELECT t.id, t.user_id, COALESCE(u.email, '') AS user_email, t.name, t.scope,
		       t.prefix, t.created_at, t.expires_at, t.last_used_at, t.revoked_at
		FROM api_tokens t
		LEFT JOIN users u ON u.id = t.user_id`
	if !includeRevoked {
		q += " WHERE t.revoked_at IS NULL"
	}
	q += " ORDER BY t.created_at DESC LIMIT $1 OFFSET $2"
	rows, err := s.Pool.Query(c.Context(), q, limit, offset)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	out := []adminAPITokenRow{}
	for rows.Next() {
		var r adminAPITokenRow
		if err := rows.Scan(&r.ID, &r.UserID, &r.UserEmail, &r.Name, &r.Scope,
			&r.Prefix, &r.CreatedAt, &r.ExpiresAt, &r.LastUsedAt, &r.RevokedAt); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, r)
	}
	return c.JSON(fiber.Map{
		"tokens": out,
		"limit":  limit,
		"offset": offset,
	})
}
