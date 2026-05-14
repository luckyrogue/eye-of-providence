package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/httperr"
)

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
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}

		out := onboardingStatus{}

		if s.Pool != nil {

			if err := s.Pool.QueryRow(c.Context(),
				`SELECT COUNT(*) FROM team_members WHERE user_id = $1`, uid,
			).Scan(&out.TeamsCount); err != nil {
				s.Logger.Warn("onboarding teams count failed", zap.Error(err))
			}

			var dismissedAt *string
			if err := s.Pool.QueryRow(c.Context(),
				`SELECT onboarding_dismissed_at::text FROM users WHERE id = $1`, uid,
			).Scan(&dismissedAt); err != nil {
				s.Logger.Warn("onboarding dismissed read failed", zap.Error(err))
			}
			out.Dismissed = dismissedAt != nil
		}

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

func onboardingDismissHandler(s MeService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		if s.Pool == nil {

			return c.JSON(fiber.Map{"status": "ok"})
		}

		if _, err := s.Pool.Exec(c.Context(),
			`UPDATE users SET onboarding_dismissed_at = COALESCE(onboarding_dismissed_at, now()) WHERE id = $1`, uid,
		); err != nil {
			s.Logger.Warn("onboarding dismiss failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}
