//go:build integration

// Запуск: EOP_TEST_PG_DSN=postgres://eop:eop_dev@localhost:5432/eop_test \
//   go test -tags=integration ./internal/migrate/...

package migrate

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
)

// TestPostgresRoundtrip — up до головы, потом down all, потом up снова.
// Если хоть одна .down.sql битая — тест поймает.
func TestPostgresRoundtrip(t *testing.T) {
	dsn := os.Getenv("EOP_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("EOP_TEST_PG_DSN not set; skipping integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1) Up — должен сходиться к latest.
	if err := RunPostgres(ctx, dsn); err != nil {
		t.Fatalf("first up: %v", err)
	}

	// 2) Снять текущую версию.
	m, err := NewPostgres(dsn)
	if err != nil {
		t.Fatalf("new migrator: %v", err)
	}
	defer func() { _, _ = m.Close() }()

	v, dirty, err := m.Version()
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if dirty {
		t.Fatalf("expected dirty=false after clean up, got dirty=true at v=%d", v)
	}
	if v == 0 {
		t.Fatalf("expected non-zero version after up")
	}
	t.Logf("after up: version=%d", v)

	// 3) Down all.
	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("down all: %v", err)
	}

	// 4) Up снова — должно быть ровно как было.
	if err := RunPostgres(ctx, dsn); err != nil {
		t.Fatalf("second up: %v", err)
	}

	m2, err := NewPostgres(dsn)
	if err != nil {
		t.Fatalf("new migrator 2: %v", err)
	}
	defer func() { _, _ = m2.Close() }()

	v2, dirty2, err := m2.Version()
	if err != nil {
		t.Fatalf("version after second up: %v", err)
	}
	if dirty2 {
		t.Fatalf("expected dirty=false after second up, got dirty=true at v=%d", v2)
	}
	if v2 != v {
		t.Fatalf("version mismatch: first=%d, second=%d", v, v2)
	}
}

// TestPostgresStepByStep — гарантирует, что каждая отдельная up.sql + down.sql
// валидна в обе стороны. Идём по одной от 0 до head, потом обратно по одной.
func TestPostgresStepByStep(t *testing.T) {
	dsn := os.Getenv("EOP_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("EOP_TEST_PG_DSN not set; skipping integration test")
	}

	m, err := NewPostgres(dsn)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	defer func() { _, _ = m.Close() }()

	// Сбросить в 0.
	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("reset down: %v", err)
	}

	// Up по одной.
	for i := 0; i < 100; i++ {
		err := m.Steps(1)
		if errors.Is(err, migrate.ErrNoChange) || errors.Is(err, os.ErrNotExist) {
			break
		}
		if err != nil {
			t.Fatalf("step up #%d: %v", i+1, err)
		}
	}
	headV, _, err := m.Version()
	if err != nil {
		t.Fatalf("head version: %v", err)
	}

	// Down по одной до 0.
	for i := 0; i < 100; i++ {
		err := m.Steps(-1)
		if errors.Is(err, migrate.ErrNoChange) {
			break
		}
		if err != nil {
			t.Fatalf("step down #%d: %v", i+1, err)
		}
		// После полного отката Version() возвращает ErrNilVersion — выходим.
		if _, _, verr := m.Version(); errors.Is(verr, migrate.ErrNilVersion) {
			break
		}
	}

	// Up обратно до head.
	for i := 0; i < 100; i++ {
		err := m.Steps(1)
		if errors.Is(err, migrate.ErrNoChange) || errors.Is(err, os.ErrNotExist) {
			break
		}
		if err != nil {
			t.Fatalf("re-up step #%d: %v", i+1, err)
		}
	}
	finalV, dirty, err := m.Version()
	if err != nil {
		t.Fatalf("final version: %v", err)
	}
	if dirty {
		t.Fatalf("dirty after step-by-step roundtrip")
	}
	if finalV != headV {
		t.Fatalf("step roundtrip changed version: head=%d, final=%d", headV, finalV)
	}
}
