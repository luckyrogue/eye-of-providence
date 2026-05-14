package reports

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/reports/periodapp"
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
		from, to, periodKey := periodapp.Resolve(periodKind, time.Now().UTC())

		nc, err := BuildContext(c.Context(), s.EventStore, claims.UserID, periodKey, from, to)
		if err != nil {
			s.Logger.Error("aggregate failed", zap.Error(err))
			return httperr.Internal(c)
		}

		body, err := s.Gemini.Generate(c.Context(), nc)
		if err != nil {
			s.Logger.Error("gemini failed", zap.Error(err))
			return httperr.BadGateway(c, "report_generation_failed", "report generation failed")
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
			return httperr.NotFound(c, "report_not_found", "report not found")
		}
		return c.JSON(r)
	})
}
