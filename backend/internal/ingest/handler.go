package ingest

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/ingest/batchapp"
	"github.com/eye-of-providence/backend/internal/httperr"
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
			return httperr.Unauthorized(c, "missing_subject", "missing subject")
		}
		var req request
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}

		valid, accepted, rejected, err := batchapp.PrepareIngest(claims.UserID, req.Events, maxEventsPerBatch)
		if err != nil {
			if errors.Is(err, batchapp.ErrBatchTooLarge) {
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
			if err := st.Insert(c.Context(), valid); err != nil {
				metrics.IngestErrors.Inc()
				logger.Error("store insert failed", zap.Error(err), zap.Int("count", len(valid)))
				return httperr.Internal(c)
			}
		}

		metrics.IngestEventsAccepted.Add(uint64(accepted))
		metrics.IngestEventsRejected.Add(uint64(rejected))
		logger.Debug("ingest batch", zap.String("user", claims.UserID), zap.Int("accepted", accepted), zap.Int("rejected", rejected))
		return c.JSON(response{Accepted: accepted, Rejected: rejected})
	})
}
