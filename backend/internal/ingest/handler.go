package ingest

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/store"
)

type request struct {
	Events []store.Event `json:"events"`
}

type response struct {
	Accepted int `json:"accepted"`
	Rejected int `json:"rejected"`
}

func RegisterRoutes(app *fiber.App, st store.EventStore, logger *zap.Logger, jwtSecret string) {
	g := app.Group("/v1", auth.Middleware(jwtSecret))

	g.Post("/ingest", func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		var req request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}

		accepted, rejected := 0, 0
		valid := make([]store.Event, 0, len(req.Events))
		for _, e := range req.Events {
			if !validEvent(e) {
				rejected++
				continue
			}
			// принудительно перепривязываем user_id из токена,
			// чтобы клиент не мог писать чужие события
			e.UserID = claims.UserID
			if e.TS.IsZero() {
				e.TS = time.Now().UTC()
			}
			valid = append(valid, e)
			accepted++
		}

		if len(valid) > 0 {
			if err := st.Insert(c.Context(), valid); err != nil {
				logger.Error("store insert failed", zap.Error(err), zap.Int("count", len(valid)))
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "insert failed"})
			}
		}

		logger.Debug("ingest batch", zap.String("user", claims.UserID), zap.Int("accepted", accepted), zap.Int("rejected", rejected))
		return c.JSON(response{Accepted: accepted, Rejected: rejected})
	})
}

// validEvent — privacy-gate: запрещаем поля, которые могли бы содержать контент.
// Phase 1: базовая валидация, в Phase 5 расширяется до полного аудита.
func validEvent(e store.Event) bool {
	if e.AppBundle == "" || e.Source == "" || e.Category == "" {
		return false
	}
	switch e.Source {
	case "os", "browser", "ide", "cli":
	default:
		return false
	}
	switch e.Category {
	case "idle", "manual", "ai", "reading", "refactor", "other":
	default:
		return false
	}
	if e.DurationMS > 24*60*60*1000 {
		// больше суток в одном event — явно баг
		return false
	}
	return true
}
