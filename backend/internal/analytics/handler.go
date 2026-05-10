package analytics

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

func RegisterRoutes(app *fiber.App, st store.EventStore, logger *zap.Logger, jwtSecret string, pool *pgxpool.Pool) {
	g := app.Group("/v1", auth.Middleware(jwtSecret, pool))

	g.Get("/events/recent", func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		limit, _ := strconv.Atoi(c.Query("limit", "50"))
		if limit <= 0 || limit > 500 {
			limit = 50
		}
		events, err := st.ListRecent(c.Context(), claims.UserID, limit)
		if err != nil {
			logger.Error("list recent failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"events": events, "count": len(events)})
	})

	g.Get("/summary/languages", func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days, _ := strconv.Atoi(c.Query("days", "30"))
		if days <= 0 || days > 365 {
			days = 30
		}
		since := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
		cells, err := st.LanguageBreakdown(c.Context(), claims.UserID, since)
		if err != nil {
			logger.Error("language breakdown failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"days": days, "cells": cells})
	})

	g.Get("/trend", func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days, _ := strconv.Atoi(c.Query("days", "30"))
		if days <= 0 || days > 365 {
			days = 30
		}
		since := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour).Truncate(24 * time.Hour)
		tz := c.Query("tz", "UTC")
		points, err := st.DailyTrend(c.Context(), claims.UserID, since, tz)
		if err != nil {
			logger.Error("trend failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"days": days, "points": points})
	})

	g.Get("/heatmap", func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days, _ := strconv.Atoi(c.Query("days", "30"))
		if days <= 0 || days > 365 {
			days = 30
		}
		since := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
		tz := c.Query("tz", "UTC")
		cells, err := st.Heatmap(c.Context(), claims.UserID, since, tz)
		if err != nil {
			logger.Error("heatmap failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"days": days, "cells": cells})
	})

	g.Get("/summary/categories", func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days, _ := strconv.Atoi(c.Query("days", "7"))
		if days <= 0 || days > 365 {
			days = 7
		}
		since := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
		agg, err := st.AggregateByCategory(c.Context(), claims.UserID, since)
		if err != nil {
			logger.Error("aggregate failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"days": days, "categories": agg})
	})
}
