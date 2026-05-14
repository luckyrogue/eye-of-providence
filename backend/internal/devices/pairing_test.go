//go:build integration

package devices

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/migrate"
)

func setupDevicesDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("EOP_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("EOP_TEST_PG_DSN not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := migrate.RunPostgres(ctx, dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	_, err = pool.Exec(ctx, "TRUNCATE users, api_tokens, pairing_codes RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			"TRUNCATE users, api_tokens, pairing_codes RESTART IDENTITY CASCADE")
		pool.Close()
	})
	return pool
}

func makeUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO users (id, email, display_name) VALUES ($1, $2, $3)",
		id, id.String()+"@example.com", "Test")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func TestPairing_HappyPath(t *testing.T) {
	pool := setupDevicesDB(t)
	uid := makeUser(t, pool)
	ctx := context.Background()

	pair, err := PairBegin(ctx, pool, "ext", "Chrome on Mac")
	if err != nil {
		t.Fatalf("PairBegin: %v", err)
	}
	if len(pair.Code) != codeLen {
		t.Fatalf("code len=%d, want %d", len(pair.Code), codeLen)
	}
	if pair.Secret == "" || pair.PairID == uuid.Nil {
		t.Fatal("empty secret or pair_id")
	}

	pr, err := Poll(ctx, pool, pair.PairID, pair.Secret)
	if err != nil {
		t.Fatalf("Poll pending: %v", err)
	}
	if pr.Status != "pending" {
		t.Errorf("status=%q, want pending", pr.Status)
	}

	dev, err := Claim(ctx, pool, uid, pair.Code, "My laptop")
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if dev.Kind != "ext" || dev.Name != "My laptop" {
		t.Errorf("dev=%+v", dev)
	}

	pr, err = Poll(ctx, pool, pair.PairID, pair.Secret)
	if err != nil {
		t.Fatalf("Poll claimed: %v", err)
	}
	if pr.Status != "claimed" {
		t.Errorf("status=%q, want claimed", pr.Status)
	}
	if pr.Token == nil || *pr.Token == "" {
		t.Fatal("token nil/empty after claim")
	}

	pr2, err := Poll(ctx, pool, pair.PairID, pair.Secret)
	if err != nil {
		t.Fatalf("Poll second: %v", err)
	}
	if pr2.Token != nil {
		t.Errorf("token leaked on second poll: %v", pr2.Token)
	}
}

func TestPairing_ExpiredCode(t *testing.T) {
	pool := setupDevicesDB(t)
	ctx := context.Background()

	pair, err := PairBegin(ctx, pool, "agent", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx,
		"UPDATE pairing_codes SET code_expires_at = now() - interval '1 minute' WHERE id = $1",
		pair.PairID)
	if err != nil {
		t.Fatal(err)
	}
	pr, err := Poll(ctx, pool, pair.PairID, pair.Secret)
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if pr.Status != "expired" {
		t.Errorf("status=%q, want expired", pr.Status)
	}
}

func TestPairing_WrongSecret(t *testing.T) {
	pool := setupDevicesDB(t)
	ctx := context.Background()

	pair, err := PairBegin(ctx, pool, "ide", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Poll(ctx, pool, pair.PairID, "not-the-real-secret")
	if err != ErrSecretMismatch {
		t.Errorf("err=%v, want ErrSecretMismatch", err)
	}
}

func TestPairing_ClaimStaleCode(t *testing.T) {
	pool := setupDevicesDB(t)
	uid := makeUser(t, pool)
	ctx := context.Background()

	pair, err := PairBegin(ctx, pool, "ext", "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx,
		"UPDATE pairing_codes SET code_expires_at = now() - interval '1 hour' WHERE id = $1",
		pair.PairID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Claim(ctx, pool, uid, pair.Code, "x")
	if err != ErrCodeNotFound {
		t.Errorf("err=%v, want ErrCodeNotFound", err)
	}
}

func TestPairing_DoubleClaim(t *testing.T) {
	pool := setupDevicesDB(t)
	uid := makeUser(t, pool)
	ctx := context.Background()

	pair, err := PairBegin(ctx, pool, "agent", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Claim(ctx, pool, uid, pair.Code, "first"); err != nil {
		t.Fatalf("first claim: %v", err)
	}
	_, err = Claim(ctx, pool, uid, pair.Code, "second")
	if err != ErrAlreadyClaimed {
		t.Errorf("err=%v, want ErrAlreadyClaimed", err)
	}
}

func TestPairing_ListAndRevoke(t *testing.T) {
	pool := setupDevicesDB(t)
	uid := makeUser(t, pool)
	ctx := context.Background()

	pair, _ := PairBegin(ctx, pool, "ext", "")
	dev, err := Claim(ctx, pool, uid, pair.Code, "browser")
	if err != nil {
		t.Fatal(err)
	}

	list, err := List(ctx, pool, uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != dev.ID {
		t.Errorf("list=%+v, want 1 device with id=%v", list, dev.ID)
	}

	ok, err := Revoke(ctx, pool, uid, dev.ID)
	if err != nil || !ok {
		t.Fatalf("revoke: ok=%v err=%v", ok, err)
	}

	list, _ = List(ctx, pool, uid)
	if len(list) != 0 {
		t.Errorf("list after revoke len=%d, want 0", len(list))
	}
}
