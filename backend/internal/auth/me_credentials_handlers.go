package auth

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth/meapp"
	"github.com/eye-of-providence/backend/internal/httperr"
)

func registerMeCredentialRoutes(g fiber.Router, s MeService) {
	g.Patch("/name", handleChangeMyName(s))
	g.Patch("/email", handleChangeMyEmail(s))
	g.Patch("/password", handleChangeMyPassword(s))
}

func handleChangeMyName(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if s.Pool == nil {
			return httperr.Unavailable(c, "db_required", "auth requires postgres")
		}
		var req struct {
			DisplayName *string `json:"display_name"`
			LastName    *string `json:"last_name"`
		}
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		var dn, ln *string
		if req.DisplayName != nil {
			v, ok := ValidateDisplayName(*req.DisplayName)
			if !ok {
				return httperr.BadRequest(c, "invalid_display_name", "display_name 1..64 chars, no newlines")
			}
			dn = &v
		}
		if req.LastName != nil {
			trimmed := strings.TrimSpace(*req.LastName)
			if trimmed == "" {
				empty := ""
				ln = &empty
			} else {
				v, ok := ValidateDisplayName(trimmed)
				if !ok {
					return httperr.BadRequest(c, "invalid_last_name", "last_name up to 64 chars, no newlines")
				}
				ln = &v
			}
		}
		uid, err := uuid.Parse(ClaimsFromCtx(c).UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		app := newMeAppService(s)
		if err := app.PatchName(c.Context(), uid, dn, ln); err != nil {
			if errors.Is(err, meapp.ErrNoFields) {
				return httperr.BadRequest(c, "no_fields", "provide display_name or last_name")
			}
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}

func handleChangeMyEmail(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if s.Pool == nil {
			return httperr.Unavailable(c, "db_required", "auth requires postgres")
		}
		var req struct {
			CurrentPassword string `json:"current_password"`
			Email           string `json:"email"`
		}
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		email, ok := ValidateEmail(req.Email)
		if !ok {
			return httperr.BadRequest(c, "invalid_email", "valid email required")
		}
		uid, err := uuid.Parse(ClaimsFromCtx(c).UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		app := newMeAppService(s)
		tok, newEmail, err := app.ChangeEmail(c.Context(), uid, email, req.CurrentPassword, VerifyPassword)
		if err != nil {
			switch {
			case errors.Is(err, meapp.ErrNoPasswordSet):
				return httperr.BadRequest(c, "no_password_set", "set a password first (account linked via OAuth)")
			case errors.Is(err, meapp.ErrInvalidCredentials):
				return httperr.Unauthorized(c, "invalid_credentials", "incorrect password")
			case errors.Is(err, meapp.ErrEmailTaken):
				return httperr.Conflict(c, "email_taken", "email already in use")
			default:
				return httperr.Internal(c)
			}
		}
		return c.JSON(fiber.Map{"token": tok, "email": newEmail})
	}
}

func handleChangeMyPassword(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if s.Pool == nil {
			return httperr.Unavailable(c, "db_required", "auth requires postgres")
		}
		var req struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		if !ValidatePassword(req.NewPassword) {
			return httperr.BadRequest(c, "invalid_password", "password must be 8..256 chars")
		}
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		app := newMeAppService(s)
		tok, err := app.ChangePassword(c.Context(), uid, claims.Email, req.CurrentPassword, req.NewPassword, VerifyPassword, HashPassword)
		if err != nil {
			switch {
			case errors.Is(err, meapp.ErrNoPasswordSet):
				return httperr.BadRequest(c, "no_password_set", "set a password first (account linked via OAuth)")
			case errors.Is(err, meapp.ErrInvalidCredentials):
				return httperr.Unauthorized(c, "invalid_credentials", "incorrect current password")
			default:
				return httperr.Internal(c)
			}
		}
		return c.JSON(fiber.Map{"token": tok})
	}
}
