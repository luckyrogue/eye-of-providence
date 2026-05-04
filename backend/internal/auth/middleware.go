package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

const ctxClaimsKey = "eop.claims"

// Middleware — проверяет Bearer JWT и кладёт Claims в context.
func Middleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing bearer token"})
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := ParseJWT(secret, token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}
		c.Locals(ctxClaimsKey, claims)
		return c.Next()
	}
}

func ClaimsFromCtx(c *fiber.Ctx) *Claims {
	v, _ := c.Locals(ctxClaimsKey).(*Claims)
	return v
}
