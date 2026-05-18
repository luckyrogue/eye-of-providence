package app

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/analytics"
	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/config"
	"github.com/eye-of-providence/backend/internal/content"
	"github.com/eye-of-providence/backend/internal/devices"
	"github.com/eye-of-providence/backend/internal/ingest"
	"github.com/eye-of-providence/backend/internal/insights"
	"github.com/eye-of-providence/backend/internal/mailer"
	"github.com/eye-of-providence/backend/internal/plans"
	"github.com/eye-of-providence/backend/internal/prcomment"
	"github.com/eye-of-providence/backend/internal/publicapi"
	"github.com/eye-of-providence/backend/internal/reports"
	"github.com/eye-of-providence/backend/internal/sso"
	"github.com/eye-of-providence/backend/internal/store"
	"github.com/eye-of-providence/backend/internal/teams"
	"github.com/eye-of-providence/backend/internal/webhooks"
)

// API wires product bounded contexts onto the Fiber app (composition root).
type API struct {
	App         *fiber.App
	Cfg         config.Config
	Log         *zap.Logger
	Pool        *pgxpool.Pool
	EventStore  store.EventStore
	ReportStore reports.ReportStore
	Gemini      *reports.GeminiClient
	Auth        auth.Service
	WebAuthn    *auth.WebAuthnService
	Mail        mailer.Mailer
	Plans       plans.Service
	Audit       audit.Service
	Webhooks    *webhooks.Service
	SSORegistry *sso.Registry
	Content     *content.Service
}

func (a *API) RegisterProductRoutes() {
	if a.Webhooks != nil && a.Pool != nil {
		webhooks.RegisterRoutes(a.App, a.Webhooks, a.Cfg.JWTSecret, a.Pool)
	}
	if a.Pool != nil && a.SSORegistry != nil {
		sso.RegisterRoutes(a.App, sso.Service{
			Pool: a.Pool, Registry: a.SSORegistry, Logger: a.Log,
			JWTSecret: a.Cfg.JWTSecret, PublicURL: strings.TrimRight(a.Cfg.PublicURL, "/"),
		})
	}
	a.Content.RegisterPublicRoute(a.App)
	// ВАЖНО: публичные routes (forgot/reset password) регистрируем
	// ДО teams.RegisterRoutes. Последний вешает auth.Middleware на `/v1`
	// через app.Group("/v1", mw) — fiber Use'ит middleware для всех
	// последующих routes на этом prefix'е. Public-routes должны быть
	// зарегистрированы раньше, иначе вернут 401 "missing_bearer".
	auth.RegisterPasswordResetRoutes(a.App, auth.PasswordResetService{
		Pool: a.Pool, Mailer: a.Mail, PublicURL: a.Cfg.PublicURL, Logger: a.Log,
	})
	teams.RegisterRoutes(a.App, teams.Service{
		Pool: a.Pool, JWTSecret: a.Cfg.JWTSecret, Logger: a.Log,
		InviteOnly: a.Cfg.InviteOnly, BetaTeamLimit: a.Cfg.BetaTeamLimit,
		Mailer: a.Mail, PublicURL: a.Cfg.PublicURL,
		Webhooks: hooksDispatcher(a.Webhooks), Plans: a.Plans, Audit: a.Audit,
		AuthProviders: a.Auth.ProvidersList(), PasskeyEnabled: a.WebAuthn != nil,
		TemplateStore: templateStore(a.Pool), EventStore: a.EventStore,
	})
	auth.RegisterIdentitiesRoutes(a.App, a.Auth)
	auth.RegisterMeRoutes(a.App, auth.MeService{
		JWTSecret: a.Cfg.JWTSecret, Pool: a.Pool, EventStore: a.EventStore, Logger: a.Log,
	})
	ingest.RegisterRoutes(a.App, a.EventStore, a.Log, a.Cfg.JWTSecret, a.Pool)
	analytics.RegisterRoutes(a.App, a.EventStore, a.Log, a.Cfg.JWTSecret, a.Pool)
	insights.RegisterRoutes(a.App, a.EventStore, a.Log, a.Cfg.JWTSecret, a.Pool)
	publicapi.RegisterRoutes(a.App, a.EventStore, a.Log, a.Cfg.JWTSecret, a.Pool)
	if a.Pool != nil {
		prcomment.RegisterRoutes(a.App, prcomment.Service{
			Pool: a.Pool, JWTSecret: a.Cfg.JWTSecret, Logger: a.Log, DashboardURL: a.Cfg.PublicURL,
		})
	}
	devices.RegisterRoutes(a.App, devices.Service{Pool: a.Pool, Logger: a.Log, JWTSecret: a.Cfg.JWTSecret})
	reports.RegisterRoutes(a.App, reports.Service{
		Store: a.ReportStore, EventStore: a.EventStore, Gemini: a.Gemini,
		Logger: a.Log, JWTSecret: a.Cfg.JWTSecret, Pool: a.Pool,
	})
}

func hooksDispatcher(h *webhooks.Service) teams.WebhookDispatcher {
	if h == nil {
		return nil
	}
	return h
}

func templateStore(pool *pgxpool.Pool) *mailer.PGTemplateStore {
	if pool == nil {
		return nil
	}
	return mailer.NewPGTemplateStore(pool)
}
