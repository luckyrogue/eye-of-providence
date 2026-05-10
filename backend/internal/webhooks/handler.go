package webhooks

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
)

// RegisterRoutes — /v1/me/webhooks CRUD. JWT-only (anti-bootstrap, как с
// api_tokens — webhook secret = elevation primitive).
func RegisterRoutes(app *fiber.App, svc *Service, jwtSecret string, pool *pgxpool.Pool) {
	g := app.Group("/v1/me/webhooks", auth.Middleware(jwtSecret, pool))
	g.Get("/", listHandler(svc))
	g.Post("/", createHandler(svc))
	g.Delete("/:id", deleteHandler(svc))
}

// authErr — sentinel-обёртка чтобы handler мог распознать "не auth-баг" vs
// "конкретный httperr response уже отправлен". Возвращаем код+detail.
type authErr struct {
	code   string
	detail string
	status int
}

func (e *authErr) Error() string { return e.detail }

func requireJWT(c *fiber.Ctx) (uuid.UUID, error) {
	if auth.ScopeFromCtx(c) != "" {
		return uuid.Nil, &authErr{status: fiber.StatusForbidden, code: "jwt_required", detail: "webhook management requires JWT"}
	}
	claims := auth.ClaimsFromCtx(c)
	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, &authErr{status: fiber.StatusUnauthorized, code: "invalid_subject", detail: "invalid subject"}
	}
	return uid, nil
}

// sendAuthErr — translate authErr → httperr response. Caller pattern:
// `if err != nil { return sendAuthErr(c, err) }`.
func sendAuthErr(c *fiber.Ctx, err error) error {
	var ae *authErr
	if errors.As(err, &ae) {
		return httperr.Send(c, httperr.ProblemDetails{Status: ae.status, Code: ae.code, Detail: ae.detail})
	}
	return httperr.Internal(c)
}

func listHandler(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, err := requireJWT(c)
		if err != nil {
			return sendAuthErr(c, err)
		}
		hooks, err := svc.List(c.Context(), uid)
		if err != nil {
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"webhooks": hooks})
	}
}

func createHandler(svc *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, err := requireJWT(c)
		if err != nil {
			return sendAuthErr(c, err)
		}
		var req struct {
			URL    string   `json:"url"`
			Events []string `json:"events"`
			Format string   `json:"format"`
		}
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		secret, hook, err := svc.Create(c.Context(), uid, req.URL, req.Events, req.Format)
		if err != nil {
			return httperr.BadRequest(c, "create_failed", err.Error())
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
			return sendAuthErr(c, err)
		}
		id, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return httperr.BadRequest(c, "invalid_id", "invalid webhook id")
		}
		ok, err := svc.Delete(c.Context(), uid, id)
		if err != nil {
			return httperr.Internal(c)
		}
		if !ok {
			return httperr.NotFound(c, "webhook_not_found", "webhook not found")
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}
