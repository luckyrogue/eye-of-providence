package teams

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/auth/passwordapp"
	"github.com/eye-of-providence/backend/internal/teams/authapp"
	"github.com/eye-of-providence/backend/internal/teams/invitesapp"
	"github.com/eye-of-providence/backend/internal/httperr"
)

func (s Service) handleAuthConfig(c *fiber.Ctx) error {
	rc, err := s.registrationApp().BeforeRegister(c.Context())
	if err != nil {
		return s.internalErr(c, err)
	}
	providers := s.AuthProviders
	if providers == nil {
		providers = []string{}
	}
	return c.JSON(fiber.Map{
		"invite_only":     s.InviteOnly,
		"is_first_user":   rc.IsFirstUser,
		"providers":       providers,
		"passkey_enabled": s.PasskeyEnabled,
	})
}

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

	rc, err := s.registrationApp().BeforeRegister(c.Context())
	if err != nil {
		return s.internalErr(c, err)
	}

	if s.InviteOnly && !rc.IsFirstUser {
		if req.InviteCode == nil || *req.InviteCode == "" {
			return httperr.Forbidden(c, "invite_required", "registration is invite-only")
		}
		if _, err := s.invitesApp().Find(c.Context(), *req.InviteCode); err != nil {
			return httperr.BadRequest(c, "invite_invalid", "invite invalid or expired")
		}
	}

	user, err := auth.CreateUser(c.Context(), s.Pool, req.Email, req.DisplayName, req.Password, nil)
	if err != nil {
		return httperr.Conflict(c, "email_taken", "email already taken (or DB error)")
	}

	_ = s.registrationApp().AfterRegister(c.Context(), user.ID.String(), rc)

	var joinedTeam *uuid.UUID
	if req.InviteCode != nil && *req.InviteCode != "" {
		if teamID, err := s.invitesApp().Accept(c.Context(), *req.InviteCode, user.ID); err == nil {
			joinedTeam = &teamID
		} else if errors.Is(err, invitesapp.ErrPlanLimitExceeded) {
			return httperr.Forbidden(c, "plan_limit_exceeded", err.Error())
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
		return httperr.Unauthorized(c, "invalid_credentials", "invalid email or password")
	}
	req.Email = email
	user, err := s.authApp().Login(c.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, passwordapp.ErrInvalidCredentials) || errors.Is(err, authapp.ErrInvalidCredentials) {
			return httperr.Unauthorized(c, "invalid_credentials", "invalid email or password")
		}
		return s.internalErr(c, err)
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
