package insights

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/store"
)

func RegisterRoutes(app *fiber.App, st store.EventStore, logger *zap.Logger, jwtSecret string, pool *pgxpool.Pool) {
	g := app.Group("/v1/me", auth.Middleware(jwtSecret, pool))
	g.Get("/insights", insightsHandler(st, logger))
}

func insightsHandler(st store.EventStore, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		now := time.Now().UTC()
		last7 := now.Add(-7 * 24 * time.Hour)
		prev7 := now.Add(-14 * 24 * time.Hour)
		last30 := now.Add(-30 * 24 * time.Hour)
		tz := c.Query("tz", "UTC")

		var (
			aggLast, aggPrev map[string]uint64
			langs            []store.LangCell
			trend            []store.TrendPoint
		)
		g, gctx := errgroup.WithContext(c.Context())
		g.Go(func() error {
			v, err := st.AggregateByCategory(gctx, claims.UserID, last7)
			if err != nil {
				return err
			}
			aggLast = v
			return nil
		})
		g.Go(func() error {
			v, err := aggregateRangeCtx(gctx, st, claims.UserID, prev7, last7)
			if err != nil {
				return err
			}
			aggPrev = v
			return nil
		})
		g.Go(func() error {
			v, err := st.LanguageBreakdown(gctx, claims.UserID, last30)
			if err != nil {
				return err
			}
			langs = v
			return nil
		})
		g.Go(func() error {
			v, err := st.DailyTrend(gctx, claims.UserID, last7, tz)
			if err != nil {
				return err
			}
			trend = v
			return nil
		})
		if err := g.Wait(); err != nil {
			logger.Error("insights fan-out failed", zap.Error(err))
			return httperr.Internal(c)
		}

		out := Generate(Inputs{
			Last7d:    aggLast,
			Prev7d:    aggPrev,
			Languages: langs,
			Trend:     trend,
		})
		return c.JSON(fiber.Map{"insights": out})
	}
}
