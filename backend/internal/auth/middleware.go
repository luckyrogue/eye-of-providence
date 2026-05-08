package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ctxClaimsKey = "eop.claims"

// Middleware — проверяет Bearer JWT и кладёт Claims в context.
// Если pool != nil, дополнительно сверяет `tv` claim с users.token_version
// в БД и отказывает если не совпадает — это revocation hook на демоут/wipe.
func Middleware(secret string, pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		// Case-insensitive Bearer — некоторые клиенты шлют lowercase.
		if len(header) < 7 || !strings.EqualFold(header[:7], "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing bearer token"})
		}
		token := strings.TrimSpace(header[7:])
		claims, err := ParseJWT(secret, token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}
		// JWT revocation: сверяем tv claim'а с актуальной версией в БД.
		// Если pool == nil (тесты или старт без БД) — пропускаем проверку.
		if pool != nil {
			uid, err := uuid.Parse(claims.UserID)
			if err == nil {
				dbTV, dbErr := TokenVersion(c.Context(), pool, uid)
				if dbErr != nil || dbTV != claims.TokenVersion {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "session invalidated"})
				}
			}
		}
		c.Locals(ctxClaimsKey, claims)
		return c.Next()
	}
}

func ClaimsFromCtx(c *fiber.Ctx) *Claims {
	v, _ := c.Locals(ctxClaimsKey).(*Claims)
	return v
}
