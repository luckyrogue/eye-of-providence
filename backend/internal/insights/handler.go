package insights

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/store"
)

// RegisterRoutes — GET /v1/me/insights → []Insight.
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

		// Fan-out: 4 independent queries в parallel. ClickHouse driver
		// поддерживает multiple connections (default pool=10), wall-clock
		// время = max(query_i) вместо sum. На typical CH Cloud это ~2-3×
		// speedup для insights endpoint.
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

// aggregateRangeCtx — AggregateByCategory с верхней границей (для prev7d).
// EventStore не имеет range-варианта, поэтому вычисляем как diff:
//
//	agg(prev7) = agg(since=prev7) - agg(since=last7)
//
// Внутри две CH queries — fan-out parallel чтобы wall-clock = max(2) вместо
// sum.
func aggregateRangeCtx(ctx context.Context, st store.EventStore, userID string, since, until time.Time) (map[string]uint64, error) {
	var full, tail map[string]uint64
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		v, err := st.AggregateByCategory(gctx, userID, since)
		if err != nil {
			return err
		}
		full = v
		return nil
	})
	g.Go(func() error {
		v, err := st.AggregateByCategory(gctx, userID, until)
		if err != nil {
			return err
		}
		tail = v
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	out := make(map[string]uint64, len(full))
	for k, v := range full {
		t := tail[k]
		if v >= t {
			out[k] = v - t
		}
	}
	return out, nil
}
