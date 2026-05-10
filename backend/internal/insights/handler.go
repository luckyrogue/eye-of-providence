package insights

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
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

		// Parallel-сборка: 4 query, выполняем sequentially т.к. CH connection
		// pool serializes. Объединение в горутины не даст значимого выигрыша
		// при <5 запросах.
		aggLast, err := st.AggregateByCategory(c.Context(), claims.UserID, last7)
		if err != nil {
			logger.Error("agg last failed", zap.Error(err))
			return c.Status(500).JSON(fiber.Map{"error": "query failed"})
		}
		aggPrev, err := aggregateRange(c, st, claims.UserID, prev7, last7)
		if err != nil {
			logger.Error("agg prev failed", zap.Error(err))
			return c.Status(500).JSON(fiber.Map{"error": "query failed"})
		}
		langs, err := st.LanguageBreakdown(c.Context(), claims.UserID, last30)
		if err != nil {
			logger.Error("langs failed", zap.Error(err))
			return c.Status(500).JSON(fiber.Map{"error": "query failed"})
		}
		trend, err := st.DailyTrend(c.Context(), claims.UserID, last7, tz)
		if err != nil {
			logger.Error("trend failed", zap.Error(err))
			return c.Status(500).JSON(fiber.Map{"error": "query failed"})
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

// aggregateRange — AggregateByCategory с верхней границей (для prev7d).
// EventStore не имеет range-варианта, поэтому вычисляем как diff:
// agg(prev7) = agg(since=prev7) - agg(since=last7).
func aggregateRange(c *fiber.Ctx, st store.EventStore, userID string, since, until time.Time) (map[string]uint64, error) {
	full, err := st.AggregateByCategory(c.Context(), userID, since)
	if err != nil {
		return nil, err
	}
	tail, err := st.AggregateByCategory(c.Context(), userID, until)
	if err != nil {
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
