package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ctxClaimsKey = "eop.claims"
	// ctxTokenScopeKey — scope ("read"/"write:ingest"/"admin") если auth был
	// через API token; пусто для JWT (full-scope dashboard session).
	ctxTokenScopeKey = "eop.token_scope"
)

// Middleware — accepts либо JWT либо API token (eop_*).
//   - JWT: ParseJWT + tv-revocation check против users.token_version.
//   - API token: lookup в api_tokens + last_used_at touch. Scope попадает
//     в context, эндпоинты могут проверить через ScopeFromCtx.
//
// Если pool == nil, TV-check для JWT отключается; API tokens требуют pool
// (без БД мы не можем lookup и middleware вернёт 401).
func Middleware(secret string, pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		// Case-insensitive Bearer — некоторые клиенты шлют lowercase.
		if len(header) < 7 || !strings.EqualFold(header[:7], "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing bearer token"})
		}
		token := strings.TrimSpace(header[7:])

		// API token — eop_xxx.
		if strings.HasPrefix(token, tokenPrefix) {
			if pool == nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "api tokens require database"})
			}
			userID, scope, err := VerifyAPIToken(c.Context(), pool, token)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
			}
			c.Locals(ctxClaimsKey, &Claims{UserID: userID.String(), Provider: "api"})
			c.Locals(ctxTokenScopeKey, scope)
			return c.Next()
		}

		// JWT path.
		claims, err := ParseJWT(secret, token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}
		if pool != nil {
			uid, err := uuid.Parse(claims.UserID)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token subject"})
			}
			dbTV, dbErr := TokenVersion(c.Context(), pool, uid)
			if dbErr != nil || dbTV != claims.TokenVersion {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "session invalidated"})
			}
		}
		c.Locals(ctxClaimsKey, claims)
		return c.Next()
	}
}

// ScopeFromCtx — scope API token'а или "" если auth через JWT (full-scope).
func ScopeFromCtx(c *fiber.Ctx) string {
	v, _ := c.Locals(ctxTokenScopeKey).(string)
	return v
}

// RequireScope — guard для эндпоинтов: разрешить если auth через JWT
// (full-scope) ИЛИ API token имеет нужный scope. allowed — список scope'ов
// которые удовлетворяют требованию (например "read" допускает и "admin").
func RequireScope(allowed ...string) fiber.Handler {
	allowedMap := make(map[string]bool, len(allowed))
	for _, s := range allowed {
		allowedMap[s] = true
	}
	return func(c *fiber.Ctx) error {
		scope := ScopeFromCtx(c)
		if scope == "" {
			// JWT (dashboard session) — full access.
			return c.Next()
		}
		// admin scope — wildcard.
		if scope == "admin" || allowedMap[scope] {
			return c.Next()
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":          "token scope insufficient",
			"required_scope": allowed,
			"actual_scope":   scope,
		})
	}
}

func ClaimsFromCtx(c *fiber.Ctx) *Claims {
	v, _ := c.Locals(ctxClaimsKey).(*Claims)
	return v
}
