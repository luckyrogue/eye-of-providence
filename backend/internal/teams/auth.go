// auth.go — register / login / public auth-config endpoints.
//
// Forgot/reset вынесены в password_reset.go.
package teams

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
)

// --- Public auth config (для UI) ---

func (s Service) handleAuthConfig(c *fiber.Ctx) error {
	var userCount int
	if s.Pool != nil {
		_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM users").Scan(&userCount)
	}
	return c.JSON(fiber.Map{
		"invite_only":   s.InviteOnly,
		"is_first_user": userCount == 0,
	})
}

// --- Register / login ---

type registerReq struct {
	Email       string  `json:"email"`
	DisplayName string  `json:"display_name"`
	Password    string  `json:"password"`
	InviteCode  *string `json:"invite_code,omitempty"`
}

func (s Service) handleRegister(c *fiber.Ctx) error {
	if s.Pool == nil {
		return httperr.Unavailable(c, "db_required", "auth requires postgres")
	}
	var req registerReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	email, ok := validateEmail(req.Email)
	if !ok {
		return httperr.BadRequest(c, "invalid_email", "valid email required")
	}
	req.Email = email
	if !validatePassword(req.Password) {
		return httperr.BadRequest(c, "invalid_password", "password must be 8..256 chars")
	}
	if req.DisplayName == "" {
		req.DisplayName = req.Email
	}
	dn, ok := validateDisplayName(req.DisplayName)
	if !ok {
		return httperr.BadRequest(c, "invalid_display_name", "display_name 1..64 chars, no newlines")
	}
	req.DisplayName = dn

	// Подсчёт пользователей. Первый user может зарегаться без invite (bootstrap),
	// он же станет super_admin.
	var userCount int
	_ = s.Pool.QueryRow(c.Context(), "SELECT count(*) FROM users").Scan(&userCount)
	isFirstUser := userCount == 0

	if s.InviteOnly && !isFirstUser {
		if req.InviteCode == nil || *req.InviteCode == "" {
			return httperr.Forbidden(c, "invite_required", "registration is invite-only")
		}
		// Проверяем валидность ДО создания юзера
		if _, err := s.findInvite(c.Context(), *req.InviteCode); err != nil {
			return httperr.BadRequest(c, "invite_invalid", "invite invalid or expired")
		}
	}

	user, err := auth.CreateUser(c.Context(), s.Pool, req.Email, req.DisplayName, req.Password, nil)
	if err != nil {
		return httperr.Conflict(c, "email_taken", "email already taken (or DB error)")
	}

	if isFirstUser {
		_, _ = s.Pool.Exec(c.Context(),
			"UPDATE users SET global_role='super_admin' WHERE id=$1", user.ID)
		s.Logger.Info("first user promoted to super_admin", zap.String("email", user.Email))
	}

	// Атомарно: invite consumes (защищает от race), затем добавляем юзера.
	var joinedTeam *uuid.UUID
	if req.InviteCode != nil && *req.InviteCode != "" {
		teamID, err := s.consumeInvite(c.Context(), *req.InviteCode, user.ID)
		if err == nil {
			_ = s.addMember(c.Context(), teamID, user.ID, "member")
			joinedTeam = &teamID
		}
	}

	tv, _ := auth.TokenVersion(c.Context(), s.Pool, user.ID)
	tok, err := auth.IssueJWT(s.JWTSecret, user.ID.String(), user.Email, "password", tv, tokenTTL)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{
		"token":        tok,
		"user_id":      user.ID,
		"display_name": user.DisplayName,
		"team_id":      joinedTeam,
	})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s Service) handleLogin(c *fiber.Ctx) error {
	if s.Pool == nil {
		return httperr.Unavailable(c, "db_required", "auth requires postgres")
	}
	var req loginReq
	if err := c.BodyParser(&req); err != nil {
		return httperr.BadRequest(c, "invalid_body", "invalid body")
	}
	email, ok := validateEmail(req.Email)
	if !ok || !validatePassword(req.Password) {
		// Не подсказываем, что именно не так — это login surface.
		return httperr.Unauthorized(c, "invalid_credentials", "invalid email or password")
	}
	req.Email = email
	user, err := auth.FindUserByEmail(c.Context(), s.Pool, req.Email)
	if err != nil || !auth.VerifyPassword(user.PasswordHash, req.Password) {
		return httperr.Unauthorized(c, "invalid_credentials", "invalid email or password")
	}
	tv, _ := auth.TokenVersion(c.Context(), s.Pool, user.ID)
	tok, err := auth.IssueJWT(s.JWTSecret, user.ID.String(), user.Email, "password", tv, tokenTTL)
	if err != nil {
		return s.internalErr(c, err)
	}
	return c.JSON(fiber.Map{
		"token":        tok,
		"user_id":      user.ID,
		"display_name": user.DisplayName,
	})
}
