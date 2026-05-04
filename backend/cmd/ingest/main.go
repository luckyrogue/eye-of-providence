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

	app := fiber.New(fiber.Config{
		AppName:               "eop-ingest",
		DisableStartupMessage: cfg.Env == "production",
	})

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "ingest"})
	})

	app.Post("/v1/ingest", func(c *fiber.Ctx) error {
		// TODO Phase 1: validate JWT, decode protobuf IngestRequest, batch-insert в ClickHouse.
		return c.SendStatus(fiber.StatusNotImplemented)
	})

	logger.Info("ingest starting", zap.String("addr", cfg.HTTPAddr))
	if err := app.Listen(cfg.HTTPAddr); err != nil {
		logger.Fatal("ingest exited", zap.Error(err))
	}
}
