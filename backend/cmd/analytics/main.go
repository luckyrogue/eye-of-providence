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

	app := fiber.New(fiber.Config{AppName: "eop-analytics"})

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "analytics"})
	})

	// TODO Phase 1+: ClickHouse queries для дашборда (heatmap, AI ratio, top providers).

	logger.Info("analytics starting", zap.String("addr", cfg.HTTPAddr))
	if err := app.Listen(cfg.HTTPAddr); err != nil {
		logger.Fatal("analytics exited", zap.Error(err))
	}
}
