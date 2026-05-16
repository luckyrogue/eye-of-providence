package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth/meapp"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/metrics"
	"github.com/eye-of-providence/backend/internal/store"
)

type MeService struct {
	JWTSecret  string
	Pool       *pgxpool.Pool
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

	g.Get("/tokens", listTokensHandler(s))
	g.Post("/tokens", createTokenHandler(s))
	g.Delete("/tokens/:id", revokeTokenHandler(s))

	// GDPR Article 20 — right to data portability. Возвращаем machine-readable
	// JSON со всеми данными пользователя: profile + projects + devices +
	// consent + reports + полная история events (cap ~200k, см. store/export.go).
	// API tokens возвращаются БЕЗ hashed_token (это сам токен, не должен
	// утечь даже владельцу) — только id/scope/created_at.
	g.Get("/export", func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		bundle, err := buildExportBundle(c.Context(), s, uid)
		if err != nil {
			s.Logger.Warn("export failed", zap.Error(err))
			return httperr.Internal(c)
		}
		filename := "eop-export-" + uid.String() + ".json"
		c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+filename+`"`)
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		// Cache-Control no-store — export содержит PII, не должен оседать
		// в промежуточных кэшах/CDN.
		c.Set(fiber.HeaderCacheControl, "no-store, max-age=0")
		s.Logger.Info("user data export",
			zap.String("user", claims.UserID),
			zap.Int("events", len(bundle.Events)),
		)
		return c.JSON(bundle)
	})

	g.Delete("/data", func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)

		if d, ok := s.EventStore.(store.UserDeleter); ok {
			if err := d.DeleteUserData(c.Context(), claims.UserID); err != nil {
				s.Logger.Warn("delete events failed", zap.Error(err))
			}
		}

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

// ExportBundle — содержимое GET /v1/me/export. Все поля JSON-serializable,
// поля, которых нет у пользователя, отдаются как `null` или `[]` (не
// опускаются), чтобы downstream-парсеры не падали на отсутствующих ключах.
type ExportBundle struct {
	GeneratedAt string          `json:"generated_at"`
	Schema      string          `json:"schema"`
	UserID      string          `json:"user_id"`
	Profile     map[string]any  `json:"profile"`
	Devices     []map[string]any `json:"devices"`
	Projects    []map[string]any `json:"projects"`
	Consent     []map[string]any `json:"consent"`
	Reports     []map[string]any `json:"reports"`
	APITokens   []map[string]any `json:"api_tokens"` // БЕЗ hashed_token!
	Events      []store.Event    `json:"events"`
}

func buildExportBundle(ctx context.Context, s MeService, userID uuid.UUID) (*ExportBundle, error) {
	out := &ExportBundle{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Schema:      "eop.export.v1",
		UserID:      userID.String(),
		Devices:     []map[string]any{},
		Projects:    []map[string]any{},
		Consent:     []map[string]any{},
		Reports:     []map[string]any{},
		APITokens:   []map[string]any{},
		Events:      []store.Event{},
	}
	if s.Pool != nil {
		if err := loadProfile(ctx, s.Pool, userID, out); err != nil {
			return nil, err
		}
		if err := loadPgList(ctx, s.Pool, userID,
			`SELECT id, os, hostname, fingerprint, last_seen FROM devices WHERE user_id = $1 ORDER BY last_seen DESC NULLS LAST`,
			[]string{"id", "os", "hostname", "fingerprint", "last_seen"},
			&out.Devices); err != nil {
			return nil, err
		}
		if err := loadPgList(ctx, s.Pool, userID,
			`SELECT id, repo_url, root_path_hash, lang_primary FROM projects WHERE user_id = $1`,
			[]string{"id", "repo_url", "root_path_hash", "lang_primary"},
			&out.Projects); err != nil {
			return nil, err
		}
		if err := loadPgList(ctx, s.Pool, userID,
			`SELECT scope, granted_at, revoked_at FROM consent WHERE user_id = $1`,
			[]string{"scope", "granted_at", "revoked_at"},
			&out.Consent); err != nil {
			return nil, err
		}
		if err := loadPgList(ctx, s.Pool, userID,
			`SELECT id, period, format, body_md, generated_at, model FROM reports WHERE user_id = $1 ORDER BY generated_at DESC`,
			[]string{"id", "period", "format", "body_md", "generated_at", "model"},
			&out.Reports); err != nil {
			return nil, err
		}
		// api_tokens — ВАЖНО: НЕ возвращаем `hashed_token`. Это shared
		// secret backend'а; даже владельцу аккаунта он бесполезен (только
		// сам токен в незахешированном виде имеет смысл, а он не хранится).
		if err := loadPgList(ctx, s.Pool, userID,
			`SELECT id, scope, expires_at, created_at FROM api_tokens WHERE user_id = $1 ORDER BY created_at DESC`,
			[]string{"id", "scope", "expires_at", "created_at"},
			&out.APITokens); err != nil {
			return nil, err
		}
	}
	if ex, ok := s.EventStore.(store.UserExporter); ok {
		events, err := ex.ExportUserEvents(ctx, userID.String())
		if err != nil {
			return nil, err
		}
		out.Events = events
	}
	return out, nil
}

func loadProfile(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, out *ExportBundle) error {
	row := pool.QueryRow(ctx,
		`SELECT email, name, github_login, role, team_id, locale, created_at
		 FROM users WHERE id = $1`,
		userID,
	)
	var (
		email, name, githubLogin, role, locale string
		teamID                                  *uuid.UUID
		createdAt                               time.Time
	)
	if err := row.Scan(&email, &name, &githubLogin, &role, &teamID, &locale, &createdAt); err != nil {
		// Профиль может отсутствовать только если уже удалён — отдаём пустой объект.
		out.Profile = map[string]any{}
		return nil
	}
	p := map[string]any{
		"email":        email,
		"name":         name,
		"github_login": githubLogin,
		"role":         role,
		"locale":       locale,
		"created_at":   createdAt,
	}
	if teamID != nil {
		p["team_id"] = teamID.String()
	} else {
		p["team_id"] = nil
	}
	out.Profile = p
	return nil
}

// loadPgList — generic helper: один запрос с одной USER_ID-связанной WHERE,
// все колонки трактуются как `any`, пакуются в map с известными ключами.
// Подходит для plain-select таблиц без вложенной структуры; реляционные
// данные (если бы понадобились) лучше делать отдельной функцией.
func loadPgList(
	ctx context.Context,
	pool *pgxpool.Pool,
	userID uuid.UUID,
	query string,
	columns []string,
	dst *[]map[string]any,
) error {
	rows, err := pool.Query(ctx, query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return err
		}
		if len(vals) != len(columns) {
			return fmt.Errorf("export: column count mismatch (%d vs %d)", len(vals), len(columns))
		}
		row := make(map[string]any, len(columns))
		for i, col := range columns {
			row[col] = vals[i]
		}
		*dst = append(*dst, row)
	}
	return rows.Err()
}

func wipeUserPg(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {

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
