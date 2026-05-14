package migrate

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sql/postgres/*.sql sql/clickhouse/*.sql
var fsys embed.FS

const (
	pgSubdir = "sql/postgres"
	chSubdir = "sql/clickhouse"
)

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

	if subdir == pgSubdir {
		dsn = toPgx5DSN(dsn)
	}

	if subdir == chSubdir {
		dsn = ensureCHMigrationsEngine(dsn)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return nil, fmt.Errorf("migrate init for %s: %w", subdir, err)
	}
	return m, nil
}

func toPgx5DSN(dsn string) string {
	if strings.HasPrefix(dsn, "postgres://") {
		return "pgx5://" + strings.TrimPrefix(dsn, "postgres://")
	}
	if strings.HasPrefix(dsn, "postgresql://") {
		return "pgx5://" + strings.TrimPrefix(dsn, "postgresql://")
	}
	return dsn
}

func ensureCHMigrationsEngine(dsn string) string {
	addParam := func(s, key, value string) string {
		if strings.Contains(s, key+"=") {
			return s
		}
		sep := "?"
		if strings.Contains(s, "?") {
			sep = "&"
		}
		return s + sep + key + "=" + value
	}
	dsn = addParam(dsn, "x-migrations-table-engine", "MergeTree")
	dsn = addParam(dsn, "x-multi-statement", "true")
	return dsn
}

func RunPostgres(ctx context.Context, dsn string) error {
	return runUp(ctx, pgSubdir, dsn, "postgres")
}

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

	done := make(chan error, 1)
	go func() { done <- m.Up() }()
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("%s up: %w", label, err)
		}
		return nil
	case <-ctx.Done():

		return ctx.Err()
	}
}

func closeMigrator(m *migrate.Migrate) {
	if m == nil {
		return
	}
	_, _ = m.Close()
}
