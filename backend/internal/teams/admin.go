// admin.go — super_admin endpoints: list teams/users, stats, delete user/team,
// update user, add member, set subscription, list payments.
//
// Все требуют requireSuperAdmin (или явный isSuperAdmin) gate в начале.
package teams

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/store"
)

// adminListPagination — общий парсер ?limit=N&offset=N для admin-list эндпоинтов.
// limit max=200, default=100. offset >= 0.
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

// requireSuperAdmin — guard для super_admin handler'ов. Возвращает true если пускаем.
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

func (s Service) handleAdminDeleteTeam(c *fiber.Ctx) error {
	if !s.requireSuperAdmin(c) {
		return nil
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "invalid team id")
	}
	if _, err := s.Pool.Exec(c.Context(), "DELETE FROM teams WHERE id=$1", teamID); err != nil {
		return s.internalErr(c, err)
	}
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
	// Запрет на самоудаление, чтобы super_admin не остался без аккаунта.
	if uid == userID(c) {
		return httperr.Conflict(c, "cannot_delete_self", "cannot delete yourself")
	}
	// Защита: не оставить систему без хотя бы одного super_admin'а.
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
	// ClickHouse cleanup ДО Postgres-удаления — после удаления users-row мы потеряем
	// uid, и события юзера останутся orphan'ами (не привязаны к существующему юзеру).
	if d, ok := EventStore.(store.UserDeleter); ok {
		if err := d.DeleteUserData(c.Context(), uid.String()); err != nil {
			s.Logger.Warn("clickhouse delete failed (continuing)",
				zap.String("user_id", uid.String()), zap.Error(err))
		}
	}
	if _, err := s.Pool.Exec(c.Context(), "DELETE FROM users WHERE id=$1", uid); err != nil {
		return s.internalErr(c, err)
	}
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
		// Запрет понизить себя — иначе можно ребутнуть систему без super_admin'а.
		if uid == userID(c) && role != "super_admin" {
			return httperr.Conflict(c, "cannot_demote_self", "cannot demote yourself")
		}
		if _, err := s.Pool.Exec(c.Context(),
			"UPDATE users SET global_role=$1 WHERE id=$2", role, uid); err != nil {
			return s.internalErr(c, err)
		}
		// Инвалидируем существующие JWT этого юзера — иначе при демоуте super_admin
		// сохранил бы доступ к /v1/admin/* до истечения 14d токена.
		_ = auth.BumpTokenVersion(c.Context(), s.Pool, uid)
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
	// Защита 1-owner invariant: super_admin может назначить owner'а на любую команду,
	// но не на ту, где этот юзер уже owner какой-то другой команды.
	if role == "owner" {
		var existingOwned int
		_ = s.Pool.QueryRow(c.Context(),
			"SELECT count(*) FROM team_members WHERE user_id=$1 AND role='owner' AND team_id<>$2",
			user.ID, teamID).Scan(&existingOwned)
		if existingOwned > 0 {
			return httperr.Conflict(c, "owner_limit", "user already owns another company")
		}
	}
	// UPSERT — если юзер уже member, обновим роль (вместо silent no-op).
	if _, err := s.Pool.Exec(c.Context(), `
		INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (team_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		teamID, user.ID, role); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"ok": true, "user_id": user.ID})
}

type setSubscriptionReq struct {
	Plan  *string `json:"plan"`  // "free" | "pro" | "team" | "enterprise"
	Until *string `json:"until"` // ISO8601 timestamptz; nil = не менять; "" = очистить (revoke)
	Note  *string `json:"note"`  // публичная заметка (видна owner'у)

	// Опциональный payment record. Если есть amount_cents > 0 и covers_until — пишем в team_payments.
	Payment *struct {
		AmountCents int    `json:"amount_cents"`
		Currency    string `json:"currency"`
		Method      string `json:"method"`
		Note        string `json:"note"`
		CoversUntil string `json:"covers_until"` // ISO8601
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

	// Вся валидация — ДО tx.Begin, чтобы плохой запрос не открывал idle txn slot
	// (DoS amplification).
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
