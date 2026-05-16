// Worker — отдельный binary, запускается рядом с api/. Crontab-style
// бекграунд-процессы, которые не имеют смысла в request-loop:
//
//   - attribution (Phase A): полит `events` table, маппит на
//     `attribution_events` для дашборда (donut chart provenance).
//
// На MVP-этапе deploy = simple: docker-compose service `worker` запускает
// этот binary, использует те же EOP_POSTGRES_DSN / EOP_CLICKHOUSE_DSN.
// В V1 переедет в k8s CronJob (или leader-election Deployment если будет
// несколько replica'ов worker'ов).
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/attribution"
	"github.com/eye-of-providence/backend/internal/config"
	"github.com/eye-of-providence/backend/internal/log"
	"github.com/eye-of-providence/backend/internal/store"
)

func main() {
	cfg := config.FromEnv()
	logger := log.New(cfg.Env)
	defer func() { _ = logger.Sync() }()

	logger.Info("worker starting",
		zap.String("env", cfg.Env),
		zap.String("pg", redactDSN(cfg.PostgresDSN)),
		zap.String("ch", redactDSN(cfg.ClickHouseDSN)),
	)

	// Graceful shutdown — на SIGTERM/SIGINT останавливаем worker'ы через
	// ctx-cancel. В Kubernetes pod получит SIGTERM при deploy/scale down.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Postgres pool — для worker_state position tracking.
	pgCtx, pgCancel := context.WithTimeout(ctx, 10*time.Second)
	pg, err := pgxpool.New(pgCtx, cfg.PostgresDSN)
	pgCancel()
	if err != nil {
		logger.Fatal("postgres connect failed", zap.Error(err))
	}

	// ClickHouse — основное хранилище. attribution worker читает events,
	// пишет attribution_events.
	ch, err := store.OpenClickHouse(cfg.ClickHouseDSN)
	if err != nil {
		pg.Close()
		logger.Fatal("clickhouse connect failed", zap.Error(err))
	}

	w := attribution.New(ch.Conn(), pg, logger.Named("attribution"))
	err = w.Run(ctx)
	// Закрываем ресурсы явно перед выходом — defer пропускается при os.Exit,
	// поэтому даже на error-пути выпускаем pg/ch чтобы коннекты не висели.
	pg.Close()
	_ = ch.Close()
	if err != nil {
		logger.Error("attribution worker exited", zap.Error(err))
		os.Exit(1) //nolint:gocritic // pg/ch closed выше; defer'ы для них убраны
	}
	logger.Info("worker stopped cleanly")
}

// redactDSN — обрезает password из DSN перед логированием. Хорошая
// практика: даже env-logged строки не должны содержать secrets.
func redactDSN(dsn string) string {
	// Простейший redact: ищем `:password@` и заменяем на `:***@`.
	// Не парсим URI полностью — overkill для строкового лога.
	at := -1
	for i := len(dsn) - 1; i >= 0; i-- {
		if dsn[i] == '@' {
			at = i
			break
		}
	}
	if at < 0 {
		return dsn
	}
	prefix := dsn[:at]
	// Найдём `:` после `://`
	colon := -1
	for i := at - 1; i >= 0; i-- {
		if prefix[i] == ':' {
			// Убедимся что это не `://`
			if i > 1 && prefix[i-2:i+1] == "://" {
				continue
			}
			colon = i
			break
		}
	}
	if colon < 0 {
		return dsn
	}
	return dsn[:colon+1] + "***" + dsn[at:]
}
