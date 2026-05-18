package attributionapp

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/attribution"
)

// Worker — application entry for attribution pipeline.
type Worker struct {
	inner *attribution.Worker
}

func New(ch driver.Conn, pg *pgxpool.Pool, logger *zap.Logger) *Worker {
	return &Worker{inner: attribution.New(ch, pg, logger)}
}

func (w *Worker) Run(ctx context.Context) error {
	if w.inner == nil {
		return nil
	}
	return w.inner.Run(ctx)
}

func (w *Worker) SetPollInterval(d time.Duration) {
	if w.inner != nil {
		w.inner.Poll = d
	}
}
