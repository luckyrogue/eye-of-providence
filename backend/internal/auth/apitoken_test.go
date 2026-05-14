//go:build integration

package auth

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/migrate"
)

func setupAuthDB(t *testing.T) *pgxpool.Pool {
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
	_, err = pool.Exec(ctx, "TRUNCATE users, api_tokens RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "TRUNCATE users, api_tokens RESTART IDENTITY CASCADE")
		pool.Close()
	})
	return pool
}

func makeUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	hash, _ := HashPassword("dev123456")
	_, err := pool.Exec(context.Background(),
		"INSERT INTO users (id, email, display_name, password_hash) VALUES ($1, $2, $3, $4)",
		id, id.String()+"@example.com", "T", hash)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func TestCreateAPIToken_Format(t *testing.T) {
	pool := setupAuthDB(t)
	uid := makeUser(t, pool)

	plain, row, err := CreateAPIToken(context.Background(), pool, uid, "test", "read", 0)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !strings.HasPrefix(plain, "eop_") {
		t.Errorf("plaintext = %q, want eop_ prefix", plain)
	}
	if len(plain) != 4+48 {
		t.Errorf("plaintext len = %d, want 52", len(plain))
	}
	if !strings.HasPrefix(plain, row.Prefix) || len(row.Prefix) != 8 {
		t.Errorf("prefix=%q, plain=%q", row.Prefix, plain)
	}
	if row.Scope != "read" {
		t.Errorf("scope=%q, want read", row.Scope)
	}
	if row.ExpiresAt != nil {
		t.Errorf("expires_at = %v, want nil (ttl=0)", row.ExpiresAt)
	}
}

func TestCreateAPIToken_TTL(t *testing.T) {
	pool := setupAuthDB(t)
	uid := makeUser(t, pool)
	_, row, err := CreateAPIToken(context.Background(), pool, uid, "ttl", "read", 7*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if row.ExpiresAt == nil {
		t.Fatal("expires_at nil with ttl=7d")
	}
	dt := time.Until(*row.ExpiresAt)
	if dt < 6*24*time.Hour || dt > 8*24*time.Hour {
		t.Errorf("expires_at delta = %v, want ~7d", dt)
	}
}

func TestCreateAPIToken_BadScope(t *testing.T) {
	pool := setupAuthDB(t)
	uid := makeUser(t, pool)
	_, _, err := CreateAPIToken(context.Background(), pool, uid, "x", "godmode", 0)
	if err == nil {
		t.Error("expected error for invalid scope")
	}
}

func TestVerifyAPIToken_HappyPath(t *testing.T) {
	pool := setupAuthDB(t)
	uid := makeUser(t, pool)
	plain, _, err := CreateAPIToken(context.Background(), pool, uid, "v", "read", 0)
	if err != nil {
		t.Fatal(err)
	}

	gotUID, gotScope, err := VerifyAPIToken(context.Background(), pool, plain)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if gotUID != uid {
		t.Errorf("uid=%v, want %v", gotUID, uid)
	}
	if gotScope != "read" {
		t.Errorf("scope=%q, want read", gotScope)
	}

	var lastUsed *time.Time
	_ = pool.QueryRow(context.Background(),
		"SELECT last_used_at FROM api_tokens WHERE user_id=$1", uid).Scan(&lastUsed)
	if lastUsed == nil {
		t.Error("last_used_at not updated")
	}
}

func TestVerifyAPIToken_BadToken(t *testing.T) {
	pool := setupAuthDB(t)
	_, _, err := VerifyAPIToken(context.Background(), pool, "eop_nonsense")
	if err != ErrTokenInvalid {
		t.Errorf("err=%v, want ErrTokenInvalid", err)
	}
}

func TestVerifyAPIToken_NoEopPrefix(t *testing.T) {
	pool := setupAuthDB(t)
	_, _, err := VerifyAPIToken(context.Background(), pool, "github_pat_abcdef")
	if err != ErrTokenInvalid {
		t.Errorf("err=%v, want ErrTokenInvalid", err)
	}
}

func TestVerifyAPIToken_Revoked(t *testing.T) {
	pool := setupAuthDB(t)
	uid := makeUser(t, pool)
	plain, row, _ := CreateAPIToken(context.Background(), pool, uid, "r", "read", 0)
	if _, err := RevokeAPIToken(context.Background(), pool, uid, row.ID); err != nil {
		t.Fatal(err)
	}
	_, _, err := VerifyAPIToken(context.Background(), pool, plain)
	if err != ErrTokenInvalid {
		t.Errorf("err=%v, want ErrTokenInvalid (revoked)", err)
	}
}

func TestVerifyAPIToken_Expired(t *testing.T) {
	pool := setupAuthDB(t)
	uid := makeUser(t, pool)
	plain, row, _ := CreateAPIToken(context.Background(), pool, uid, "e", "read", time.Hour)

	_, _ = pool.Exec(context.Background(),
		"UPDATE api_tokens SET expires_at = now() - interval '1 second' WHERE id = $1", row.ID)
	_, _, err := VerifyAPIToken(context.Background(), pool, plain)
	if err != ErrTokenInvalid {
		t.Errorf("err=%v, want ErrTokenInvalid (expired)", err)
	}
}

func TestListAPITokens(t *testing.T) {
	pool := setupAuthDB(t)
	uid := makeUser(t, pool)
	_, _, _ = CreateAPIToken(context.Background(), pool, uid, "a", "read", 0)
	_, _, _ = CreateAPIToken(context.Background(), pool, uid, "b", "write:ingest", 0)
	tokens, err := ListAPITokens(context.Background(), pool, uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 2 {
		t.Errorf("count=%d, want 2", len(tokens))
	}
	for _, tok := range tokens {
		if !strings.HasPrefix(tok.Prefix, "eop_") {
			t.Errorf("prefix=%q, missing eop_", tok.Prefix)
		}
	}
}

func TestRevokeAPIToken_OtherUser(t *testing.T) {
	pool := setupAuthDB(t)
	uid1 := makeUser(t, pool)
	uid2 := makeUser(t, pool)
	_, row, _ := CreateAPIToken(context.Background(), pool, uid1, "x", "read", 0)
	ok, err := RevokeAPIToken(context.Background(), pool, uid2, row.ID)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("revoke succeeded for wrong user")
	}
}
