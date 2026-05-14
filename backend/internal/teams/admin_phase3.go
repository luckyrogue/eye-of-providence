package teams

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/mailer"
	"github.com/eye-of-providence/backend/internal/plans"
	"github.com/eye-of-providence/backend/internal/teams/adminlists"
	"github.com/eye-of-providence/backend/internal/teams/emailtemplates"
	"github.com/eye-of-providence/backend/internal/teams/teamflags"
	"github.com/eye-of-providence/backend/internal/teams/teamplanlimits"
)

func (s Service) handleAdminListEmailTemplates(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	app := s.newEmailTemplatesService()
	entries, err := app.ListMatrix(c.Context())
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"entries": entries})
}

func (s Service) handleAdminGetEmailTemplate(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	key := c.Params("key")
	locale := c.Params("locale")
	app := s.newEmailTemplatesService()
	view, err := app.Get(c.Context(), key, locale)
	if err != nil {
		return s.mapEmailTemplateHTTPError(c, err)
	}
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
	key := c.Params("key")
	locale := c.Params("locale")
	var req adminUpsertTemplateReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	app := s.newEmailTemplatesService()
	actorID, actorEmail := s.actorInfo(c)
	meta := emailtemplates.RequestMeta{
		IP:        audit.ClientIP(c),
		UserAgent: c.Get("User-Agent"),
	}
	out, err := app.Upsert(c.Context(), meta, actorID, actorEmail, key, locale, emailtemplates.UpsertCommand{
		Subject:  req.Subject,
		BodyHTML: req.BodyHTML,
		BodyText: req.BodyText,
	})
	if err != nil {
		if detail, ok := emailtemplates.IsInvalidSyntax(err); ok {
			return httperr.BadRequest(c, "invalid_template_syntax", detail)
		}
		return s.mapEmailTemplateHTTPError(c, err)
	}
	return c.JSON(out)
}

func (s Service) handleAdminDeleteEmailTemplate(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	key := c.Params("key")
	locale := c.Params("locale")
	app := s.newEmailTemplatesService()
	actorID, actorEmail := s.actorInfo(c)
	meta := emailtemplates.RequestMeta{
		IP:        audit.ClientIP(c),
		UserAgent: c.Get("User-Agent"),
	}
	_, err := app.Delete(c.Context(), meta, actorID, actorEmail, key, locale)
	if err != nil {
		return s.mapEmailTemplateHTTPError(c, err)
	}
	return c.Status(http.StatusNoContent).Send(nil)
}

func (s Service) mapEmailTemplateHTTPError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, emailtemplates.ErrInvalidKey):
		return httperr.BadRequest(c, "invalid_template_key",
			"key must be one of "+emailtemplates.JoinSupported(mailer.SupportedTemplateKeys))
	case errors.Is(err, emailtemplates.ErrInvalidLocale):
		return httperr.BadRequest(c, "invalid_template_locale",
			"locale must be one of "+emailtemplates.JoinSupported(mailer.SupportedLocales))
	case errors.Is(err, emailtemplates.ErrMissingField):
		return httperr.BadRequest(c, "missing_field", "subject and body_html are required")
	case errors.Is(err, emailtemplates.ErrBodyTooLarge):
		return httperr.TooLarge(c, "body_too_large", "body exceeds 256 KB")
	case errors.Is(err, emailtemplates.ErrStoreUnavailable):
		return httperr.Unavailable(c, "template_store_unavailable", "template store not configured")
	case errors.Is(err, emailtemplates.ErrNoBaseline):
		return httperr.NotFound(c, "template_not_found", err.Error())
	case errors.Is(err, emailtemplates.ErrNoOverride):
		return httperr.NotFound(c, "template_not_overridden", err.Error())
	default:
		return s.internalErr(c, err)
	}
}

func (s Service) handleAdminGetTeamFlags(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	app := s.newTeamFlagsService()
	flags, err := app.Get(c.Context(), teamID)
	if err != nil {
		if errors.Is(err, teamflags.ErrTeamNotFound) {
			return httperr.NotFound(c, "team_not_found", "team not found")
		}
		return s.internalErr(c, err)
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
	app := s.newTeamFlagsService()
	actorID, actorEmail := s.actorInfo(c)
	meta := teamflags.RequestMeta{
		IP:        audit.ClientIP(c),
		UserAgent: c.Get("User-Agent"),
	}
	merged, err := app.Patch(c.Context(), meta, actorID, actorEmail, teamID, req.Flags)
	if err != nil {
		if errors.Is(err, teamflags.ErrMissingFlags) {
			return httperr.BadRequest(c, "missing_flags", "flags object required")
		}
		var fe *plans.FlagError
		if errors.As(err, &fe) {
			return mapFlagError(c, err)
		}
		if errors.Is(err, teamflags.ErrTeamNotFound) {
			return httperr.NotFound(c, "team_not_found", "team not found")
		}
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{
		"team_id": teamID,
		"flags":   merged,
	})
}

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

func (s Service) handleAdminGetTeamPlanLimits(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	app := s.newTeamPlanLimitsService()
	view, err := app.Get(c.Context(), teamID)
	if err != nil {
		if errors.Is(err, teamplanlimits.ErrTeamNotFound) {
			return httperr.NotFound(c, "team_not_found", "team not found")
		}
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{
		"team_id":            view.TeamID,
		"plan":               view.Plan,
		"overrides":          view.Overrides,
		"effective_defaults": view.EffectiveDefaults,
	})
}

func (s Service) handleAdminPatchTeamPlanLimits(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &raw); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	limitsRaw, present := raw["limits"]
	if !present {
		return httperr.BadRequest(c, "missing_limits", "limits field required (use null to reset)")
	}
	actorID, actorEmail := s.actorInfo(c)
	meta := teamplanlimits.RequestMeta{
		IP:        audit.ClientIP(c),
		UserAgent: c.Get("User-Agent"),
	}
	app := s.newTeamPlanLimitsService()
	var cmd teamplanlimits.PatchLimitsCmd
	if string(limitsRaw) == "null" {
		cmd.FullReset = true
	} else {
		var patch map[string]any
		if err := json.Unmarshal(limitsRaw, &patch); err != nil {
			return httperr.BadRequest(c, "invalid_limits", "limits must be object or null")
		}
		cmd.Patch = patch
	}
	out, err := app.Patch(c.Context(), meta, actorID, actorEmail, teamID, cmd)
	if err != nil {
		var fe *plans.FlagError
		if errors.As(err, &fe) {
			return mapFlagError(c, err)
		}
		if errors.Is(err, teamplanlimits.ErrTeamNotFound) {
			return httperr.NotFound(c, "team_not_found", "team not found")
		}
		return s.internalErr(c, err)
	}
	if out.FullReset {
		return c.JSON(fiber.Map{
			"team_id":   teamID,
			"overrides": nil,
		})
	}
	return c.JSON(fiber.Map{
		"team_id":   teamID,
		"overrides": out.Overrides,
	})
}

func (s Service) handleAdminListAllWebhooks(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	limit, offset := adminListPagination(c)
	app := adminlists.New(adminlists.NewPGListQuerier(s.Pool))
	rows, err := app.ListWebhooks(c.Context(), limit, offset)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{
		"webhooks": rows,
		"limit":    limit,
		"offset":   offset,
	})
}

func (s Service) handleAdminListAllAPITokens(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	limit, offset := adminListPagination(c)
	includeRevoked := c.Query("include_revoked") == "1"
	app := adminlists.New(adminlists.NewPGListQuerier(s.Pool))
	rows, err := app.ListAPITokens(c.Context(), limit, offset, includeRevoked)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{
		"tokens": rows,
		"limit":  limit,
		"offset": offset,
	})
}
