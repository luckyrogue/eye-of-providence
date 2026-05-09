// Package migrate — тонкая обвёртка вокруг golang-migrate/migrate/v4.
//
// Файлы embed'ятся из:
//   - sql/postgres/NNN_*.up.sql, NNN_*.down.sql       — Postgres
//   - sql/clickhouse/NNN_*.up.sql, NNN_*.down.sql     — ClickHouse Cloud
//
// На старте API (cmd/api при EOP_AUTO_MIGRATE=true) применяется только Up.
// Down/Force/Version — через cmd/migrate (отдельный binary, ручной запуск).
//
// golang-migrate ведёт state в `schema_migrations` (PG) и
// `schema_migrations` (CH) таблицах + использует свой advisory_lock на PG —
// ручной pg_advisory_lock больше не нужен.
//
// При первом deploy этой версии на prod, где БД уже была накачена нашим
// прежним idempotent-runner'ом, надо однократно сделать `migrate force 5`
// (PG) и `migrate force 2` (CH), чтобы проставить current version без
// повторного применения. См. cmd/migrate для деталей.
package migrate

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	// pgx/v5 driver — та же библиотека, что и app pool (jackc/pgx/v5),
	// одинаковая семантика DSN. Альтернатива (database/postgres → lib/pq)
	// требует явный sslmode=disable, иначе на серверах без SSL даёт
	// "SSL is not enabled on the server" — ловили в prod.
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sql/postgres/*.sql sql/clickhouse/*.sql
var fsys embed.FS

const (
	pgSubdir = "sql/postgres"
	chSubdir = "sql/clickhouse"
)

// NewPostgres / NewClickHouse возвращают готовый *migrate.Migrate для CLI.
// Caller обязан вызвать .Close() после использования.
func NewPostgres(dsn string) (*migrate.Migrate, error) {
	return newMigrator(pgSubdir, dsn)
}

func NewClickHouse(dsn string) (*migrate.Migrate, error) {
	return newMigrator(chSubdir, dsn)
}

func newMigrator(subdir, dsn string) (*migrate.Migrate, error) {
	if dsn == "" {
		return nil, errors.New("empty dsn")
	}
	src, err := iofs.New(fsys, subdir)
	if err != nil {
		return nil, fmt.Errorf("iofs source for %s: %w", subdir, err)
	}
	// PG: переключаем схему DSN с `postgres://` на `pgx5://`, чтобы
	// golang-migrate использовал pgx/v5 driver (а не lib/pq).
	if subdir == pgSubdir {
		dsn = toPgx5DSN(dsn)
	}
	// CH: golang-migrate по умолчанию создаёт schema_migrations с Engine=TinyLog,
	// но ClickHouse Cloud (Shared databases) запрещает движки на локальный диск.
	if subdir == chSubdir {
		dsn = ensureCHMigrationsEngine(dsn)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return nil, fmt.Errorf("migrate init for %s: %w", subdir, err)
	}
	return m, nil
}

// toPgx5DSN — `postgres://...` / `postgresql://...` → `pgx5://...`.
// Юзер держит в env стандартный postgres:// DSN, мы перенаправляем golang-migrate
// на pgx/v5 driver. Если уже pgx5:// — оставляем как есть.
func toPgx5DSN(dsn string) string {
	if strings.HasPrefix(dsn, "postgres://") {
		return "pgx5://" + strings.TrimPrefix(dsn, "postgres://")
	}
	if strings.HasPrefix(dsn, "postgresql://") {
		return "pgx5://" + strings.TrimPrefix(dsn, "postgresql://")
	}
	return dsn
}

// ensureCHMigrationsEngine — добавляет `x-migrations-table-engine=MergeTree`
// в CH DSN если не задан. Нужно для CH Cloud Shared databases (default
// TinyLog запрещён). MergeTree в Cloud автоматически реплицируется.
func ensureCHMigrationsEngine(dsn string) string {
	if strings.Contains(dsn, "x-migrations-table-engine=") {
		return dsn
	}
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + "x-migrations-table-engine=MergeTree"
}

// RunPostgres — auto-migrate on API startup. No-op if dsn пустой.
// Возвращает nil если БД уже на последней версии (ErrNoChange = ok).
func RunPostgres(ctx context.Context, dsn string) error {
	return runUp(ctx, pgSubdir, dsn, "postgres")
}

// RunClickHouse — auto-migrate on API startup. No-op если dsn пустой.
func RunClickHouse(ctx context.Context, dsn string) error {
	return runUp(ctx, chSubdir, dsn, "clickhouse")
}

func runUp(ctx context.Context, subdir, dsn, label string) error {
	if dsn == "" {
		return nil
	}
	m, err := newMigrator(subdir, dsn)
	if err != nil {
		return err
	}
	defer closeMigrator(m)

	// m.Up() игнорирует ctx; делаем cancel через GracefulStop.
	done := make(chan error, 1)
	go func() { done <- m.Up() }()
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("%s up: %w", label, err)
		}
		return nil
	case <-ctx.Done():
		// Не пытаемся стопать middle-of-DDL — это опаснее, чем дождаться.
		// Просто возвращаем ctx.Err() — caller увидит таймаут.
		// Сама миграция доедет в фоне и вернёт ошибку через done (которую мы дропнем).
		return ctx.Err()
	}
}

// closeMigrator — лучшее усилие: source + database close. Ошибки не валим
// наверх (миграции уже отработали к этому моменту).
func closeMigrator(m *migrate.Migrate) {
	if m == nil {
		return
	}
	_, _ = m.Close()
}
