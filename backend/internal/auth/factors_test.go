//go:build integration

package auth

import (
	"context"
	"crypto/rand"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/migrate"
)

func setupFactorsDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("EOP_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("EOP_TEST_PG_DSN not set; skipping integration test")
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

	if _, err := pool.Exec(ctx, "TRUNCATE users, user_identities, webauthn_credentials CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			"TRUNCATE users, user_identities, webauthn_credentials CASCADE")
		pool.Close()
	})
	return pool
}

func insertUserWithPassword(t *testing.T, pool *pgxpool.Pool, email string, hasPassword bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	var pwh *string
	if hasPassword {
		h, _ := HashPassword("password123")
		pwh = &h
	}
	_, err := pool.Exec(context.Background(),
		"INSERT INTO users (id, email, display_name, password_hash) VALUES ($1, $2, $3, $4)",
		id, email, "u", pwh)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func insertIdentity(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID, provider, subject string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO user_identities (id, user_id, provider, subject, email)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, userID, provider, subject, "ident@example.com")
	if err != nil {
		t.Fatalf("insert identity (%s/%s): %v", provider, subject, err)
	}
	return id
}

func insertPasskey(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) (passkeyID uuid.UUID, credentialID []byte) {
	t.Helper()
	credentialID = make([]byte, 32)
	if _, err := rand.Read(credentialID); err != nil {
		t.Fatalf("rand: %v", err)
	}
	passkeyID = uuid.New()

	pubKey := make([]byte, 65)
	_, err := pool.Exec(context.Background(), `
		INSERT INTO webauthn_credentials
			(id, user_id, credential_id, public_key, sign_count, transports, nickname)
		VALUES ($1, $2, $3, $4, 0, '', 'test')
	`, passkeyID, userID, credentialID, pubKey)
	if err != nil {
		t.Fatalf("insert passkey: %v", err)
	}
	return passkeyID, credentialID
}

func TestCountAuthFactors_PasswordOnly(t *testing.T) {
	pool := setupFactorsDB(t)
	uid := insertUserWithPassword(t, pool, "pwd-only@example.com", true)

	f, err := CountAuthFactors(context.Background(), pool, uid, nil, nil)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if !f.PasswordSet {
		t.Error("PasswordSet=false, want true")
	}
	if f.Identities != 0 {
		t.Errorf("Identities=%d, want 0", f.Identities)
	}
	if f.Passkeys != 0 {
		t.Errorf("Passkeys=%d, want 0", f.Passkeys)
	}
	if got, want := f.Total(), 1; got != want {
		t.Errorf("Total=%d, want %d", got, want)
	}
}

func TestCountAuthFactors_TwoIdentitiesNoPassword(t *testing.T) {
	pool := setupFactorsDB(t)
	uid := insertUserWithPassword(t, pool, "oauth-only@example.com", false)
	insertIdentity(t, pool, uid, "github", "12345")
	insertIdentity(t, pool, uid, "google", "google-sub-678")

	f, err := CountAuthFactors(context.Background(), pool, uid, nil, nil)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if f.PasswordSet {
		t.Error("PasswordSet=true, want false")
	}
	if f.Identities != 2 {
		t.Errorf("Identities=%d, want 2", f.Identities)
	}
	if f.Passkeys != 0 {
		t.Errorf("Passkeys=%d, want 0", f.Passkeys)
	}
	if got, want := f.Total(), 2; got != want {
		t.Errorf("Total=%d, want %d", got, want)
	}
}

func TestCountAuthFactors_AllThreeKinds(t *testing.T) {
	pool := setupFactorsDB(t)
	uid := insertUserWithPassword(t, pool, "everything@example.com", true)
	insertIdentity(t, pool, uid, "github", "g-1")
	insertPasskey(t, pool, uid)

	f, err := CountAuthFactors(context.Background(), pool, uid, nil, nil)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if got, want := f.Total(), 3; got != want {
		t.Errorf("Total=%d (password=%v, identities=%d, passkeys=%d), want 3",
			got, f.PasswordSet, f.Identities, f.Passkeys)
	}
}

func TestCountAuthFactors_ExcludeIdentity(t *testing.T) {
	pool := setupFactorsDB(t)
	uid := insertUserWithPassword(t, pool, "exclude-id@example.com", false)
	idGH := insertIdentity(t, pool, uid, "github", "gh-9")
	insertIdentity(t, pool, uid, "google", "goog-9")

	f1, _ := CountAuthFactors(context.Background(), pool, uid, nil, nil)
	if f1.Total() != 2 {
		t.Fatalf("baseline Total=%d, want 2", f1.Total())
	}

	f2, err := CountAuthFactors(context.Background(), pool, uid, &idGH, nil)
	if err != nil {
		t.Fatalf("count with exclude: %v", err)
	}
	if f2.Identities != 1 {
		t.Errorf("Identities after exclude = %d, want 1", f2.Identities)
	}
	if f2.Total() != 1 {
		t.Errorf("Total after exclude = %d, want 1", f2.Total())
	}
}

func TestCountAuthFactors_ExcludePasskey(t *testing.T) {
	pool := setupFactorsDB(t)
	uid := insertUserWithPassword(t, pool, "exclude-pk@example.com", false)
	insertIdentity(t, pool, uid, "github", "gh-only")
	_, credID := insertPasskey(t, pool, uid)

	f1, _ := CountAuthFactors(context.Background(), pool, uid, nil, nil)
	if f1.Total() != 2 {
		t.Fatalf("baseline Total=%d, want 2", f1.Total())
	}

	f2, err := CountAuthFactors(context.Background(), pool, uid, nil, credID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if f2.Passkeys != 0 {
		t.Errorf("Passkeys after exclude = %d, want 0", f2.Passkeys)
	}
	if f2.Total() != 1 {
		t.Errorf("Total after exclude = %d, want 1", f2.Total())
	}
}

func TestCountAuthFactors_ExcludeLast_TotalZero(t *testing.T) {
	pool := setupFactorsDB(t)
	uid := insertUserWithPassword(t, pool, "last-id@example.com", false)
	idOnly := insertIdentity(t, pool, uid, "github", "only-one")

	f, err := CountAuthFactors(context.Background(), pool, uid, &idOnly, nil)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if f.Total() != 0 {
		t.Errorf("Total=%d, want 0 (last factor about to be removed)", f.Total())
	}
}

func TestCountAuthFactors_PasswordPlusSingleIdentity_DropIdentity_StillSafe(t *testing.T) {
	pool := setupFactorsDB(t)
	uid := insertUserWithPassword(t, pool, "safe@example.com", true)
	id := insertIdentity(t, pool, uid, "google", "g-safe")

	f, err := CountAuthFactors(context.Background(), pool, uid, &id, nil)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if !f.PasswordSet {
		t.Error("PasswordSet=false (lost!), want true")
	}
	if f.Total() != 1 {
		t.Errorf("Total=%d, want 1 (password remains)", f.Total())
	}
}

func TestCountAuthFactors_NilPool(t *testing.T) {
	f, err := CountAuthFactors(context.Background(), nil, uuid.New(), nil, nil)
	if err != nil {
		t.Fatalf("err on nil pool: %v", err)
	}
	if f.Total() != 0 {
		t.Errorf("expected empty struct for nil pool; got %+v", f)
	}
}
