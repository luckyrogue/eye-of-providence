package ingest

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/metrics"
	"github.com/eye-of-providence/backend/internal/store"
)

type request struct {
	Events []store.Event `json:"events"`
}

type response struct {
	Accepted int `json:"accepted"`
	Rejected int `json:"rejected"`
}

// maxEventsPerBatch — защита от 1 запроса на миллион событий, который может
// насильно нагрузить ClickHouse. Клиенты должны слать по batch'ам.
const maxEventsPerBatch = 5000

func RegisterRoutes(app *fiber.App, st store.EventStore, logger *zap.Logger, jwtSecret string, pool *pgxpool.Pool) {
	g := app.Group("/v1", auth.Middleware(jwtSecret, pool))

	// JWT (dashboard/agent session) — full access. API token — нужен
	// scope=write:ingest или admin. read-only token не может ingestить.
	g.Post("/ingest", auth.RequireScope("write:ingest", "admin"), func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		if claims.UserID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing subject"})
		}
		var req request
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		if len(req.Events) > maxEventsPerBatch {
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error":     "too many events in one batch",
				"max_batch": maxEventsPerBatch,
			})
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
				metrics.IngestErrors.Inc()
				logger.Error("store insert failed", zap.Error(err), zap.Int("count", len(valid)))
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "insert failed"})
			}
		}

		metrics.IngestEventsAccepted.Add(uint64(accepted))
		metrics.IngestEventsRejected.Add(uint64(rejected))
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
