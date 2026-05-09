// cmd/migrate — CLI для ручного управления миграциями (up / down / force / version).
//
// Auto-migrate при старте API делает только Up. Down и Force — деструктивны
// и должны идти через ручной запуск с человеком в loop'е.
//
// Usage:
//
//	migrate -db postgres up
//	migrate -db postgres down 1            # шаг назад на 1 миграцию
//	migrate -db postgres goto 3            # привести БД к версии 3 (вверх или вниз)
//	migrate -db postgres force 5           # пометить N как applied БЕЗ запуска SQL
//	                                       # (нужно при первом deploy этой версии
//	                                       # на старую БД, где schema_migrations
//	                                       # ещё нет, но таблицы уже накачены)
//	migrate -db postgres version           # показать текущую версию + dirty flag
//	migrate -db clickhouse up
//
// DSN читаются из EOP_POSTGRES_DSN / EOP_CLICKHOUSE_DSN, либо передаётся флагом -dsn.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"

	eopmigrate "github.com/eye-of-providence/backend/internal/migrate"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run возвращает ошибку вместо os.Exit'а — это даёт defer'ам корректно
// отработать и упрощает тестирование (в будущем).
func run() error {
	db := flag.String("db", "", "target database: postgres | clickhouse")
	dsn := flag.String("dsn", "", "DSN; default: EOP_POSTGRES_DSN или EOP_CLICKHOUSE_DSN")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if *db == "" || len(args) == 0 {
		usage()
		os.Exit(2) //nolint:gocritic // намеренный exit: usage без defer'ов
	}

	if *dsn == "" {
		switch *db {
		case "postgres":
			*dsn = os.Getenv("EOP_POSTGRES_DSN")
		case "clickhouse":
			*dsn = os.Getenv("EOP_CLICKHOUSE_DSN")
		default:
			return fmt.Errorf("unknown -db %q (expected postgres|clickhouse)", *db)
		}
	}
	if *dsn == "" {
		return fmt.Errorf("no DSN: pass -dsn or set EOP_%s_DSN", upper(*db))
	}

	m, err := newMigrator(*db, *dsn)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	cmd, rest := args[0], args[1:]
	switch cmd {
	case "up":
		if len(rest) > 0 {
			return fmt.Errorf("up: ожидаю 0 аргументов, получил %d", len(rest))
		}
		return reportErr(cmd, m.Up())
	case "down":
		// `down N` → шаг на N миграций назад. Без аргумента — ВСЁ. Опасно;
		// требуем явный confirm через env var.
		if len(rest) == 0 {
			return errors.New("down without arg = revert ALL migrations.\n  Если ты уверен — `down all` (требует EOP_MIGRATE_CONFIRM=yes-i-mean-it)")
		}
		if rest[0] == "all" {
			if os.Getenv("EOP_MIGRATE_CONFIRM") != "yes-i-mean-it" {
				return errors.New("`down all` требует EOP_MIGRATE_CONFIRM=yes-i-mean-it")
			}
			return reportErr(cmd, m.Down())
		}
		n, err := strconv.Atoi(rest[0])
		if err != nil || n <= 0 {
			return fmt.Errorf("down: ожидаю положительное число шагов, получил %q", rest[0])
		}
		return reportErr(cmd, m.Steps(-n))
	case "goto":
		if len(rest) != 1 {
			return errors.New("goto: ожидаю одну версию (uint)")
		}
		v, err := strconv.ParseUint(rest[0], 10, 32)
		if err != nil {
			return fmt.Errorf("goto: %w", err)
		}
		return reportErr(cmd, m.Migrate(uint(v)))
	case "force":
		if len(rest) != 1 {
			return errors.New("force: ожидаю одну версию (int)")
		}
		v, err := strconv.Atoi(rest[0])
		if err != nil {
			return fmt.Errorf("force: %w", err)
		}
		if err := m.Force(v); err != nil {
			return fmt.Errorf("force: %w", err)
		}
		fmt.Printf("forced version=%d (dirty=false)\n", v)
		return nil
	case "version":
		v, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("no migrations applied yet")
			return nil
		}
		if err != nil {
			return fmt.Errorf("version: %w", err)
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
		return nil
	default:
		usage()
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func newMigrator(db, dsn string) (*migrate.Migrate, error) {
	switch db {
	case "postgres":
		return eopmigrate.NewPostgres(dsn)
	case "clickhouse":
		return eopmigrate.NewClickHouse(dsn)
	}
	return nil, fmt.Errorf("unknown db %q", db)
}

// reportErr — печатает короткий success-сигнал и возвращает err для main.
func reportErr(cmd string, err error) error {
	if err == nil {
		fmt.Printf("%s: ok\n", cmd)
		return nil
	}
	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Printf("%s: no change\n", cmd)
		return nil
	}
	return fmt.Errorf("%s: %w", cmd, err)
}

func usage() {
	fmt.Fprintln(os.Stderr, `migrate — manual schema migrations.

Usage:
  migrate -db postgres|clickhouse <command> [args]

Commands:
  up                   apply all pending up migrations
  down N               step back N migrations
  down all             revert ALL (requires EOP_MIGRATE_CONFIRM=yes-i-mean-it)
  goto N               migrate up or down to version N
  force N              mark version N as applied without running SQL
                       (one-time use after switching from idempotent runner)
  version              print current version + dirty flag

Flags:
  -db string           target database (postgres|clickhouse)
  -dsn string          override default DSN

Env (defaults):
  EOP_POSTGRES_DSN     used when -db=postgres and -dsn empty
  EOP_CLICKHOUSE_DSN   used when -db=clickhouse and -dsn empty`)
}

func upper(s string) string {
	out := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		out[i] = c
	}
	return string(out)
}
