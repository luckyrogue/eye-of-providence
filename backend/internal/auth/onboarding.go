package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// onboardingStatus — то, что фронт показывает в wizard'е и что лежит
// в основе routing-решения "редиректить ли на /onboarding после login".
type onboardingStatus struct {
	TeamsCount int  `json:"teams_count"`
	HasEvent   bool `json:"has_event"`
	Dismissed  bool `json:"dismissed"`
}

func onboardingStatusHandler(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token subject"})
		}

		out := onboardingStatus{}

		if s.Pool != nil {
			// teams_count — сколько команд юзер участник.
			if err := s.Pool.QueryRow(c.Context(),
				`SELECT COUNT(*) FROM team_members WHERE user_id = $1`, uid,
			).Scan(&out.TeamsCount); err != nil {
				s.Logger.Warn("onboarding teams count failed", zap.Error(err))
			}

			// dismissed — флаг, что юзер прошёл/закрыл wizard.
			var dismissedAt *string
			if err := s.Pool.QueryRow(c.Context(),
				`SELECT onboarding_dismissed_at::text FROM users WHERE id = $1`, uid,
			).Scan(&dismissedAt); err != nil {
				s.Logger.Warn("onboarding dismissed read failed", zap.Error(err))
			}
			out.Dismissed = dismissedAt != nil
		}

		// has_event — пришло ли хоть одно событие. Дешевле всего: ListRecent(limit=1).
		if s.EventStore != nil {
			events, err := s.EventStore.ListRecent(c.Context(), claims.UserID, 1)
			if err != nil {
				s.Logger.Warn("onboarding event check failed", zap.Error(err))
			} else {
				out.HasEvent = len(events) > 0
			}
		}

		return c.JSON(out)
	}
}

// supportedLocales — синхронизирован с frontend dashboard/src/shared/i18n/index.ts.
var supportedLocales = map[string]bool{"ru": true, "en": true, "kk": true, "es": true}

func localeHandler(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token subject"})
		}
		var req struct {
			Locale string `json:"locale"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		if !supportedLocales[req.Locale] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "unsupported locale",
				"code":  "invalid_locale",
			})
		}
		if s.Pool == nil {
			return c.JSON(fiber.Map{"locale": req.Locale})
		}
		if _, err := s.Pool.Exec(c.Context(),
			`UPDATE users SET locale = $1 WHERE id = $2`, req.Locale, uid,
		); err != nil {
			s.Logger.Warn("locale update failed", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save"})
		}
		return c.JSON(fiber.Map{"locale": req.Locale})
	}
}

func onboardingDismissHandler(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token subject"})
		}
		if s.Pool == nil {
			// Без БД нечего сохранять; возвращаем ok, фронт переключит state локально.
			return c.JSON(fiber.Map{"status": "ok"})
		}
		// Идемпотентно: повторный dismiss оставляет первую дату.
		if _, err := s.Pool.Exec(c.Context(),
			`UPDATE users SET onboarding_dismissed_at = COALESCE(onboarding_dismissed_at, now()) WHERE id = $1`, uid,
		); err != nil {
			s.Logger.Warn("onboarding dismiss failed", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save"})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}
