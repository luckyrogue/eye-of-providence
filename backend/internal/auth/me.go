package auth

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/metrics"
	"github.com/eye-of-providence/backend/internal/store"
)

// MeService — endpoints для текущего пользователя:
//   GET    /v1/me              — профиль (id, email, github_login)
//   DELETE /v1/me/data         — стирает все события и отчёты пользователя
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
		out := fiber.Map{
			"user_id":  claims.UserID,
			"email":    claims.Email,
			"provider": claims.Provider,
		}
		if s.Pool != nil {
			uid, err := uuid.Parse(claims.UserID)
			if err == nil {
				var ghLogin *string
				var globalRole *string
				var displayName *string
				_ = s.Pool.QueryRow(c.Context(),
					"SELECT github_login, global_role, display_name FROM users WHERE id = $1", uid).
					Scan(&ghLogin, &globalRole, &displayName)
				if ghLogin != nil {
					out["github_login"] = *ghLogin
				}
				if globalRole != nil {
					out["global_role"] = *globalRole
				}
				if displayName != nil {
					out["display_name"] = *displayName
				}
			}
		}
		return c.JSON(out)
	})

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
