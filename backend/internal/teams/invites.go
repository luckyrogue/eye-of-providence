package teams

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/mailer"
)

type Invite struct {
	ID       uuid.UUID
	TeamID   uuid.UUID
	Code     string
	MaxUses  int
	UseCount int
	Expires  *time.Time
}

func (s Service) findInvite(ctx context.Context, code string) (*Invite, error) {
	var inv Invite
	err := s.Pool.QueryRow(ctx, `
		SELECT id, team_id, code, max_uses, use_count, expires_at
		FROM team_invites
		WHERE code = $1
		  AND (expires_at IS NULL OR expires_at > now())
		  AND use_count < max_uses
		LIMIT 1`, code).Scan(&inv.ID, &inv.TeamID, &inv.Code, &inv.MaxUses, &inv.UseCount, &inv.Expires)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (s Service) consumeInvite(ctx context.Context, code string, userID uuid.UUID) (uuid.UUID, error) {
	var teamID uuid.UUID
	err := s.Pool.QueryRow(ctx, `
		UPDATE team_invites
		SET use_count = use_count + 1, used_by = $2, used_at = now()
		WHERE code = $1
		  AND use_count < max_uses
		  AND (expires_at IS NULL OR expires_at > now())
		RETURNING team_id`, code, userID).Scan(&teamID)
	return teamID, err
}

func (s Service) handleCreateInvite(c *fiber.Ctx) error {
	uid := userID(c)
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.BadRequest(c, "invalid_team_id", "bad team id")
	}
	role, ok := s.teamRole(c.Context(), uid, teamID)
	if !ok || (role != "owner" && role != "admin") {
		return httperr.Forbidden(c, "role_insufficient", "only owner/admin can create invites")
	}

	var req struct {
		Email string `json:"email"`
	}
	_ = c.BodyParser(&req)
	var emailPtr *string
	if req.Email != "" {
		clean, ok := validateEmail(req.Email)
		if !ok {
			return httperr.BadRequest(c, "invalid_email", "invalid email")
		}

		emailPtr = &clean
	}
	maxUses := 10
	if emailPtr != nil {
		maxUses = 1
	}

	code := randomCode(16)
	expires := time.Now().Add(7 * 24 * time.Hour)
	if _, err = s.Pool.Exec(c.Context(), `
		INSERT INTO team_invites (team_id, code, created_by, max_uses, expires_at, email)
		VALUES ($1, $2, $3, $4, $5, $6)`, teamID, code, uid, maxUses, expires, emailPtr); err != nil {
		return s.internalErr(c, err)
	}

	sentAt := time.Time{}
	if emailPtr != nil && s.Mailer != nil {
		if err := s.sendInviteEmail(c.Context(), teamID, uid, *emailPtr, code); err != nil {
			s.Logger.Warn("invite email send failed",
				zap.String("team", teamID.String()),
				zap.String("to", *emailPtr),
				zap.Error(err))
		} else {
			sentAt = time.Now()
			_, _ = s.Pool.Exec(c.Context(),
				`UPDATE team_invites SET sent_at = $1 WHERE code = $2`, sentAt, code)
		}
	}

	out := fiber.Map{"code": code, "expires_at": expires, "max_uses": maxUses}
	if emailPtr != nil {
		out["email"] = *emailPtr
		out["sent"] = !sentAt.IsZero()
	}
	return c.JSON(out)
}

func (s Service) sendInviteEmail(ctx context.Context, teamID, inviterID uuid.UUID, to, code string) error {
	var teamName, inviterName string
	var inviterLocale *string
	_ = s.Pool.QueryRow(ctx, `SELECT name FROM teams WHERE id = $1`, teamID).Scan(&teamName)
	_ = s.Pool.QueryRow(ctx,
		`SELECT COALESCE(display_name, ''), locale FROM users WHERE id = $1`, inviterID,
	).Scan(&inviterName, &inviterLocale)

	loc := mailer.Locale("")
	if inviterLocale != nil {
		loc = mailer.Locale(*inviterLocale)
	}
	inviteURL := strings.TrimRight(s.PublicURL, "/") + "/?invite=" + code
	subject, html, text := mailer.InviteEmail(teamName, inviteURL, inviterName, loc)
	return s.Mailer.Send(ctx, to, subject, html, text)
}

func (s Service) handleInvitePreview(c *fiber.Ctx) error {
	inv, err := s.findInvite(c.Context(), c.Params("code"))
	if err != nil {
		return httperr.NotFound(c, "invite_invalid", "invalid or expired invite")
	}
	var teamName string
	_ = s.Pool.QueryRow(c.Context(), "SELECT name FROM teams WHERE id=$1", inv.TeamID).Scan(&teamName)
	return c.JSON(fiber.Map{
		"valid":      true,
		"team_id":    inv.TeamID,
		"team_name":  teamName,
		"uses_left":  inv.MaxUses - inv.UseCount,
		"expires_at": inv.Expires,
	})
}

func (s Service) handleInviteAccept(c *fiber.Ctx) error {
	uid := userID(c)

	inv, err := s.findInvite(c.Context(), c.Params("code"))
	if err != nil {
		return httperr.NotFound(c, "invite_invalid", "invalid or expired invite")
	}
	var teamPlan string
	_ = s.Pool.QueryRow(c.Context(),
		"SELECT subscription_plan FROM teams WHERE id=$1", inv.TeamID).Scan(&teamPlan)
	limits := s.Plans.Limits(teamPlan)
	if limits.MaxUsersPerTeam > 0 {
		var count int
		if err := s.Pool.QueryRow(c.Context(),
			"SELECT count(*) FROM team_members WHERE team_id=$1", inv.TeamID).Scan(&count); err != nil {
			return s.internalErr(c, err)
		}
		if count >= limits.MaxUsersPerTeam {
			return httperr.Forbidden(c, "plan_limit_exceeded",
				fmt.Sprintf("team is on %s plan, limited to %d members (current: %d)",
					limits.Plan, limits.MaxUsersPerTeam, count))
		}
	}

	teamID, err := s.consumeInvite(c.Context(), c.Params("code"), uid)
	if err != nil {
		return httperr.NotFound(c, "invite_invalid", "invalid or expired invite")
	}
	if err := s.addMember(c.Context(), teamID, uid, "member"); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"team_id": teamID})
}
