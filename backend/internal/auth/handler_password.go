package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth/passwordresetapp"
	"github.com/eye-of-providence/backend/internal/httperr"
)

func RegisterPasswordResetRoutes(app *fiber.App, s PasswordResetService) {
	g := app.Group("/v1/auth")
	g.Post("/forgot-password", handleForgotPassword(s))
	g.Post("/reset-password", handleResetPassword(s))
}

func handleForgotPassword(s PasswordResetService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if s.Pool == nil {
			return c.JSON(fiber.Map{"status": "ok"})
		}
		var req struct {
			Email string `json:"email"`
		}
		_ = c.BodyParser(&req)
		email, ok := ValidateEmail(req.Email)
		if !ok {
			return c.JSON(fiber.Map{"status": "ok"})
		}
		app := newPasswordResetApp(s)
		if err := app.RequestReset(c.Context(), email); err != nil && s.Logger != nil {
			s.Logger.Warn("forgot-password failed", zap.Error(err))
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}

func handleResetPassword(s PasswordResetService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if s.Pool == nil {
			return httperr.Unavailable(c, "db_required", "auth requires postgres")
		}
		var req struct {
			Token    string `json:"token"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		app := newPasswordResetApp(s)
		err := app.ResetPassword(c.Context(), req.Token, req.Password, ValidatePassword, HashPassword)
		if err != nil {
			switch {
			case errors.Is(err, passwordresetapp.ErrInvalidPassword):
				return httperr.BadRequest(c, "invalid_password", "password must be 8..256 chars")
			case errors.Is(err, passwordresetapp.ErrMissingToken):
				return httperr.BadRequest(c, "missing_token", "missing token")
			case errors.Is(err, passwordresetapp.ErrTokenInvalid):
				return httperr.BadRequest(c, "reset_token_invalid", "invalid or expired reset token")
			default:
				if s.Logger != nil {
					s.Logger.Warn("reset-password failed", zap.Error(err))
				}
				return httperr.Internal(c)
			}
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}
