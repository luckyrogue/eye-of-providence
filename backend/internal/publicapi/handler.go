// Package publicapi — стабильный read-only data export для интеграций (BI,
// custom dashboards, exporters). Аутентификация через API tokens (eop_*) или
// JWT, scope=read достаточен.
//
// Endpoints (все под /v1/public, версия в URL для будущей миграции):
//
//	GET /v1/public/events       — paginated event log
//	GET /v1/public/summary      — agg by category за период
//	GET /v1/public/languages    — language breakdown
//	GET /v1/public/trend        — daily trend
//
// Стабильность: schema полей не меняется без увеличения версии в URL.
// Backward compatibility — обязательная (intern integrations будут опираться).
package publicapi

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/store"
)

// RegisterRoutes — public API под /v1/public с auth middleware + scope check.
func RegisterRoutes(app *fiber.App, st store.EventStore, logger *zap.Logger, jwtSecret string, pool *pgxpool.Pool) {
	g := app.Group("/v1/public",
		auth.Middleware(jwtSecret, pool),
		auth.RequireScope("read", "admin"),
	)

	g.Get("/events", eventsHandler(st, logger))
	g.Get("/summary", summaryHandler(st, logger))
	g.Get("/languages", languagesHandler(st, logger))
	g.Get("/trend", trendHandler(st, logger))
}

// daysParam — единый парсер ?days=N с safe-defaults и cap 365.
func daysParam(c *fiber.Ctx, fallback int) int {
	n, _ := strconv.Atoi(c.Query("days", strconv.Itoa(fallback)))
	if n <= 0 || n > 365 {
		return fallback
	}
	return n
}

func eventsHandler(st store.EventStore, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		limit, _ := strconv.Atoi(c.Query("limit", "100"))
		if limit <= 0 || limit > 1000 {
			limit = 100
		}
		events, err := newEventsApp(st).ListRecent(c.Context(), claims.UserID, limit)
		if err != nil {
			logger.Error("public events failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{
			"events": events,
			"count":  len(events),
			"limit":  limit,
		})
	}
}

func summaryHandler(st store.EventStore, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days := daysParam(c, 7)
		since := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
		agg, err := st.AggregateByCategory(c.Context(), claims.UserID, since)
		if err != nil {
			logger.Error("public summary failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{
			"days":       days,
			"since":      since,
			"categories": agg,
		})
	}
}

func languagesHandler(st store.EventStore, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days := daysParam(c, 30)
		since := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
		cells, err := st.LanguageBreakdown(c.Context(), claims.UserID, since)
		if err != nil {
			logger.Error("public langs failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{
			"days":  days,
			"since": since,
			"cells": cells,
		})
	}
}

func trendHandler(st store.EventStore, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days := daysParam(c, 30)
		since := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour).Truncate(24 * time.Hour)
		tz := c.Query("tz", "UTC")
		points, err := st.DailyTrend(c.Context(), claims.UserID, since, tz)
		if err != nil {
			logger.Error("public trend failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{
			"days":   days,
			"tz":     tz,
			"points": points,
		})
	}
}
