package publicapi

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/publicapi/publicapiapp"
	"github.com/eye-of-providence/backend/internal/store"
)

func RegisterRoutes(app *fiber.App, st store.EventStore, logger *zap.Logger, jwtSecret string, pool *pgxpool.Pool) {
	svc := newPublicAPIApp(st)
	g := app.Group("/v1/public",
		auth.Middleware(jwtSecret, pool),
		auth.RequireScope("read", "admin"),
	)

	g.Get("/events", eventsHandler(svc, logger))
	g.Get("/summary", summaryHandler(svc, logger))
	g.Get("/languages", languagesHandler(svc, logger))
	g.Get("/trend", trendHandler(svc, logger))
}

func daysParam(c *fiber.Ctx, fallback int) int {
	n, _ := strconv.Atoi(c.Query("days", strconv.Itoa(fallback)))
	if n <= 0 || n > 365 {
		return fallback
	}
	return n
}

func eventsHandler(svc *publicapiapp.Service, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		limit, _ := strconv.Atoi(c.Query("limit", "100"))
		if limit <= 0 || limit > 1000 {
			limit = 100
		}
		events, err := svc.ListRecent(c.Context(), claims.UserID, limit)
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

func summaryHandler(svc *publicapiapp.Service, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days := daysParam(c, 7)
		_, since := publicapiapp.WindowFromDays(days)
		agg, err := svc.Summary(c.Context(), claims.UserID, since)
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

func languagesHandler(svc *publicapiapp.Service, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days := daysParam(c, 30)
		_, since := publicapiapp.WindowFromDays(days)
		cells, err := svc.Languages(c.Context(), claims.UserID, since)
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

func trendHandler(svc *publicapiapp.Service, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		days := daysParam(c, 30)
		_, since := publicapiapp.WindowFromDays(days)
		since = since.Truncate(24 * time.Hour)
		tz := c.Query("tz", "UTC")
		points, err := svc.Trend(c.Context(), claims.UserID, since, tz)
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
