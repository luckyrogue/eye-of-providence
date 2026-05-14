package auth

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth/meapp"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/metrics"
	"github.com/eye-of-providence/backend/internal/store"
)

// MeService — endpoints для текущего пользователя:
//
//	GET    /v1/me                       — профиль (id, email, github_login, locale)
//	PATCH  /v1/me/locale                — обновить locale (ru | en | kk | es)
//	GET    /v1/me/onboarding-status     — состояние онбординга (для wizard'а)
//	POST   /v1/me/onboarding/dismiss    — пометить wizard завершённым/закрытым
//	DELETE /v1/me/data                  — стирает все события и отчёты пользователя
type MeService struct {
	JWTSecret  string
	Pool       *pgxpool.Pool // может быть nil (in-memory mode)
	EventStore store.EventStore
	Logger     *zap.Logger
}

func RegisterMeRoutes(app *fiber.App, s MeService) {
	g := app.Group("/v1/me", Middleware(s.JWTSecret, s.Pool))

	g.Get("/", func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		app := newMeAppService(s)
		out, err := app.GetProfile(c.Context(), meapp.SessionClaims{
			UserID: claims.UserID, Email: claims.Email, Provider: claims.Provider,
		})
		if err != nil {
			if errors.Is(err, meapp.ErrInvalidSubject) {
				return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
			}
			return httperr.Internal(c)
		}
		return c.JSON(out)
	})

	g.Get("/onboarding-status", onboardingStatusHandler(s))
	g.Post("/onboarding/dismiss", onboardingDismissHandler(s))
	g.Patch("/locale", func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		var req struct {
			Locale string `json:"locale"`
		}
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		app := newMeAppService(s)
		loc, err := app.PatchLocale(c.Context(), uid, req.Locale)
		if err != nil {
			if errors.Is(err, meapp.ErrUnsupportedLocale) {
				return httperr.BadRequest(c, "invalid_locale", "unsupported locale")
			}
			s.Logger.Warn("locale update failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"locale": loc})
	})

	// API tokens — для public API. Все CRUD требуют JWT (RequireScope не
	// применяется т.к. ScopeFromCtx вернёт "" для JWT — full access).
	// API token нельзя использовать для управления API tokens (anti-bootstrap).
	g.Get("/tokens", listTokensHandler(s))
	g.Post("/tokens", createTokenHandler(s))
	g.Delete("/tokens/:id", revokeTokenHandler(s))

	g.Delete("/data", func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)

		// 1. Стираем events в EventStore (если backend поддерживает)
		if d, ok := s.EventStore.(store.UserDeleter); ok {
			if err := d.DeleteUserData(c.Context(), claims.UserID); err != nil {
				s.Logger.Warn("delete events failed", zap.Error(err))
			}
		}

		// 2. Стираем reports + user row в Postgres (если есть)
		if s.Pool != nil {
			uid, err := uuid.Parse(claims.UserID)
			if err == nil {
				if err := wipeUserPg(c.Context(), s.Pool, uid); err != nil {
					s.Logger.Warn("delete pg rows failed", zap.Error(err))
				}
			}
		}

		metrics.UsersDeleted.Inc()
		s.Logger.Info("user data deleted", zap.String("user", claims.UserID))
		return c.JSON(fiber.Map{"status": "ok", "deleted_user": claims.UserID})
	})
}

func wipeUserPg(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
	// ON DELETE CASCADE на reports/devices/projects/consent через FK,
	// но удалим явно для надёжности.
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, q := range []string{
		"DELETE FROM reports WHERE user_id = $1",
		"DELETE FROM api_tokens WHERE user_id = $1",
		"DELETE FROM consent WHERE user_id = $1",
		"DELETE FROM projects WHERE user_id = $1",
		"DELETE FROM devices WHERE user_id = $1",
		"DELETE FROM users WHERE id = $1",
	} {
		if _, err := tx.Exec(ctx, q, userID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
