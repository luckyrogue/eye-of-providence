// cmd/api — единый dev-сервер. В production будет splitting на отдельные binaries
// (cmd/auth, cmd/ingest, cmd/analytics, cmd/reports), но в Phase 1 всё в одном
// процессе с in-memory store, чтобы поднималось без Docker.
package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/analytics"
	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/config"
	"github.com/eye-of-providence/backend/internal/ingest"
	eoplog "github.com/eye-of-providence/backend/internal/log"
	"github.com/eye-of-providence/backend/internal/store"
)

func main() {
	cfg := config.FromEnv()
	log := eoplog.New(cfg.Env)
	defer func() { _ = log.Sync() }()

	// Phase 1: in-memory store. ClickHouse adapter — Phase 2.
	st := store.NewMemory()
	defer st.Close()

	app := fiber.New(fiber.Config{
		AppName:               "eop-api",
		DisableStartupMessage: cfg.Env == "production",
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173,http://localhost:5174",
		AllowMethods:     "GET,POST,OPTIONS",
		AllowHeaders:     "Authorization,Content-Type",
		AllowCredentials: true,
	}))
	app.Use(logger.New(logger.Config{Format: "[${time}] ${status} ${method} ${path} ${latency}\n"}))

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "api"})
	})

	auth.RegisterRoutes(app, auth.Service{
		JWTSecret: cfg.JWTSecret,
		GitHub:    auth.NewGitHubOAuth(cfg.GitHubClientID, cfg.GitHubClientSec, "http://localhost:8080/v1/auth/github/callback"),
		Logger:    log,
	})
	ingest.RegisterRoutes(app, st, log, cfg.JWTSecret)
	analytics.RegisterRoutes(app, st, log, cfg.JWTSecret)

	log.Info("api starting", zap.String("addr", cfg.HTTPAddr), zap.String("env", cfg.Env))
	if err := app.Listen(cfg.HTTPAddr); err != nil {
		log.Fatal("api exited", zap.Error(err))
	}
}
