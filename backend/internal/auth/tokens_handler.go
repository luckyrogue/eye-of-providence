package auth

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/httperr"
)

// listTokensHandler — GET /v1/me/tokens.
// Anti-bootstrap: API token не может смотреть/создавать токены.
func listTokensHandler(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if ScopeFromCtx(c) != "" {
			return httperr.Forbidden(c, "jwt_required", "tokens management requires JWT (dashboard session)")
		}
		if s.Pool == nil {
			return c.JSON(fiber.Map{"tokens": []APIToken{}})
		}
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid subject")
		}
		tokens, err := ListAPITokens(c.Context(), s.Pool, uid)
		if err != nil {
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"tokens": tokens})
	}
}

// createTokenHandler — POST /v1/me/tokens. Body: {name, scope, ttl_days}.
// Возвращает plaintext "token" ровно один раз.
func createTokenHandler(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if ScopeFromCtx(c) != "" {
			return httperr.Forbidden(c, "jwt_required", "tokens management requires JWT")
		}
		if s.Pool == nil {
			return httperr.Unavailable(c, "db_not_configured", "database not configured")
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
		req.Name = strings.TrimSpace(req.Name)
		if len(req.Name) > 64 {
			return httperr.BadRequest(c, "name_too_long", "name too long (max 64)")
		}
		if req.Scope == "" {
			req.Scope = "read"
		}
		if req.TTLDays < 0 || req.TTLDays > 365 {
			return httperr.BadRequest(c, "ttl_out_of_range", "ttl_days must be 0..365")
		}
		ttl := time.Duration(req.TTLDays) * 24 * time.Hour

		plaintext, row, err := CreateAPIToken(c.Context(), s.Pool, uid, req.Name, req.Scope, ttl)
		if err != nil {
			return httperr.BadRequest(c, "create_failed", err.Error())
		}
		// Plaintext отдаётся ровно один раз. UI должен показать
		// "скопируй сейчас, больше не покажем".
		return c.JSON(fiber.Map{
			"token":    plaintext,
			"metadata": row,
		})
	}
}

// revokeTokenHandler — DELETE /v1/me/tokens/:id.
func revokeTokenHandler(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if ScopeFromCtx(c) != "" {
			return httperr.Forbidden(c, "jwt_required", "tokens management requires JWT")
		}
		if s.Pool == nil {
			return httperr.Unavailable(c, "db_not_configured", "database not configured")
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
		ok, err := RevokeAPIToken(c.Context(), s.Pool, uid, tokenID)
		if err != nil {
			return httperr.Internal(c)
		}
		if !ok {
			return httperr.NotFound(c, "token_not_found", "token not found")
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}
