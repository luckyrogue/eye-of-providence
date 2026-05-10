// invites.go — invite-link / invite-by-email lifecycle:
// findInvite (read), consumeInvite (atomic accept), handleCreateInvite,
// sendInviteEmail, handleInvitePreview, handleInviteAccept.
package teams

import (
	"context"
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

// consumeInvite — атомарно увеличивает use_count если invite ещё валиден.
// Возвращает team_id если получилось (invite ещё не исчерпан и не expired),
// иначе ошибку. Защищает от concurrent accept'ов с одной же ссылки.
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

	// Optional payload: {"email": "..."}. Без email — link-only (как раньше).
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
		// Email-invite — только 1 use (линк не должен переисполь­зоваться).
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

	// Если email задан — шлём письмо. Любой fail при отправке логируем, но
	// не падаем: инвайт уже создан, юзер может скопировать ссылку руками.
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

// sendInviteEmail — собирает данные (имя команды + display_name приглашающего)
// и шлёт письмо через Mailer. Не делает ретраев — Mailer.Send сам отвечает за HTTP.
//
// Locale: invite шлётся "слепо" — у получателя ещё нет аккаунта (значит и
// users.locale). Используем locale приглашающего как best-guess (часто это
// тот же регион). Fallback в NormalizeLocale (ru).
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
	// consumeInvite атомарно проверяет лимиты и инкрементирует use_count.
	teamID, err := s.consumeInvite(c.Context(), c.Params("code"), uid)
	if err != nil {
		return httperr.NotFound(c, "invite_invalid", "invalid or expired invite")
	}
	if err := s.addMember(c.Context(), teamID, uid, "member"); err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{"team_id": teamID})
}
