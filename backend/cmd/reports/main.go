package main

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/config"
	"github.com/eye-of-providence/backend/internal/log"
)

func main() {
	cfg := config.FromEnv()
	logger := log.New(cfg.Env)
	defer func() { _ = logger.Sync() }()

	app := fiber.New(fiber.Config{AppName: "eop-reports"})

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "reports"})
	})

	logger.Info("reports starting", zap.String("addr", cfg.HTTPAddr), zap.Bool("gemini_configured", cfg.GeminiAPIKey != ""))
	if err := app.Listen(cfg.HTTPAddr); err != nil {
		logger.Fatal("reports exited", zap.Error(err))
	}
}
