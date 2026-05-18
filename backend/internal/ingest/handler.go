package ingest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
	"github.com/eye-of-providence/backend/internal/ingest/domain"
	"github.com/eye-of-providence/backend/internal/ingest/ingestapp"
	"github.com/eye-of-providence/backend/internal/metrics"
	"github.com/eye-of-providence/backend/internal/store"
)

type request struct {
	Events []domain.Event `json:"events"`
}

type response struct {
	Accepted int `json:"accepted"`
	Rejected int `json:"rejected"`
}

const maxEventsPerBatch = 5000

func RegisterRoutes(app *fiber.App, st store.EventStore, logger *zap.Logger, jwtSecret string, pool *pgxpool.Pool) {
	svc := newIngestApp(st)
	g := app.Group("/v1", auth.Middleware(jwtSecret, pool))

	g.Post("/ingest", auth.RequireScope("write:ingest", "admin"), func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		if claims.UserID == "" {
			return httperr.Unauthorized(c, "missing_subject", "missing subject")
		}
		var req request
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}

		valid, res, err := svc.PrepareBatch(claims.UserID, req.Events, maxEventsPerBatch)
		if err != nil {
			if ingestapp.IsBatchTooLarge(err) {
				return httperr.Send(c, httperr.ProblemDetails{
					Status: fiber.StatusRequestEntityTooLarge,
					Code:   "batch_too_large",
					Detail: "too many events in one batch",
					Extensions: map[string]any{
						"max_batch": maxEventsPerBatch,
					},
				})
			}
			return httperr.Internal(c)
		}

		if len(valid) > 0 {
			if err := svc.PersistBatch(c.Context(), valid); err != nil {
				metrics.IngestErrors.Inc()
				logger.Error("store insert failed", zap.Error(err), zap.Int("count", len(valid)))
				return httperr.Internal(c)
			}
		}

		metrics.IngestEventsAccepted.Add(uint64(res.Accepted))
		metrics.IngestEventsRejected.Add(uint64(res.Rejected))
		logger.Debug("ingest batch", zap.String("user", claims.UserID), zap.Int("accepted", res.Accepted), zap.Int("rejected", res.Rejected))
		return c.JSON(response{Accepted: res.Accepted, Rejected: res.Rejected})
	})
}
