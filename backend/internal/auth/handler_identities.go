package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth/identitiesapp"
	"github.com/eye-of-providence/backend/internal/auth/passkeysapp"
	"github.com/eye-of-providence/backend/internal/httperr"
)

func RegisterIdentitiesRoutes(app *fiber.App, s Service) {
	g := app.Group("/v1/me", Middleware(s.JWTSecret, s.Pool))
	g.Get("/identities", handleListIdentities(s))
	g.Delete("/identities/:id", handleDeleteIdentity(s))

	if s.WebAuthn != nil {
		g.Get("/passkeys", handleListPasskeys(s))
		g.Delete("/passkeys/:id", handleDeletePasskey(s))
	}
}

func handleListIdentities(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		rows, err := newIdentitiesApp(s.Pool).List(c.Context(), uid)
		if err != nil {
			return s.internalErr(c, err)
		}
		return c.JSON(fiber.Map{"identities": rows})
	}
}

func handleDeleteIdentity(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		identityID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return httperr.BadRequest(c, "invalid_id", "invalid identity id")
		}
		if err := newIdentitiesApp(s.Pool).Delete(c.Context(), uid, identityID); err != nil {
			switch {
			case errors.Is(err, identitiesapp.ErrDBRequired):
				return httperr.Unavailable(c, "db_required", "identities require database")
			case errors.Is(err, identitiesapp.ErrLastAuthFactor):
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":  "last_auth_factor",
					"detail": "cannot remove last sign-in method; set a password or add a passkey first",
				})
			case errors.Is(err, identitiesapp.ErrNotFound):
				return httperr.NotFound(c, "not_found", "identity not found")
			default:
				return s.internalErr(c, err)
			}
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

func handleListPasskeys(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		rows, err := newPasskeysApp(s.WebAuthn, s.Pool).List(c.Context(), uid)
		if err != nil {
			return s.internalErr(c, err)
		}
		return c.JSON(fiber.Map{"passkeys": rows})
	}
}

func handleDeletePasskey(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		passkeyID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return httperr.BadRequest(c, "invalid_id", "invalid passkey id")
		}
		if err := newPasskeysApp(s.WebAuthn, s.Pool).Delete(c.Context(), uid, passkeyID); err != nil {
			switch {
			case errors.Is(err, passkeysapp.ErrLastAuthFactor):
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":  "last_auth_factor",
					"detail": "cannot remove last sign-in method; set a password or add an identity first",
				})
			case errors.Is(err, passkeysapp.ErrPasskeyNotFound):
				return httperr.NotFound(c, "not_found", "passkey not found")
			default:
				return s.internalErr(c, err)
			}
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}
