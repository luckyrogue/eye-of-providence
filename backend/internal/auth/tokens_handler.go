package auth

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// listTokensHandler — GET /v1/me/tokens.
// Anti-bootstrap: API token не может смотреть/создавать токены.
func listTokensHandler(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if ScopeFromCtx(c) != "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "tokens management requires JWT (dashboard session)",
			})
		}
		if s.Pool == nil {
			return c.JSON(fiber.Map{"tokens": []APIToken{}})
		}
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid subject"})
		}
		tokens, err := ListAPITokens(c.Context(), s.Pool, uid)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "query failed"})
		}
		return c.JSON(fiber.Map{"tokens": tokens})
	}
}

// createTokenHandler — POST /v1/me/tokens. Body: {name, scope, ttl_days}.
// Возвращает plaintext "token" ровно один раз.
func createTokenHandler(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if ScopeFromCtx(c) != "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "tokens management requires JWT",
			})
		}
		if s.Pool == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not configured"})
		}
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid subject"})
		}
		var req struct {
			Name    string `json:"name"`
			Scope   string `json:"scope"`
			TTLDays int    `json:"ttl_days"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		req.Name = strings.TrimSpace(req.Name)
		if len(req.Name) > 64 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name too long (max 64)"})
		}
		// Дефолт scope=read — самый безопасный для интеграций.
		if req.Scope == "" {
			req.Scope = "read"
		}
		// TTL: 0 — без истечения, max 365 дней (чтобы forced rotation хотя бы раз в год).
		if req.TTLDays < 0 || req.TTLDays > 365 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ttl_days must be 0..365"})
		}
		ttl := time.Duration(req.TTLDays) * 24 * time.Hour

		plaintext, row, err := CreateAPIToken(c.Context(), s.Pool, uid, req.Name, req.Scope, ttl)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
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
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "tokens management requires JWT",
			})
		}
		if s.Pool == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not configured"})
		}
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid subject"})
		}
		tokenID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid token id"})
		}
		ok, err := RevokeAPIToken(c.Context(), s.Pool, uid, tokenID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "query failed"})
		}
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "token not found"})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}
