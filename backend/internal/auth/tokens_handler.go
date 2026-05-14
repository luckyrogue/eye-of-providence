package auth

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth/meapp"
	"github.com/eye-of-providence/backend/internal/httperr"
)

func listTokensHandler(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if ScopeFromCtx(c) != "" {
			return httperr.Forbidden(c, "jwt_required", "tokens management requires JWT (dashboard session)")
		}
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid subject")
		}
		app := newMeAppService(s)
		tokens, err := app.ListAPITokens(c.Context(), uid)
		if err != nil {
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"tokens": tokens})
	}
}

func createTokenHandler(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if ScopeFromCtx(c) != "" {
			return httperr.Forbidden(c, "jwt_required", "tokens management requires JWT")
		}
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid subject")
		}
		var req struct {
			Name    string `json:"name"`
			Scope   string `json:"scope"`
			TTLDays int    `json:"ttl_days"`
		}
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		app := newMeAppService(s)
		plaintext, row, err := app.CreateAPIToken(c.Context(), uid, meapp.CreateAPITokenInput{
			Name: req.Name, Scope: req.Scope, TTLDays: req.TTLDays,
		})
		if err != nil {
			return mapMeAppTokenErr(c, err)
		}
		return c.JSON(fiber.Map{
			"token":    plaintext,
			"metadata": row,
		})
	}
}

func revokeTokenHandler(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if ScopeFromCtx(c) != "" {
			return httperr.Forbidden(c, "jwt_required", "tokens management requires JWT")
		}
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid subject")
		}
		tokenID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return httperr.BadRequest(c, "invalid_id", "invalid token id")
		}
		app := newMeAppService(s)
		ok, err := app.RevokeAPIToken(c.Context(), uid, tokenID)
		if err != nil {
			return mapMeAppTokenErr(c, err)
		}
		if !ok {
			return httperr.NotFound(c, "token_not_found", "token not found")
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}

func mapMeAppTokenErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, meapp.ErrDBNotConfigured):
		return httperr.Unavailable(c, "db_not_configured", "database not configured")
	case errors.Is(err, meapp.ErrNameTooLong):
		return httperr.BadRequest(c, "name_too_long", err.Error())
	case errors.Is(err, meapp.ErrTTLOutOfRange):
		return httperr.BadRequest(c, "ttl_out_of_range", err.Error())
	default:
		if strings.Contains(err.Error(), "invalid scope") {
			return httperr.BadRequest(c, "invalid_scope", err.Error())
		}
		return httperr.Internal(c)
	}
}
