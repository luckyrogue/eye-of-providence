// cmd/api — единый dev-сервер. Подключается к Postgres + ClickHouse если они доступны,
// иначе fallback на in-memory (удобно для unit-тестов и quick demo без Docker).
package main

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/analytics"
	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/config"
	"github.com/eye-of-providence/backend/internal/ingest"
	eoplog "github.com/eye-of-providence/backend/internal/log"
	"github.com/eye-of-providence/backend/internal/metrics"
	"github.com/eye-of-providence/backend/internal/reports"
	"github.com/eye-of-providence/backend/internal/store"
	"github.com/eye-of-providence/backend/internal/teams"
)

func main() {
	cfg := config.FromEnv()
	log := eoplog.New(cfg.Env)
	defer func() { _ = log.Sync() }()

	pgPool := openPgPool(cfg, log)
	defer func() {
		if pgPool != nil {
			pgPool.Close()
		}
	}()

	eventStore := chooseEventStore(cfg, log)
	defer eventStore.Close()

	reportStore := chooseReportStore(log, pgPool)

	app := fiber.New(fiber.Config{
		AppName:               "eop-api",
		DisableStartupMessage: cfg.Env == "production",
	})
	// Если AllowedOrigins=="*" — браузеры запрещают AllowCredentials=true с wildcard.
	allowCreds := cfg.AllowedOrigins != "*"
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     "GET,POST,DELETE,OPTIONS",
		AllowHeaders:     "Authorization,Content-Type",
		AllowCredentials: allowCreds,
	}))
	app.Use(logger.New(logger.Config{Format: "[${time}] ${status} ${method} ${path} ${latency}\n"}))

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "api"})
	})
	app.Get("/metrics", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/plain; version=0.0.4")
		return c.SendString(metrics.Render())
	})
	app.Get("/v1/admin/cost", auth.Middleware(cfg.JWTSecret), func(c *fiber.Ctx) error {
		return c.JSON(metrics.Snapshot())
	})

	auth.RegisterRoutes(app, auth.Service{
		JWTSecret: cfg.JWTSecret,
		GitHub:    auth.NewGitHubOAuth(cfg.GitHubClientID, cfg.GitHubClientSec, "http://localhost:8080/v1/auth/github/callback"),
		Logger:    log,
		Users:     auth.NewUsersPG(pgPool),
	})
	auth.RegisterMeRoutes(app, auth.MeService{
		JWTSecret:  cfg.JWTSecret,
		Pool:       pgPool,
		EventStore: eventStore,
		Logger:     log,
	})
	ingest.RegisterRoutes(app, eventStore, log, cfg.JWTSecret)
	analytics.RegisterRoutes(app, eventStore, log, cfg.JWTSecret)

	// Teams + email/password auth + invites
	teams.EventStore = eventStore
	teams.RegisterRoutes(app, teams.Service{
		Pool:      pgPool,
		JWTSecret: cfg.JWTSecret,
		Logger:    log,
	})

	gemini := reports.NewGeminiClient(cfg.GeminiAPIKey, "gemini-2.5-flash")
	reports.RegisterRoutes(app, reports.Service{
		Store:      reportStore,
		EventStore: eventStore,
		Gemini:     gemini,
		Logger:     log,
		JWTSecret:  cfg.JWTSecret,
	})

	if cfg.ReportsCronSec > 0 {
		cron := &reports.Cron{
			Interval:   time.Duration(cfg.ReportsCronSec) * time.Second,
			Store:      reportStore,
			EventStore: eventStore,
			Gemini:     gemini,
			Logger:     log,
		}
		go cron.Run(context.Background())
		log.Info("reports cron started", zap.Int("interval_sec", cfg.ReportsCronSec))
	}

	log.Info("api starting", zap.String("addr", cfg.HTTPAddr), zap.String("env", cfg.Env), zap.Int("routes", len(app.GetRoutes())))
	if err := app.Listen(cfg.HTTPAddr); err != nil {
		log.Fatal("api exited", zap.Error(err))
	}
}

func openPgPool(cfg config.Config, log *zap.Logger) *pgxpool.Pool {
	if cfg.PostgresDSN == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Warn("postgres pool open failed", zap.Error(err))
		return nil
	}
	if err := pool.Ping(ctx); err != nil {
		log.Warn("postgres ping failed", zap.Error(err))
		pool.Close()
		return nil
	}
	log.Info("postgres pool ready")
	return pool
}

func chooseEventStore(cfg config.Config, log *zap.Logger) store.EventStore {
	if cfg.ClickHouseDSN == "" {
		log.Info("clickhouse dsn empty, using in-memory event store")
		return store.NewMemory()
	}
	ch, err := store.OpenClickHouse(cfg.ClickHouseDSN)
	if err != nil {
		log.Warn("clickhouse unavailable, falling back to in-memory", zap.Error(err))
		return store.NewMemory()
	}
	log.Info("clickhouse event store ready")
	return ch
}

func chooseReportStore(log *zap.Logger, pool *pgxpool.Pool) reports.ReportStore {
	if pool == nil {
		log.Info("no postgres pool, using in-memory report store")
		return reports.NewStore()
	}
	log.Info("postgres report store ready")
	return reports.NewPostgresStore(pool)
}
