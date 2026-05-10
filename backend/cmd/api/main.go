// cmd/api — единый production-сервер. Подключается к Postgres + ClickHouse если они доступны,
// иначе fallback на in-memory (удобно для unit-тестов и quick demo без Docker).
package main

import (
	"context"
	"errors"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/analytics"
	"github.com/eye-of-providence/backend/internal/anomaly"
	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/cache"
	"github.com/eye-of-providence/backend/internal/config"
	"github.com/eye-of-providence/backend/internal/ingest"
	"github.com/eye-of-providence/backend/internal/insights"
	eoplog "github.com/eye-of-providence/backend/internal/log"
	"github.com/eye-of-providence/backend/internal/prcomment"
	"github.com/eye-of-providence/backend/internal/publicapi"
	"github.com/eye-of-providence/backend/internal/push"
	"github.com/eye-of-providence/backend/internal/mailer"
	"github.com/eye-of-providence/backend/internal/metrics"
	"github.com/eye-of-providence/backend/internal/migrate"
	"github.com/eye-of-providence/backend/internal/reports"
	"github.com/eye-of-providence/backend/internal/store"
	"github.com/eye-of-providence/backend/internal/teams"
	"github.com/eye-of-providence/backend/internal/webhooks"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--healthcheck" {
		runHealthcheck()
		return
	}

	cfg := config.FromEnv()
	log := eoplog.New(cfg.Env)
	defer func() { _ = log.Sync() }()

	if err := cfg.Validate(); err != nil {
		log.Fatal("invalid configuration", zap.Error(err))
	}

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgPool := openPgPool(cfg, log)
	defer func() {
		if pgPool != nil {
			pgPool.Close()
		}
	}()

	if cfg.AutoMigrate {
		mctx, mcancel := context.WithTimeout(rootCtx, 60*time.Second)
		if err := migrate.RunPostgres(mctx, cfg.PostgresDSN); err != nil {
			mcancel()
			log.Fatal("postgres migrate failed", zap.Error(err))
		}
		if err := migrate.RunClickHouse(mctx, cfg.ClickHouseDSN); err != nil {
			mcancel()
			log.Fatal("clickhouse migrate failed", zap.Error(err))
		}
		mcancel()
		log.Info("migrations applied")
	}

	eventStore := chooseEventStore(cfg, log)
	defer eventStore.Close()

	// Cache layer — оборачиваем CH-store в Redis-cached decorator если Redis
	// reachable. На failure (Redis down) — продолжаем без cache (degraded
	// performance, но functional).
	if cfg.RedisAddr != "" {
		cacheCtx, ccancel := context.WithTimeout(rootCtx, 3*time.Second)
		c, err := cache.New(cacheCtx, cfg.RedisAddr)
		ccancel()
		if err != nil {
			log.Warn("cache: redis unreachable, continuing without cache", zap.Error(err))
		} else {
			log.Info("cache: redis ready, wrapping event store")
			eventStore = store.NewCached(eventStore, c, log)
			defer c.Close()
		}
	}

	reportStore := chooseReportStore(log, pgPool)

	app := fiber.New(fiber.Config{
		AppName:               "eop-api",
		DisableStartupMessage: cfg.Env == "production",
		BodyLimit:             cfg.BodyLimitBytes,
		ReadTimeout:           15 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           120 * time.Second,
		ErrorHandler:          fiberErrorHandler(log),
	})

	app.Use(fiberrecover.New(fiberrecover.Config{EnableStackTrace: cfg.Env != "production"}))
	app.Use(requestid.New())
	// Request-latency histogram для всех HTTP-запросов. Не различаем route'ы
	// per-prom-label (overhead на UA-string'и), хватает агрегата по всему API.
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		metrics.RequestLatency.ObserveSince(start)
		return err
	})

	// Если AllowedOrigins=="*" — браузеры запрещают AllowCredentials=true с wildcard.
	allowCreds := cfg.AllowedOrigins != "*"
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     "GET,POST,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Authorization,Content-Type,X-Request-Id",
		AllowCredentials: allowCreds,
	}))
	app.Use(logger.New(logger.Config{
		Format: "[${time}] rid=${locals:requestid} ${status} ${method} ${path} ${latency}\n",
	}))

	app.Get("/healthz", healthzHandler(pgPool, eventStore))
	app.Get("/metrics", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/plain; version=0.0.4")
		return c.SendString(metrics.Render())
	})
	app.Get("/v1/admin/cost", auth.Middleware(cfg.JWTSecret, pgPool), func(c *fiber.Ctx) error {
		return c.JSON(metrics.Snapshot())
	})

	// Rate-limit на чувствительные auth endpoints (10 req / min / IP).
	authLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many requests"})
		},
	})
	app.Use("/v1/auth/login", authLimiter)
	app.Use("/v1/auth/register", authLimiter)
	app.Use("/v1/auth/dev-token", authLimiter)
	app.Use("/v1/auth/github/callback", authLimiter)

	// Глобальный rate-limit на authed endpoints. Ключ = JWT-юзер если есть, иначе IP.
	// 120 req/min нормальной нагрузки достаточно для UI; защищает /v1/teams POST,
	// /v1/admin/*, /v1/commits, /v1/ingest от наглого abuse'а.
	app.Use("/v1/", limiter.New(limiter.Config{
		Max:        120,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			h := c.Get("Authorization")
			if h != "" {
				return "auth:" + h
			}
			return "ip:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many requests"})
		},
	}))

	// Public routes регистрируются ПЕРВЫМИ. Fiber `app.Group(prefix, handler)`
	// работает как `app.Use(prefix, handler)` — middleware ловит все
	// последующие маршруты под префиксом. Если ingest/analytics зарегистрировать
	// до teams, они навесят auth на весь /v1, и /v1/auth/register перестанет
	// быть public.
	auth.RegisterRoutes(app, auth.Service{
		JWTSecret:      cfg.JWTSecret,
		GitHub:         auth.NewGitHubOAuth(cfg.GitHubClientID, cfg.GitHubClientSec, cfg.GitHubCallback),
		Logger:         log,
		Users:          auth.NewUsersPG(pgPool),
		Pool:           pgPool,
		EnableDevToken: cfg.EnableDevToken,
	})

	mail := chooseMailer(cfg, log)
	// hookSvc — concrete *webhooks.Service. Удовлетворяет двум интерфейсам:
	// teams.WebhookDispatcher и anomaly.Dispatcher (одна Dispatch-сигнатура).
	// Если pgPool == nil (in-memory), hookSvc остаётся nil → consumers
	// видят nil-interface и пропускают dispatch.
	var hookSvc *webhooks.Service
	var hooksDispatcher teams.WebhookDispatcher
	if pgPool != nil {
		hookSvc = webhooks.New(pgPool, log)
		webhooks.RegisterRoutes(app, hookSvc, cfg.JWTSecret, pgPool)
		hooksDispatcher = hookSvc
	}
	teams.EventStore = eventStore
	teams.RegisterRoutes(app, teams.Service{
		Pool:          pgPool,
		JWTSecret:     cfg.JWTSecret,
		Logger:        log,
		InviteOnly:    cfg.InviteOnly,
		BetaTeamLimit: cfg.BetaTeamLimit,
		Mailer:        mail,
		PublicURL:     cfg.PublicURL,
		Webhooks:      hooksDispatcher,
	})

	// Protected routes — навешивают auth middleware на весь /v1.
	auth.RegisterMeRoutes(app, auth.MeService{
		JWTSecret:  cfg.JWTSecret,
		Pool:       pgPool,
		EventStore: eventStore,
		Logger:     log,
	})
	ingest.RegisterRoutes(app, eventStore, log, cfg.JWTSecret, pgPool)
	analytics.RegisterRoutes(app, eventStore, log, cfg.JWTSecret, pgPool)
	insights.RegisterRoutes(app, eventStore, log, cfg.JWTSecret, pgPool)
	publicapi.RegisterRoutes(app, eventStore, log, cfg.JWTSecret, pgPool)
	if pgPool != nil {
		prcomment.RegisterRoutes(app, prcomment.Service{
			Pool:         pgPool,
			JWTSecret:    cfg.JWTSecret,
			Logger:       log,
			DashboardURL: cfg.PublicURL,
		})
	}

	// Web Push (PWA notifications) — endpoints всегда зарегистрированы; если
	// VAPID keys пусты, vapid-key handler вернёт 503 "not configured".
	var pushSvc *push.Service
	if pgPool != nil {
		pushSvc = &push.Service{
			Pool:            pgPool,
			Logger:          log,
			VAPIDPublicKey:  cfg.VAPIDPublicKey,
			VAPIDPrivateKey: cfg.VAPIDPrivateKey,
			VAPIDSubject:    cfg.VAPIDSubject,
		}
		push.RegisterRoutes(app, push.SvcConfig{Service: pushSvc, JWTSecret: cfg.JWTSecret})
	}
	_ = pushSvc // hooked в anomaly cron ниже

	gemini := reports.NewGeminiClient(cfg.GeminiAPIKey, "gemini-2.5-flash")
	reports.RegisterRoutes(app, reports.Service{
		Store:      reportStore,
		EventStore: eventStore,
		Gemini:     gemini,
		Logger:     log,
		Pool:       pgPool,
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
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Error("reports cron panicked", zap.Any("recover", r))
				}
			}()
			cron.Run(rootCtx)
		}()
		log.Info("reports cron started", zap.Int("interval_sec", cfg.ReportsCronSec))
	}

	// Anomaly detection cron — daily Z-score check, доставка через webhooks
	// + Web Push (если VAPID настроен). hookSvc и pushSvc оба satisfy
	// соответствующие interface'ы через идентичную Dispatch/SendToUser
	// сигнатуру.
	if hookSvc != nil {
		var pushOpt anomaly.PushSender
		if pushSvc != nil && cfg.VAPIDPublicKey != "" {
			pushOpt = pushSvc
		}
		anomalyCron := &anomaly.Cron{
			Interval:   24 * time.Hour,
			EventStore: eventStore,
			Webhooks:   hookSvc,
			Push:       pushOpt,
			Logger:     log,
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Error("anomaly cron panicked", zap.Any("recover", r))
				}
			}()
			anomalyCron.Run(rootCtx)
		}()
		log.Info("anomaly cron started", zap.Duration("interval", 24*time.Hour))
	}

	// SIGTERM/SIGINT → graceful shutdown.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Info("shutdown signal received", zap.String("signal", s.String()))
		cancel()
		if err := app.ShutdownWithTimeout(20 * time.Second); err != nil {
			log.Warn("graceful shutdown error", zap.Error(err))
		}
	}()

	log.Info("api starting",
		zap.String("addr", cfg.HTTPAddr),
		zap.String("env", cfg.Env),
		zap.Int("routes", len(app.GetRoutes())),
		zap.Bool("auto_migrate", cfg.AutoMigrate),
		zap.Bool("dev_token_enabled", cfg.EnableDevToken),
	)
	if err := app.Listen(cfg.HTTPAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("api exited", zap.Error(err))
	}
	log.Info("api stopped")
}

// fiberErrorHandler — единая точка обработки errors из handler'ов.
// Не отдаём raw err.Error() клиенту в production. Логируем полный текст.
func fiberErrorHandler(log *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		var fe *fiber.Error
		if errors.As(err, &fe) {
			code = fe.Code
		}
		rid, _ := c.Locals("requestid").(string)
		log.Warn("request failed",
			zap.Int("status", code),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("rid", rid),
			zap.Error(err),
		)
		// 4xx могут безопасно отдавать message; 5xx — generic.
		if code >= 500 {
			return c.Status(code).JSON(fiber.Map{
				"error":      "internal error",
				"request_id": rid,
			})
		}
		return c.Status(code).JSON(fiber.Map{
			"error":      err.Error(),
			"request_id": rid,
		})
	}
}

func healthzHandler(pool *pgxpool.Pool, ev store.EventStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		out := fiber.Map{"status": "ok", "service": "api"}
		degraded := false
		if pool != nil {
			if err := pool.Ping(ctx); err != nil {
				out["postgres"] = "down"
				degraded = true
			} else {
				out["postgres"] = "ok"
			}
		} else {
			out["postgres"] = "disabled"
		}
		if ch, ok := ev.(interface{ Ping(context.Context) error }); ok {
			if err := ch.Ping(ctx); err != nil {
				out["clickhouse"] = "down"
				degraded = true
			} else {
				out["clickhouse"] = "ok"
			}
		} else {
			out["clickhouse"] = "in-memory"
		}
		if degraded {
			out["status"] = "degraded"
			return c.Status(fiber.StatusServiceUnavailable).JSON(out)
		}
		return c.JSON(out)
	}
}

func openPgPool(cfg config.Config, log *zap.Logger) *pgxpool.Pool {
	if cfg.PostgresDSN == "" {
		return nil
	}
	pcfg, err := pgxpool.ParseConfig(cfg.PostgresDSN)
	if err != nil {
		log.Warn("postgres dsn parse failed", zap.Error(err))
		return nil
	}
	// Tuned defaults — переопределяемы через DSN-параметры pool_max_conns и т.п.
	if pcfg.MaxConns == 0 || pcfg.MaxConns == 4 {
		pcfg.MaxConns = envIntOr("EOP_PG_MAX_CONNS", 20)
	}
	pcfg.MinConns = envIntOr("EOP_PG_MIN_CONNS", 2)
	pcfg.MaxConnLifetime = 30 * time.Minute
	pcfg.MaxConnIdleTime = 5 * time.Minute
	pcfg.HealthCheckPeriod = 1 * time.Minute
	if pcfg.ConnConfig.RuntimeParams == nil {
		pcfg.ConnConfig.RuntimeParams = map[string]string{}
	}
	// Защита от слоу-квери, повисших на одном коннекте.
	if _, ok := pcfg.ConnConfig.RuntimeParams["statement_timeout"]; !ok {
		pcfg.ConnConfig.RuntimeParams["statement_timeout"] = "5000" // ms
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		log.Warn("postgres pool open failed", zap.Error(err))
		return nil
	}
	if err := pool.Ping(ctx); err != nil {
		log.Warn("postgres ping failed", zap.Error(err))
		pool.Close()
		return nil
	}
	log.Info("postgres pool ready", zap.Int32("max_conns", pcfg.MaxConns))
	return pool
}

func envIntOr(key string, fallback int32) int32 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 || n > math.MaxInt32 {
		return fallback
	}
	return int32(n)
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

func chooseMailer(cfg config.Config, log *zap.Logger) mailer.Mailer {
	if cfg.ResendAPIKey == "" {
		log.Info("mailer noop: EOP_RESEND_API_KEY empty (emails will only be logged)")
		return mailer.Noop(log)
	}
	log.Info("mailer ready", zap.String("from", cfg.MailFrom))
	return mailer.NewResend(cfg.ResendAPIKey, cfg.MailFrom, log)
}

func chooseReportStore(log *zap.Logger, pool *pgxpool.Pool) reports.ReportStore {
	if pool == nil {
		log.Info("no postgres pool, using in-memory report store")
		return reports.NewStore()
	}
	log.Info("postgres report store ready")
	return reports.NewPostgresStore(pool)
}

// runHealthcheck — отдельный mode для Docker HEALTHCHECK.
// Distroless image не имеет shell/curl, поэтому используем сам binary.
func runHealthcheck() {
	addr := os.Getenv("EOP_HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	url := "http://localhost" + addr + "/healthz"
	c := &http.Client{Timeout: 3 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	os.Exit(0)
}
