package insights

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/insights/insightsapp"
	"github.com/eye-of-providence/backend/internal/store"
)

func RegisterRoutes(app *fiber.App, st store.EventStore, logger *zap.Logger, jwtSecret string, pool *pgxpool.Pool) {
	svc := newInsightsApp(st)
	g := app.Group("/v1/me", auth.Middleware(jwtSecret, pool))
	g.Get("/insights", insightsHandler(svc, logger))
}

func insightsHandler(svc *insightsapp.Service, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		tz := c.Query("tz", "UTC")
		out, err := svc.Generate(c.Context(), claims.UserID, tz, time.Now().UTC())
		if err != nil {
			logger.Error("insights fan-out failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"insights": out})
	}
}
