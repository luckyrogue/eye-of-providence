package teams

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/teams/adminapp"
)

func (s Service) handleAdminListAllTeams(c *fiber.Ctx) error {
	if !s.isSuperAdmin(c) {
		return httperr.Forbidden(c, "super_admin_required", "super_admin only")
	}
	limit, offset := adminListPagination(c)
	out, err := s.adminApp().ListTeams(c.Context(), limit, offset)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"teams": out, "limit": limit, "offset": offset})
}

func (s Service) handleAdminListAllUsers(c *fiber.Ctx) error {
	if !s.isSuperAdmin(c) {
		return httperr.Forbidden(c, "super_admin_required", "super_admin only")
	}
	limit, offset := adminListPagination(c)
	out, err := s.adminApp().ListUsers(c.Context(), limit, offset)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"users": out, "limit": limit, "offset": offset})
}

func (s Service) handleAdminStats(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	st, err := s.adminApp().Stats(c.Context())
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{
		"users_total":   st.UsersTotal,
		"teams_total":   st.TeamsTotal,
		"members_total": st.MembersTotal,
		"beta_limit":    s.BetaTeamLimit,
	})
}

func (s Service) handleAdminAudit(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	f := audit.ListFilter{
		Action:     c.Query("action"),
		TargetType: c.Query("target_type"),
		TargetID:   c.Query("target_id"),
	}
	if a := c.Query("actor_id"); a != "" {
		id, err := uuid.Parse(a)
		if err != nil {
			return httperr.BadRequest(c, "invalid_actor_id", "actor_id must be uuid")
		}
		f.ActorID = id
	}
	if l := c.Query("limit"); l != "" {
		n, err := strconv.Atoi(l)
		if err == nil && n > 0 {
			f.Limit = n
		}
	}
	if o := c.Query("offset"); o != "" {
		n, err := strconv.Atoi(o)
		if err == nil && n > 0 {
			f.Offset = n
		}
	}
	entries, err := s.adminApp().ListAudit(c.Context(), f)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{
		"entries": entries,
		"limit":   f.Limit,
		"offset":  f.Offset,
	})
}

func (s Service) handleAdminRevenue(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	rep, err := s.adminApp().Revenue(c.Context())
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{
		"total_cents":    rep.TotalCents,
		"last_30d_cents": rep.Last30dCents,
		"currency":       rep.Currency,
		"paying_teams":   rep.PayingTeams,
		"by_plan":        rep.ByPlan,
		"recent":         rep.Recent,
	})
}

func (s Service) handleAdminSSOList(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	out, err := s.adminApp().ListSSOConfigs(c.Context())
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"configs": out})
}

func (s Service) handleAdminSSODisable(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	if err := s.adminApp().DisableSSO(c.Context(), teamID); err != nil {
		if errors.Is(err, adminapp.ErrSSONotConfigured) {
			return httperr.NotFound(c, "sso_not_configured", "no SSO config for team")
		}
		return s.internalErr(c, err)
	}
	actorID, actorEmail := s.actorInfo(c)
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionSSOSaved, "team", teamID.String(), map[string]any{
		"enabled": false, "force_off": true, "by": "super_admin",
	})
	return c.JSON(fiber.Map{"ok": true})
}

func (s Service) handleAdminDeleteTeam(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	del, err := s.adminApp().DeleteTeam(c.Context(), teamID)
	if err != nil {
		return s.internalErr(c, err)
	}
	actorID, actorEmail := s.actorInfo(c)
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionTeamDeleted, "team", teamID.String(), map[string]any{
		"name": del.TeamName,
	})
	return c.JSON(fiber.Map{"ok": true})
}

func (s Service) handleAdminDeleteUser(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	uid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_user_id", "invalid user id")
	}
	del, err := s.adminApp().DeleteUser(c.Context(), uid, userID(c))
	if err != nil {
		if errors.Is(err, adminapp.ErrCannotDeleteSelf) {
			return httperr.Conflict(c, "cannot_delete_self", "cannot delete yourself")
		}
		if errors.Is(err, adminapp.ErrLastSuperAdmin) {
			return httperr.Conflict(c, "last_super_admin", "cannot delete last super_admin")
		}
		return s.internalErr(c, err)
	}
	actorID, actorEmail := s.actorInfo(c)
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionUserDeleted, "user", uid.String(), map[string]any{
		"email": del.Email, "role": del.Role,
	})
	return c.JSON(fiber.Map{"ok": true})
}

type adminUpdateUserReq struct {
	GlobalRole  *string `json:"global_role"`
	DisplayName *string `json:"display_name"`
}

func (s Service) handleAdminUpdateUser(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	uid, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_user_id", "invalid user id")
	}
	var req adminUpdateUserReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	out, err := s.adminApp().UpdateUser(c.Context(), adminapp.UpdateUserInput{
		TargetID: uid, ActorID: userID(c), GlobalRole: req.GlobalRole, DisplayName: req.DisplayName,
	})
	if err != nil {
		switch {
		case errors.Is(err, adminapp.ErrInvalidGlobalRole):
			return httperr.BadRequest(c, "invalid_global_role", "global_role must be user | super_admin")
		case errors.Is(err, adminapp.ErrCannotDemoteSelf):
			return httperr.Conflict(c, "cannot_demote_self", "cannot demote yourself")
		case errors.Is(err, adminapp.ErrInvalidDisplayName):
			return httperr.BadRequest(c, "invalid_display_name", "display_name 1..64 chars, no newlines")
		default:
			return s.internalErr(c, err)
		}
	}
	if out.RoleChanged {
		actorID, actorEmail := s.actorInfo(c)
		s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionUserRoleChanged, "user", uid.String(), map[string]any{
			"email": out.VictimEmail, "role_from": out.PrevRole, "role_to": out.NewRole,
		})
	}
	return c.JSON(fiber.Map{"ok": true})
}

type adminAddMemberReq struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (s Service) handleAdminAddMember(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	var req adminAddMemberReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	out, err := s.adminApp().AddMember(c.Context(), adminapp.AddMemberInput{
		TeamID: teamID, Email: req.Email, Role: req.Role,
	})
	if err != nil {
		switch {
		case errors.Is(err, adminapp.ErrInvalidEmail):
			return httperr.BadRequest(c, "invalid_email", "valid email required")
		case errors.Is(err, adminapp.ErrInvalidRole):
			return httperr.BadRequest(c, "invalid_role", "role must be owner | admin | member")
		case errors.Is(err, adminapp.ErrUserNotFound):
			return httperr.NotFound(c, "user_not_found", "user with this email not found")
		case errors.Is(err, adminapp.ErrOwnerLimit):
			return httperr.Conflict(c, "owner_limit", "user already owns another company")
		default:
			return s.internalErr(c, err)
		}
	}
	return c.JSON(fiber.Map{"ok": true, "user_id": out.UserID})
}

type setSubscriptionReq struct {
	Plan  *string `json:"plan"`
	Until *string `json:"until"`
	Note  *string `json:"note"`
	Payment *struct {
		AmountCents int    `json:"amount_cents"`
		Currency    string `json:"currency"`
		Method      string `json:"method"`
		Note        string `json:"note"`
		CoversUntil string `json:"covers_until"`
	} `json:"payment"`
}

func (s Service) handleSetSubscription(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	var req setSubscriptionReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	in := adminapp.SetSubscriptionInput{
		TeamID: teamID, RecordedBy: userID(c), Plan: req.Plan, Until: req.Until, Note: req.Note,
	}
	if req.Payment != nil {
		if req.Payment.AmountCents <= 0 {
			return httperr.BadRequest(c, "invalid_payment_amount", "payment.amount_cents must be > 0 if payment is set")
		}
		in.Payment = &adminapp.SubscriptionPayment{
			AmountCents: req.Payment.AmountCents,
			Currency:    req.Payment.Currency,
			Method:      req.Payment.Method,
			Note:        req.Payment.Note,
			CoversUntil: req.Payment.CoversUntil,
		}
	}
	out, err := s.adminApp().SetSubscription(c.Context(), in)
	if err != nil {
		switch {
		case errors.Is(err, adminapp.ErrInvalidPlan):
			return httperr.BadRequest(c, "invalid_plan", "plan must be free | pro | team | enterprise")
		case errors.Is(err, adminapp.ErrInvalidUntil):
			return httperr.BadRequest(c, "invalid_until", "until must be ISO8601 (RFC3339)")
		case errors.Is(err, adminapp.ErrInvalidPayment):
			return httperr.BadRequest(c, "invalid_covers_until", "payment.covers_until must be ISO8601")
		default:
			return s.internalErr(c, err)
		}
	}
	actorID, actorEmail := s.actorInfo(c)
	meta := map[string]any{}
	if out.HasPlan {
		meta["plan_to"] = out.PlanNorm
	}
	if out.HasUntil {
		if out.UntilTS != nil {
			meta["until"] = out.UntilTS.Format(time.RFC3339)
		} else {
			meta["until"] = nil
		}
	}
	if out.PaymentID != nil && out.PaymentMeta != nil {
		meta["payment_id"] = out.PaymentID.String()
		meta["amount_cents"] = out.PaymentMeta.AmountCents
		meta["currency"] = out.PaymentMeta.Currency
	}
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionSubscriptionSet, "team", teamID.String(), meta)
	return c.JSON(fiber.Map{"ok": true, "payment_id": out.PaymentID})
}

func (s Service) handleListPayments(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	out, err := s.adminApp().ListTeamPayments(c.Context(), teamID)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"payments": out})
}
