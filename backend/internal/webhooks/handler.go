package webhooks

import (
	"context"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/plans"
)

// effectiveUserPlan — возвращает "лучший" план среди команд, в которых юзер
// состоит. Используется для webhook plan-gate (webhooks per-user, plan
// per-team — несогласованный scope, см. handler.go createHandler TODO).
//
// Иерархия: enterprise > business > pro > free.
func effectiveUserPlan(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) string {
	rows, err := pool.Query(ctx, `
		SELECT t.subscription_plan
		FROM teams t
		JOIN team_members tm ON tm.team_id = t.id
		WHERE tm.user_id = $1`, userID)
	if err != nil {
		return plans.PlanFree
	}
	defer rows.Close()
	rank := map[string]int{plans.PlanFree: 0, plans.PlanPro: 1, plans.PlanBusiness: 2, plans.PlanEnterprise: 3}
	best, bestRank := plans.PlanFree, -1
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			continue
		}
		if r, ok := rank[p]; ok && r > bestRank {
			best, bestRank = p, r
		}
	}
	return best
}

// itoa — локальный alias чтобы не тащить strconv в каждый callsite.
func itoa(n int) string { return strconv.Itoa(n) }

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
		hooks, err := newWebhookListService(svc).List(c.Context(), uid)
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
		// Plan gate: webhooks per-user count vs Limits.MaxWebhooks. План берём
		// от "лучшей" команды юзера (см. effectiveUserPlan); при Enforce=false
		// возвращается Enterprise (limit=0, gate skip'ается).
		// TODO: на GA — определиться: webhook-per-team scope или per-user.
		// Сейчас per-user с эффективным планом — это допустимая интерпретация.
		if limits := svc.Plans.Limits(effectiveUserPlan(c.Context(), svc.Pool, uid)); limits.MaxWebhooks > 0 {
			var count int
			if err := svc.Pool.QueryRow(c.Context(),
				"SELECT count(*) FROM webhooks WHERE user_id=$1", uid).Scan(&count); err != nil {
				return httperr.Internal(c)
			}
			if count >= limits.MaxWebhooks {
				return httperr.Forbidden(c, "plan_limit_exceeded",
					"plan "+limits.Plan+" limited to "+itoa(limits.MaxWebhooks)+" webhook(s)")
			}
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
