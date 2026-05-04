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

	// TODO Phase 5: POST /v1/reports/generate — собрать numeric context из ClickHouse,
	// отправить в Gemini (gemini-2.5-flash) с context caching system prompt,
	// сохранить markdown в Postgres reports table.

	logger.Info("reports starting", zap.String("addr", cfg.HTTPAddr), zap.Bool("gemini_configured", cfg.GeminiAPIKey != ""))
	if err := app.Listen(cfg.HTTPAddr); err != nil {
		logger.Fatal("reports exited", zap.Error(err))
	}
}
