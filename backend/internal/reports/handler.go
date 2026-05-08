package reports

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/store"
)

const promptVersion = "v0.1"

type Service struct {
	Store      ReportStore
	EventStore store.EventStore
	Gemini     *GeminiClient
	Logger     *zap.Logger
	JWTSecret  string
	Pool       *pgxpool.Pool
}

func RegisterRoutes(app *fiber.App, s Service) {
	g := app.Group("/v1/reports", auth.Middleware(s.JWTSecret, s.Pool))

	g.Post("/generate", func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)

		periodKind := c.Query("period", "weekly")
		from, to, periodKey := resolvePeriod(periodKind)

		nc, err := BuildContext(c.Context(), s.EventStore, claims.UserID, periodKey, from, to)
		if err != nil {
			s.Logger.Error("aggregate failed", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "aggregate failed"})
		}

		body, err := s.Gemini.Generate(c.Context(), nc)
		if err != nil {
			s.Logger.Error("gemini failed", zap.Error(err))
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "report generation failed", "detail": err.Error()})
		}

		report := Report{
			ID:            uuid.NewString(),
			UserID:        claims.UserID,
			Period:        periodKey,
			Model:         s.Gemini.Model,
			BodyMD:        body,
			GeneratedAt:   time.Now().UTC(),
			PromptVersion: promptVersion,
		}
		s.Store.Save(report)
		s.Logger.Info("report generated", zap.String("user", claims.UserID), zap.String("period", periodKey), zap.Int("body_len", len(body)))
		return c.JSON(report)
	})

	g.Get("/", func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		out := s.Store.ListForUser(claims.UserID, 20)
		return c.JSON(fiber.Map{"reports": out})
	})

	g.Get("/:id", func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		r, ok := s.Store.Get(c.Params("id"), claims.UserID)
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
		}
		return c.JSON(r)
	})
}

func resolvePeriod(kind string) (time.Time, time.Time, string) {
	now := time.Now().UTC()
	switch kind {
	case "monthly":
		from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, 0).Add(-time.Second)
		key := fmt.Sprintf("monthly_%04d_%02d", now.Year(), now.Month())
		return from, to, key
	case "daily":
		from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		to := from.Add(24 * time.Hour).Add(-time.Second)
		key := fmt.Sprintf("daily_%s", from.Format("2006-01-02"))
		return from, to, key
	default: // weekly
		// ISO week-of-year
		year, week := now.ISOWeek()
		// найдём понедельник этой недели
		offset := int(now.Weekday())
		if offset == 0 {
			offset = 7
		}
		monday := now.AddDate(0, 0, 1-offset)
		from := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 0, 7).Add(-time.Second)
		key := fmt.Sprintf("weekly_%04d_W%02d", year, week)
		return from, to, key
	}
}
