package push

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth"
)

// JWTSecret поле для middleware. Добавляется к Service field-set чтобы
// RegisterRoutes мог вызвать middleware без отдельного конструктора.
type SvcConfig struct {
	*Service
	JWTSecret string
}

// RegisterRoutes — /v1/me/push/* CRUD + public VAPID key endpoint.
func RegisterRoutes(app *fiber.App, c SvcConfig) {
	g := app.Group("/v1/me/push", auth.Middleware(c.JWTSecret, c.Pool))
	g.Get("/vapid-key", vapidKeyHandler(c.Service))
	g.Get("/subscriptions", listHandler(c.Service))
	g.Post("/subscribe", subscribeHandler(c.Service))
	g.Post("/unsubscribe", unsubscribeHandler(c.Service))
}

func vapidKeyHandler(s *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if s.VAPIDPublicKey == "" {
			return c.Status(503).JSON(fiber.Map{"error": "push not configured"})
		}
		return c.JSON(fiber.Map{"key": s.VAPIDPublicKey, "subject": s.VAPIDSubject})
	}
}

func listHandler(s *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, err := userID(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid subject"})
		}
		subs, err := s.List(c.Context(), uid)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "query failed"})
		}
		return c.JSON(fiber.Map{"subscriptions": subs})
	}
}

func subscribeHandler(s *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, err := userID(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid subject"})
		}
		var req struct {
			Endpoint string `json:"endpoint"`
			Keys     struct {
				P256dh string `json:"p256dh"`
				Auth   string `json:"auth"`
			} `json:"keys"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
		}
		ua := c.Get("User-Agent")
		if len(ua) > 256 {
			ua = ua[:256]
		}
		if err := s.Subscribe(c.Context(), uid, req.Endpoint, req.Keys.P256dh, req.Keys.Auth, ua); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}

func unsubscribeHandler(s *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, err := userID(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid subject"})
		}
		var req struct {
			Endpoint string `json:"endpoint"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
		}
		ok, err := s.Unsubscribe(c.Context(), uid, req.Endpoint)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "query failed"})
		}
		if !ok {
			return c.Status(404).JSON(fiber.Map{"error": "not found"})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}

func userID(c *fiber.Ctx) (uuid.UUID, error) {
	claims := auth.ClaimsFromCtx(c)
	return uuid.Parse(claims.UserID)
}
