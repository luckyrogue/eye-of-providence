package reports

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/store"
)

type Service struct {
	Store      ReportStore
	EventStore store.EventStore
	Gemini     *GeminiClient
	Logger     *zap.Logger
	JWTSecret  string
	Pool       *pgxpool.Pool
}

func RegisterRoutes(app *fiber.App, s Service) {
	appSvc := NewReportsApp(s.Store, s.EventStore, s.Gemini)
	g := app.Group("/v1/reports", auth.Middleware(s.JWTSecret, s.Pool))

	g.Post("/generate", func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		report, err := appSvc.Generate(c.Context(), claims.UserID, c.Query("period", "weekly"), time.Now().UTC())
		if err != nil {
			if errors.Is(err, errReportGeneration) {
				return httperr.BadGateway(c, "report_generation_failed", "report generation failed")
			}
			s.Logger.Error("report generate failed", zap.Error(err))
			return httperr.Internal(c)
		}
		s.Logger.Info("report generated", zap.String("user", claims.UserID), zap.String("period", report.Period), zap.Int("body_len", len(report.BodyMD)))
		return c.JSON(report)
	})

	g.Get("/", func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		return c.JSON(fiber.Map{"reports": appSvc.List(claims.UserID, 20)})
	})

	g.Get("/:id", func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		r, ok := appSvc.Get(c.Params("id"), claims.UserID)
		if !ok {
			return httperr.NotFound(c, "report_not_found", "report not found")
		}
		return c.JSON(r)
	})
}
