package teams

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/store"
)

func adminListPagination(c *fiber.Ctx) (limit, offset int) {
	limit = 100
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		limit = min(v, 200)
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil && v > 0 {
		offset = v
	}
	return limit, offset
}

func (s Service) requireSuperAdmin(c *fiber.Ctx) bool {
	if !s.isSuperAdmin(c) {
		_ = httperr.Forbidden(c, "super_admin_required", "super_admin only")
		return false
	}
	return true
}

func (s Service) handleAdminListAllTeams(c *fiber.Ctx) error {
	if !s.isSuperAdmin(c) {
		return httperr.Forbidden(c, "super_admin_required", "super_admin only")
	}
	limit, offset := adminListPagination(c)
	rows, err := s.Pool.Query(c.Context(), `
		SELECT t.id, t.name, t.plan, t.created_at,
		       t.subscription_plan, t.subscription_until, t.subscription_note,
		       (SELECT count(*) FROM team_members WHERE team_id = t.id) AS member_count,
		       (SELECT u.email FROM team_members tm JOIN users u ON u.id = tm.user_id
		        WHERE tm.team_id = t.id AND tm.role = 'owner' ORDER BY u.created_at LIMIT 1) AS owner_email
		FROM teams t ORDER BY t.created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	type teamRow struct {
		ID                uuid.UUID  `json:"id"`
		Name              string     `json:"name"`
		Plan              string     `json:"plan"`
		SubscriptionPlan  string     `json:"subscription_plan"`
		SubscriptionUntil *time.Time `json:"subscription_until"`
		SubscriptionNote  *string    `json:"subscription_note"`
		MemberCount       int        `json:"member_count"`
		OwnerEmail        *string    `json:"owner_email"`
		CreatedAt         time.Time  `json:"created_at"`
	}
	out := []teamRow{}
	for rows.Next() {
		var t teamRow
		if err := rows.Scan(&t.ID, &t.Name, &t.Plan, &t.CreatedAt,
			&t.SubscriptionPlan, &t.SubscriptionUntil, &t.SubscriptionNote,
			&t.MemberCount, &t.OwnerEmail); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, t)
	}
	return c.JSON(fiber.Map{"teams": out, "limit": limit, "offset": offset})
}

func (s Service) handleAdminListAllUsers(c *fiber.Ctx) error {
	if !s.isSuperAdmin(c) {
		return httperr.Forbidden(c, "super_admin_required", "super_admin only")
	}
	limit, offset := adminListPagination(c)
	rows, err := s.Pool.Query(c.Context(), `
		SELECT u.id, u.email, COALESCE(u.display_name, u.email), u.global_role, u.created_at,
		       (SELECT count(*) FROM team_members WHERE user_id = u.id) AS teams_count
		FROM users u ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	type userRow struct {
		ID          uuid.UUID `json:"id"`
		Email       string    `json:"email"`
		DisplayName string    `json:"display_name"`
		GlobalRole  string    `json:"global_role"`
		CreatedAt   time.Time `json:"created_at"`
		TeamsCount  int       `json:"teams_count"`
	}
	out := []userRow{}
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.GlobalRole, &u.CreatedAt, &u.TeamsCount); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, u)
	}
	return c.JSON(fiber.Map{"users": out, "limit": limit, "offset": offset})
}

func (s Service) handleAdminStats(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	var users, teams, members int
	_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM users").Scan(&users)
	_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM teams").Scan(&teams)
	_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM team_members").Scan(&members)
	return c.JSON(fiber.Map{
		"users_total":   users,
		"teams_total":   teams,
		"members_total": members,
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
	entries, err := s.Audit.List(c.Context(), f)
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
	ctx := c.Context()

	var totalCents, last30Cents int64
	var payingTeams int
	var currency string

	_ = s.Pool.QueryRow(ctx, "SELECT COALESCE(SUM(amount_cents), 0) FROM team_payments").Scan(&totalCents)
	_ = s.Pool.QueryRow(ctx,
		"SELECT COALESCE(SUM(amount_cents), 0) FROM team_payments WHERE paid_at > now() - interval '30 days'",
	).Scan(&last30Cents)
	_ = s.Pool.QueryRow(ctx, "SELECT count(DISTINCT team_id) FROM team_payments").Scan(&payingTeams)
	_ = s.Pool.QueryRow(ctx, `
		SELECT currency FROM team_payments
		WHERE currency IS NOT NULL AND currency <> ''
		GROUP BY currency
		ORDER BY count(*) DESC LIMIT 1`).Scan(&currency)
	if currency == "" {
		currency = "USD"
	}

	byPlan := map[string]int{}
	rows, err := s.Pool.Query(ctx,
		"SELECT COALESCE(subscription_plan, 'free') AS plan, count(*) FROM teams GROUP BY 1")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var plan string
			var cnt int
			if err := rows.Scan(&plan, &cnt); err == nil {
				byPlan[plan] = cnt
			}
		}
	}

	type recentPayment struct {
		ID          uuid.UUID `json:"id"`
		TeamID      uuid.UUID `json:"team_id"`
		TeamName    string    `json:"team_name"`
		AmountCents *int      `json:"amount_cents,omitempty"`
		Currency    *string   `json:"currency,omitempty"`
		Method      *string   `json:"method,omitempty"`
		CoversUntil time.Time `json:"covers_until"`
		PaidAt      time.Time `json:"paid_at"`
		Note        *string   `json:"note,omitempty"`
	}
	recent := []recentPayment{}
	rrows, err := s.Pool.Query(ctx, `
		SELECT p.id, p.team_id, t.name, p.amount_cents, p.currency, p.method, p.covers_until, p.paid_at, p.note
		FROM team_payments p
		JOIN teams t ON t.id = p.team_id
		ORDER BY p.paid_at DESC
		LIMIT 10`)
	if err == nil {
		defer rrows.Close()
		for rrows.Next() {
			var r recentPayment
			if err := rrows.Scan(&r.ID, &r.TeamID, &r.TeamName, &r.AmountCents, &r.Currency, &r.Method, &r.CoversUntil, &r.PaidAt, &r.Note); err == nil {
				recent = append(recent, r)
			}
		}
	}

	return c.JSON(fiber.Map{
		"total_cents":    totalCents,
		"last_30d_cents": last30Cents,
		"currency":       currency,
		"paying_teams":   payingTeams,
		"by_plan":        byPlan,
		"recent":         recent,
	})
}

func (s Service) handleAdminSSOList(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	type ssoEntry struct {
		TeamID         uuid.UUID `json:"team_id"`
		TeamName       string    `json:"team_name"`
		Provider       string    `json:"provider"`
		Enabled        bool      `json:"enabled"`
		OIDCIssuer     string    `json:"oidc_issuer"`
		OIDCClientID   string    `json:"oidc_client_id"`
		AllowedDomains []string  `json:"allowed_domains"`
		JITProvision   bool      `json:"jit_provision"`
		JITRole        string    `json:"jit_role"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
	}
	rows, err := s.Pool.Query(c.Context(), `
		SELECT sc.team_id, t.name, sc.provider, sc.enabled, COALESCE(sc.oidc_issuer, ''),
		       COALESCE(sc.oidc_client_id, ''), COALESCE(sc.allowed_domains, ARRAY[]::text[]),
		       sc.jit_provision, sc.jit_role, sc.created_at, sc.updated_at
		FROM sso_configs sc
		JOIN teams t ON t.id = sc.team_id
		ORDER BY sc.updated_at DESC`)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	out := []ssoEntry{}
	for rows.Next() {
		var e ssoEntry
		if err := rows.Scan(&e.TeamID, &e.TeamName, &e.Provider, &e.Enabled, &e.OIDCIssuer,
			&e.OIDCClientID, &e.AllowedDomains, &e.JITProvision, &e.JITRole, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return s.internalErr(c, err)
		}
		out = append(out, e)
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
	tag, err := s.Pool.Exec(c.Context(),
		"UPDATE sso_configs SET enabled = false, updated_at = now() WHERE team_id = $1", teamID)
	if err != nil {
		return s.internalErr(c, err)
	}
	if tag.RowsAffected() == 0 {
		return httperr.NotFound(c, "sso_not_configured", "no SSO config for team")
	}
	actorID, actorEmail := s.actorInfo(c)
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionSSOSaved, "team", teamID.String(), map[string]any{
		"enabled":   false,
		"force_off": true,
		"by":        "super_admin",
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

	var teamName string
	_ = s.Pool.QueryRow(c.Context(), "SELECT name FROM teams WHERE id=$1", teamID).Scan(&teamName)
	if _, err := s.Pool.Exec(c.Context(), "DELETE FROM teams WHERE id=$1", teamID); err != nil {
		return s.internalErr(c, err)
	}
	actorID, actorEmail := s.actorInfo(c)
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionTeamDeleted, "team", teamID.String(), map[string]any{
		"name": teamName,
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

	if uid == userID(c) {
		return httperr.Conflict(c, "cannot_delete_self", "cannot delete yourself")
	}

	var role string
	_ = s.Pool.QueryRow(c.Context(), "SELECT global_role FROM users WHERE id=$1", uid).Scan(&role)
	if role == "super_admin" {
		var count int
		_ = s.Pool.QueryRow(c.Context(),
			"SELECT count(*) FROM users WHERE global_role='super_admin'").Scan(&count)
		if count <= 1 {
			return httperr.Conflict(c, "last_super_admin", "cannot delete last super_admin")
		}
	}

	if d, ok := EventStore.(store.UserDeleter); ok {
		if err := d.DeleteUserData(c.Context(), uid.String()); err != nil {
			s.Logger.Warn("clickhouse delete failed (continuing)",
				zap.String("user_id", uid.String()), zap.Error(err))
		}
	}

	var victimEmail string
	_ = s.Pool.QueryRow(c.Context(), "SELECT email FROM users WHERE id=$1", uid).Scan(&victimEmail)
	if _, err := s.Pool.Exec(c.Context(), "DELETE FROM users WHERE id=$1", uid); err != nil {
		return s.internalErr(c, err)
	}
	actorID, actorEmail := s.actorInfo(c)
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionUserDeleted, "user", uid.String(), map[string]any{
		"email": victimEmail,
		"role":  role,
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
	if req.GlobalRole != nil {
		role := strings.ToLower(strings.TrimSpace(*req.GlobalRole))
		if role != "user" && role != "super_admin" {
			return httperr.BadRequest(c, "invalid_global_role", "global_role must be user | super_admin")
		}

		if uid == userID(c) && role != "super_admin" {
			return httperr.Conflict(c, "cannot_demote_self", "cannot demote yourself")
		}
		var prevRole, victimEmail string
		_ = s.Pool.QueryRow(c.Context(),
			"SELECT global_role, email FROM users WHERE id=$1", uid).Scan(&prevRole, &victimEmail)
		if _, err := s.Pool.Exec(c.Context(),
			"UPDATE users SET global_role=$1 WHERE id=$2", role, uid); err != nil {
			return s.internalErr(c, err)
		}

		_ = auth.BumpTokenVersion(c.Context(), s.Pool, uid)
		if prevRole != role {
			actorID, actorEmail := s.actorInfo(c)
			s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionUserRoleChanged, "user", uid.String(), map[string]any{
				"email":     victimEmail,
				"role_from": prevRole,
				"role_to":   role,
			})
		}
	}
	if req.DisplayName != nil {
		dn, ok := validateDisplayName(*req.DisplayName)
		if !ok {
			return httperr.BadRequest(c, "invalid_display_name", "display_name 1..64 chars, no newlines")
		}
		if _, err := s.Pool.Exec(c.Context(),
			"UPDATE users SET display_name=$1 WHERE id=$2", dn, uid); err != nil {
			return s.internalErr(c, err)
		}
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
	email, ok := validateEmail(req.Email)
	if !ok {
		return httperr.BadRequest(c, "invalid_email", "valid email required")
	}
	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role == "" {
		role = "member"
	}
	if role != "owner" && role != "admin" && role != "member" {
		return httperr.BadRequest(c, "invalid_role", "role must be owner | admin | member")
	}
	user, err := auth.FindUserByEmail(c.Context(), s.Pool, email)
	if err != nil {
		return httperr.NotFound(c, "user_not_found", "user with this email not found")
	}

	if role == "owner" {
		var existingOwned int
		_ = s.Pool.QueryRow(c.Context(),
			"SELECT count(*) FROM team_members WHERE user_id=$1 AND role='owner' AND team_id<>$2",
			user.ID, teamID).Scan(&existingOwned)
		if existingOwned > 0 {
			return httperr.Conflict(c, "owner_limit", "user already owns another company")
		}
	}

	if _, err := s.Pool.Exec(c.Context(), `
		INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (team_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		teamID, user.ID, role); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true, "user_id": user.ID})
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

	var planNorm string
	if req.Plan != nil {
		planNorm = strings.ToLower(strings.TrimSpace(*req.Plan))
		if planNorm != "free" && planNorm != "pro" && planNorm != "team" && planNorm != "enterprise" {
			return httperr.BadRequest(c, "invalid_plan", "plan must be free | pro | team | enterprise")
		}
	}
	var untilTS *time.Time
	clearUntil := false
	if req.Until != nil {
		if *req.Until == "" {
			clearUntil = true
		} else {
			ts, err := time.Parse(time.RFC3339, *req.Until)
			if err != nil {
				return httperr.BadRequest(c, "invalid_until", "until must be ISO8601 (RFC3339)")
			}
			untilTS = &ts
		}
	}
	var coversUntil *time.Time
	if req.Payment != nil {
		if req.Payment.AmountCents <= 0 {
			return httperr.BadRequest(c, "invalid_payment_amount", "payment.amount_cents must be > 0 if payment is set")
		}
		ts, err := time.Parse(time.RFC3339, req.Payment.CoversUntil)
		if err != nil {
			return httperr.BadRequest(c, "invalid_covers_until", "payment.covers_until must be ISO8601")
		}
		coversUntil = &ts
	}

	tx, err := s.Pool.Begin(c.Context())
	if err != nil {
		return s.internalErr(c, err)
	}
	defer tx.Rollback(c.Context())

	if req.Plan != nil {
		if _, err := tx.Exec(c.Context(),
			"UPDATE teams SET subscription_plan=$1 WHERE id=$2", planNorm, teamID); err != nil {
			return s.internalErr(c, err)
		}
	}
	if clearUntil {
		if _, err := tx.Exec(c.Context(),
			"UPDATE teams SET subscription_until=NULL WHERE id=$1", teamID); err != nil {
			return s.internalErr(c, err)
		}
	} else if untilTS != nil {
		if _, err := tx.Exec(c.Context(),
			"UPDATE teams SET subscription_until=$1 WHERE id=$2", *untilTS, teamID); err != nil {
			return s.internalErr(c, err)
		}
	}
	if req.Note != nil {
		note := strings.TrimSpace(*req.Note)
		var noteVal any
		if note == "" {
			noteVal = nil
		} else {
			noteVal = note
		}
		if _, err := tx.Exec(c.Context(),
			"UPDATE teams SET subscription_note=$1 WHERE id=$2", noteVal, teamID); err != nil {
			return s.internalErr(c, err)
		}
	}

	var paymentID *uuid.UUID
	if coversUntil != nil {
		method := strings.TrimSpace(req.Payment.Method)
		if method == "" {
			method = "manual_transfer"
		}
		currency := strings.ToUpper(strings.TrimSpace(req.Payment.Currency))
		if currency == "" {
			currency = "USD"
		}
		pid := uuid.New()
		if _, err := tx.Exec(c.Context(), `
			INSERT INTO team_payments
			  (id, team_id, amount_cents, currency, method, note, covers_until, recorded_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			pid, teamID, req.Payment.AmountCents, currency, method,
			strings.TrimSpace(req.Payment.Note), *coversUntil, userID(c)); err != nil {
			return s.internalErr(c, err)
		}
		paymentID = &pid
	}

	if err := tx.Commit(c.Context()); err != nil {
		return s.internalErr(c, err)
	}
	actorID, actorEmail := s.actorInfo(c)
	meta := map[string]any{}
	if req.Plan != nil {
		meta["plan_to"] = planNorm
	}
	if req.Until != nil {
		if untilTS != nil {
			meta["until"] = untilTS.Format(time.RFC3339)
		} else {
			meta["until"] = nil
		}
	}
	if paymentID != nil {
		meta["payment_id"] = paymentID.String()
		meta["amount_cents"] = req.Payment.AmountCents
		meta["currency"] = req.Payment.Currency
	}
	s.Audit.LogFromCtx(c, actorID, actorEmail, audit.ActionSubscriptionSet, "team", teamID.String(), meta)
	return c.JSON(fiber.Map{"ok": true, "payment_id": paymentID})
}

func (s Service) handleListPayments(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	rows, err := s.Pool.Query(c.Context(), `
		SELECT id, amount_cents, currency, method, note, covers_until, paid_at, recorded_by
		FROM team_payments WHERE team_id=$1 ORDER BY paid_at DESC LIMIT 200`, teamID)
	if err != nil {
		return s.internalErr(c, err)
	}
	defer rows.Close()
	type payment struct {
		ID          uuid.UUID `json:"id"`
		AmountCents int       `json:"amount_cents"`
		Currency    string    `json:"currency"`
		Method      string    `json:"method"`
		Note        string    `json:"note"`
		CoversUntil time.Time `json:"covers_until"`
		PaidAt      time.Time `json:"paid_at"`
		RecordedBy  uuid.UUID `json:"recorded_by"`
	}
	out := []payment{}
	for rows.Next() {
		var p payment
		var note *string
		if err := rows.Scan(&p.ID, &p.AmountCents, &p.Currency, &p.Method, &note,
			&p.CoversUntil, &p.PaidAt, &p.RecordedBy); err != nil {
			return s.internalErr(c, err)
		}
		if note != nil {
			p.Note = *note
		}
		out = append(out, p)
	}
	return c.JSON(fiber.Map{"payments": out})
}
