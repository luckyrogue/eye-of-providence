// Package migrate — простой идемпотентный runner SQL-миграций для Postgres и ClickHouse.
//
// Файлы embed'ятся в бинарь из backend/migrations/. Все миграции написаны
// idempotent через CREATE ... IF NOT EXISTS / ALTER ... IF NOT EXISTS,
// так что применять их повторно безопасно — отдельной таблицы версий не нужно.
//
// Запуск: на старте API при EOP_AUTO_MIGRATE=true.
package migrate

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/*.sql
var fs embed.FS

// pgMigrationLockID — sentinel id для pg_advisory_lock.
// Сериализует одновременные попытки прогнать миграции из нескольких replica'ей
// при rolling deploy. Тот, кто получает lock, прогоняет миграции; остальные
// ждут, потом видят что миграции уже применены (CREATE IF NOT EXISTS) и идут
// дальше.
const pgMigrationLockID int64 = 8331_2026_002

// RunPostgres — применяет все sql/NNN_*.up.sql (где NNN — номер) к Postgres.
// Безопасен при параллельном запуске на нескольких replica'ях: использует
// session-level pg_advisory_lock, чтобы только один процесс прогонял DDL.
func RunPostgres(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return nil
	}
	files, err := listSorted("sql", func(name string) bool {
		return strings.HasSuffix(name, ".up.sql") && !strings.HasPrefix(name, "clickhouse_")
	})
	if err != nil {
		return err
	}
	// Берём один коннект и держим lock на всё время DDL.
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire conn for migrate: %w", err)
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", pgMigrationLockID); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		// Лок сам отпустится при Release/disconnect, но явный unlock корректнее.
		_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", pgMigrationLockID)
	}()
	for _, name := range files {
		body, err := fs.ReadFile("sql/" + name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if _, err := conn.Exec(ctx, string(body)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
	}
	return nil
}

// RunClickHouse — применяет sql/clickhouse_*.sql к ClickHouse.
// Использует тот же DSN, что и event store, но открывает свой коннект (closing после).
func RunClickHouse(ctx context.Context, dsn string) error {
	if dsn == "" {
		return nil
	}
	files, err := listSorted("sql", func(name string) bool {
		return strings.HasPrefix(name, "clickhouse_") && strings.HasSuffix(name, ".sql")
	})
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}
	conn, err := openCH(dsn)
	if err != nil {
		return fmt.Errorf("clickhouse open: %w", err)
	}
	defer conn.Close()

	for _, name := range files {
		body, err := fs.ReadFile("sql/" + name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		// Каждый statement в .sql файле — отдельный Exec. ClickHouse driver
		// не любит multi-statement в одном Exec-вызове.
		for _, stmt := range splitStatements(string(body)) {
			if stmt == "" {
				continue
			}
			if err := conn.Exec(ctx, stmt); err != nil {
				return fmt.Errorf("apply %s [%s...]: %w", name, firstChars(stmt, 60), err)
			}
		}
	}
	return nil
}

func listSorted(dir string, pred func(string) bool) ([]string, error) {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if pred(e.Name()) {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

func openCH(dsn string) (clickhouse.Conn, error) {
	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	if opts.DialTimeout == 0 {
		opts.DialTimeout = 5 * time.Second
	}
	return clickhouse.Open(opts)
}

// splitStatements — наивный сплит по `;` в конце строки. Достаточно для
// наших миграций (без процедур и вложенных стрингов с ; внутри).
func splitStatements(body string) []string {
	body = strings.TrimSpace(body)
	parts := strings.Split(body, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func firstChars(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n]
}
