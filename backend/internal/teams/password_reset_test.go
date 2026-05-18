//go:build integration

package teams

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"github.com/eye-of-providence/backend/internal/auth"
)

// newResetToken / hashResetToken — inline-копии хелперов из старого
// `password_reset.go`. Сам файл переехал в `auth/passwordresetapp/`
// после refactor'а и перестал экспортировать эти функции. Тесты
// сидят на интеграционном уровне (вставляют hash в `password_resets`
// через raw SQL), поэтому проще держать helpers здесь, чем тащить
// passwordresetapp.TokenGenerator зависимость в integration suite.
func newResetToken() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	tok := hex.EncodeToString(b)
	return tok, hashResetToken(tok), nil
}

func hashResetToken(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func TestNewResetToken_DistinctTokens(t *testing.T) {
	tok1, hash1, err := newResetToken()
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	tok2, hash2, err := newResetToken()
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if tok1 == tok2 {
		t.Fatal("tokens collided — random source broken")
	}
	if hash1 == hash2 {
		t.Fatal("hashes collided")
	}
	if len(tok1) != 64 {
		t.Errorf("token len = %d, want 64 (32 bytes hex)", len(tok1))
	}
	if len(hash1) != 64 {
		t.Errorf("hash len = %d, want 64 (sha256 hex)", len(hash1))
	}
}

func TestHashResetToken_Deterministic(t *testing.T) {
	h1 := hashResetToken("abc123")
	h2 := hashResetToken("abc123")
	if h1 != h2 {
		t.Fatal("hash должен быть детерминированным")
	}
	h3 := hashResetToken("abc124")
	if h1 == h3 {
		t.Fatal("разные input'ы дали одинаковый hash")
	}
}

func TestForgotPassword_ExistingEmail_Returns200_AndStoresToken(t *testing.T) {
	pool := setupTestDB(t)
	app, _ := newTestApp(t, pool)

	uid := createUser(t, pool, "user@example.com")

	status, _ := do(t, app, "POST", "/v1/auth/forgot-password", "", map[string]string{"email": "user@example.com"})
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}

	var count int
	_ = pool.QueryRow(context.Background(),
		"SELECT count(*) FROM password_resets WHERE user_id = $1", uid).Scan(&count)
	if count != 1 {
		t.Errorf("password_resets row count = %d, want 1", count)
	}
}

func TestForgotPassword_NonExistingEmail_StillReturns200(t *testing.T) {
	pool := setupTestDB(t)
	app, _ := newTestApp(t, pool)

	status, _ := do(t, app, "POST", "/v1/auth/forgot-password", "", map[string]string{"email": "ghost@example.com"})

	if status != 200 {
		t.Fatalf("status = %d, want 200 (даже для несуществующего email)", status)
	}

	var count int
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM password_resets").Scan(&count)
	if count != 0 {
		t.Errorf("password_resets created для несуществующего email: %d", count)
	}
}

func TestForgotPassword_MalformedEmail_StillReturns200(t *testing.T) {
	pool := setupTestDB(t)
	app, _ := newTestApp(t, pool)

	status, _ := do(t, app, "POST", "/v1/auth/forgot-password", "", map[string]string{"email": "not-an-email"})
	if status != 200 {
		t.Fatalf("status = %d, want 200 (no info leak)", status)
	}
}

func TestResetPassword_HappyPath(t *testing.T) {
	pool := setupTestDB(t)
	app, _ := newTestApp(t, pool)

	uid := createUser(t, pool, "reset-me@example.com")

	token, hash, err := newResetToken()
	if err != nil {
		t.Fatalf("new token: %v", err)
	}
	expires := time.Now().Add(1 * time.Hour)
	_, err = pool.Exec(context.Background(),
		"INSERT INTO password_resets (token_hash, user_id, expires_at) VALUES ($1, $2, $3)",
		hash, uid, expires)
	if err != nil {
		t.Fatalf("insert reset: %v", err)
	}

	status, _ := do(t, app, "POST", "/v1/auth/reset-password", "", map[string]string{
		"token":    token,
		"password": "newpass-secure-123",
	})
	if status != 200 {
		t.Fatalf("reset status = %d, want 200", status)
	}

	user, err := auth.FindUserByEmail(context.Background(), pool, "reset-me@example.com")
	if err != nil {
		t.Fatalf("find user: %v", err)
	}
	if auth.VerifyPassword(user.PasswordHash, "password123") {
		t.Error("старый пароль ещё валидный — должен был замениться")
	}
	if !auth.VerifyPassword(user.PasswordHash, "newpass-secure-123") {
		t.Error("новый пароль не работает")
	}

	tv, _ := auth.TokenVersion(context.Background(), pool, uid)
	if tv == 0 {
		t.Error("token_version не bump'нулся после reset")
	}

	var usedAt *time.Time
	_ = pool.QueryRow(context.Background(),
		"SELECT used_at FROM password_resets WHERE token_hash = $1", hash).Scan(&usedAt)
	if usedAt == nil {
		t.Error("used_at не set после reset")
	}
}

func TestResetPassword_ExpiredToken(t *testing.T) {
	pool := setupTestDB(t)
	app, _ := newTestApp(t, pool)

	uid := createUser(t, pool, "expired@example.com")

	token, hash, _ := newResetToken()
	expired := time.Now().Add(-1 * time.Hour)
	_, err := pool.Exec(context.Background(),
		"INSERT INTO password_resets (token_hash, user_id, expires_at) VALUES ($1, $2, $3)",
		hash, uid, expired)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	status, body := do(t, app, "POST", "/v1/auth/reset-password", "", map[string]string{
		"token":    token,
		"password": "newpass-secure-123",
	})
	if status != 400 {
		t.Fatalf("status = %d body=%s, want 400", status, string(body))
	}
}

func TestResetPassword_UsedToken_Refused(t *testing.T) {
	pool := setupTestDB(t)
	app, _ := newTestApp(t, pool)

	uid := createUser(t, pool, "used-token@example.com")

	token, hash, _ := newResetToken()
	now := time.Now()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO password_resets (token_hash, user_id, expires_at, used_at) VALUES ($1, $2, $3, $4)",
		hash, uid, now.Add(time.Hour), now)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	status, _ := do(t, app, "POST", "/v1/auth/reset-password", "", map[string]string{
		"token":    token,
		"password": "newpass-secure-123",
	})
	if status != 400 {
		t.Fatalf("used token принят: status = %d, want 400", status)
	}
}

func TestResetPassword_InvalidToken_Refused(t *testing.T) {
	pool := setupTestDB(t)
	app, _ := newTestApp(t, pool)

	status, _ := do(t, app, "POST", "/v1/auth/reset-password", "", map[string]string{
		"token":    "not-a-real-token-just-random-string-of-similar-length-aaaaaaaaaa",
		"password": "newpass-secure-123",
	})
	if status != 400 {
		t.Fatalf("invalid token принят: status = %d", status)
	}
}

func TestResetPassword_ShortPassword_Refused(t *testing.T) {
	pool := setupTestDB(t)
	app, _ := newTestApp(t, pool)

	uid := createUser(t, pool, "short-pass@example.com")
	token, hash, _ := newResetToken()
	_, _ = pool.Exec(context.Background(),
		"INSERT INTO password_resets (token_hash, user_id, expires_at) VALUES ($1, $2, $3)",
		hash, uid, time.Now().Add(time.Hour))

	status, _ := do(t, app, "POST", "/v1/auth/reset-password", "", map[string]string{
		"token":    token,
		"password": "short",
	})
	if status != http.StatusBadRequest {
		t.Errorf("short password принят: status = %d", status)
	}
}
