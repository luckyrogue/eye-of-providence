package webhooks

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/auth"
)

// RegisterRoutes — /v1/me/webhooks CRUD. JWT-only (anti-bootstrap, как с
// api_tokens — webhook secret = elevation primitive).
func RegisterRoutes(app *fiber.App, svc *Service, jwtSecret string, pool *pgxpool.Pool) {
	g := app.Group("/v1/me/webhooks", auth.Middleware(jwtSecret, pool))
	g.Get("/", listHandler(svc))
	g.Post("/", createHandler(svc))
	g.Delete("/:id", deleteHandler(svc))
}

func requireJWT(c *fiber.Ctx) (uuid.UUID, error) {
	if auth.ScopeFromCtx(c) != "" {
		return uuid.Nil, fiber.NewError(fiber.StatusForbidden, "webhook management requires JWT")
	}
	claims := auth.ClaimsFromCtx(c)
	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "invalid subject")
	}
	return uid, nil
}

func listHandler(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, err := requireJWT(c)
		if err != nil {
			return err
		}
		hooks, err := svc.List(c.Context(), uid)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "query failed"})
		}
		return c.JSON(fiber.Map{"webhooks": hooks})
	}
}

func createHandler(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, err := requireJWT(c)
		if err != nil {
			return err
		}
		var req struct {
			URL    string   `json:"url"`
			Events []string `json:"events"`
			Format string   `json:"format"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
		}
		secret, hook, err := svc.Create(c.Context(), uid, req.URL, req.Events, req.Format)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{
			"secret":  secret,
			"webhook": hook,
		})
	}
}

func deleteHandler(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, err := requireJWT(c)
		if err != nil {
			return err
		}
		id, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
		}
		ok, err := svc.Delete(c.Context(), uid, id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "query failed"})
		}
		if !ok {
			return c.Status(404).JSON(fiber.Map{"error": "not found"})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}
