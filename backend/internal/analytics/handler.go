package analytics

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/analytics/analyticsapp"
	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/store"
)

func RegisterRoutes(app *fiber.App, st store.EventStore, logger *zap.Logger, jwtSecret string, pool *pgxpool.Pool) {
	svc := newAnalyticsApp(st)
	g := app.Group("/v1", auth.Middleware(jwtSecret, pool))

	g.Get("/events/recent", recentHandler(svc, logger))
	g.Get("/summary/languages", languagesHandler(svc, logger))
	g.Get("/trend", trendHandler(svc, logger))
	g.Get("/heatmap", heatmapHandler(svc, logger))
	g.Get("/summary/categories", categoriesHandler(svc, logger))
}

func daysParam(c *fiber.Ctx, fallback int) int {
	n, _ := strconv.Atoi(c.Query("days", strconv.Itoa(fallback)))
	if n <= 0 || n > 365 {
		return fallback
	}
	return n
}

func recentHandler(svc *analyticsapp.Service, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		limit, _ := strconv.Atoi(c.Query("limit", "50"))
		if limit <= 0 || limit > 500 {
			limit = 50
		}
		events, err := svc.ListRecent(c.Context(), claims.UserID, limit)
		if err != nil {
			logger.Error("list recent failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"events": events, "count": len(events)})
	}
}

func languagesHandler(svc *analyticsapp.Service, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days := daysParam(c, 30)
		w := analyticsapp.WindowFromDays(days)
		cells, err := svc.LanguageBreakdown(c.Context(), claims.UserID, w)
		if err != nil {
			logger.Error("language breakdown failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"days": days, "cells": cells})
	}
}

func trendHandler(svc *analyticsapp.Service, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days := daysParam(c, 30)
		q := analyticsapp.TrendWindowFromDays(days)
		q.TZ = c.Query("tz", "UTC")
		points, err := svc.DailyTrend(c.Context(), claims.UserID, q)
		if err != nil {
			logger.Error("trend failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"days": days, "points": points})
	}
}

func heatmapHandler(svc *analyticsapp.Service, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days := daysParam(c, 30)
		w := analyticsapp.WindowFromDays(days)
		tz := c.Query("tz", "UTC")
		cells, err := svc.Heatmap(c.Context(), claims.UserID, w, tz)
		if err != nil {
			logger.Error("heatmap failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"days": days, "cells": cells})
	}
}

func categoriesHandler(svc *analyticsapp.Service, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days := daysParam(c, 7)
		w := analyticsapp.WindowFromDays(days)
		agg, err := svc.Categories(c.Context(), claims.UserID, w)
		if err != nil {
			logger.Error("aggregate failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"days": days, "categories": agg})
	}
}
