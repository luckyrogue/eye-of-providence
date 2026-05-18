package push

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
)

type SvcConfig struct {
	*Service
	JWTSecret string
}

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
			return httperr.Unavailable(c, "push_not_configured", "push not configured")
		}
		return c.JSON(fiber.Map{"key": s.VAPIDPublicKey, "subject": s.VAPIDSubject})
	}
}

func listHandler(s *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, err := userID(c)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid subject")
		}
		subs, err := newPushApp(s).List(c.Context(), uid)
		if err != nil {
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"subscriptions": subs})
	}
}

func subscribeHandler(s *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, err := userID(c)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid subject")
		}
		var req struct {
			Endpoint string `json:"endpoint"`
			Keys     struct {
				P256dh string `json:"p256dh"`
				Auth   string `json:"auth"`
			} `json:"keys"`
		}
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		ua := c.Get("User-Agent")
		if len(ua) > 256 {
			ua = ua[:256]
		}
		if err := newPushApp(s).Subscribe(c.Context(), uid, req.Endpoint, req.Keys.P256dh, req.Keys.Auth, ua); err != nil {
			return httperr.BadRequest(c, "subscribe_failed", err.Error())
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}

func unsubscribeHandler(s *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, err := userID(c)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid subject")
		}
		var req struct {
			Endpoint string `json:"endpoint"`
		}
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		ok, err := newPushApp(s).Unsubscribe(c.Context(), uid, req.Endpoint)
		if err != nil {
			return httperr.Internal(c)
		}
		if !ok {
			return httperr.NotFound(c, "subscription_not_found", "subscription not found")
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}

func userID(c *fiber.Ctx) (uuid.UUID, error) {
	claims := auth.ClaimsFromCtx(c)
	return uuid.Parse(claims.UserID)
}
