//go:build integration

// Запуск: EOP_TEST_PG_DSN=postgres://eop:eop_dev@localhost:5432/eop_test \
//   go test -tags=integration ./internal/auth/...

package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/migrate"
	"github.com/eye-of-providence/backend/internal/store"
)

// setupAuthTestApp — Fiber app с MeRoutes + middleware.
func setupAuthTestApp(t *testing.T) (*fiber.App, *pgxpool.Pool, MeService) {
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
	_, _ = pool.Exec(ctx, "TRUNCATE users CASCADE")
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "TRUNCATE users CASCADE")
		pool.Close()
	})

	logger, _ := zap.NewDevelopment()
	app := fiber.New()
	svc := MeService{
		JWTSecret:  "test-secret-32-chars-or-longer-aaaa",
		Pool:       pool,
		EventStore: store.NewMemory(),
		Logger:     logger,
	}
	RegisterMeRoutes(app, svc)
	return app, pool, svc
}

// createTestUser — вставляет user в БД и issue'ит JWT.
func createTestUser(t *testing.T, pool *pgxpool.Pool, secret, email string) (uuid.UUID, string) {
	t.Helper()
	id := uuid.New()
	hash, _ := HashPassword("password123")
	_, err := pool.Exec(context.Background(),
		"INSERT INTO users (id, email, display_name, password_hash) VALUES ($1, $2, $3, $4)",
		id, email, strings.Split(email, "@")[0], hash)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	tv, _ := TokenVersion(context.Background(), pool, id)
	tok, err := IssueJWT(secret, id.String(), email, "test", tv, time.Hour)
	if err != nil {
		t.Fatalf("issue jwt: %v", err)
	}
	return id, tok
}

func doAuth(t *testing.T, app *fiber.App, method, path, token string, body any) (int, []byte) {
	t.Helper()
	var br io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		br = strings.NewReader(string(raw))
	}
	req := httptest.NewRequest(method, path, br)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, raw
}

// --- /v1/me ---

func TestMe_ReturnsProfile(t *testing.T) {
	app, _, svc := setupAuthTestApp(t)
	_, tok := createTestUser(t, svc.Pool, svc.JWTSecret, "me@example.com")

	status, body := doAuth(t, app, "GET", "/v1/me", tok, nil)
	if status != 200 {
		t.Fatalf("status = %d body=%s", status, string(body))
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out["email"] != "me@example.com" {
		t.Errorf("email = %v", out["email"])
	}
	if out["user_id"] == nil {
		t.Error("user_id missing")
	}
}

func TestMe_NoToken_401(t *testing.T) {
	app, _, _ := setupAuthTestApp(t)
	status, _ := doAuth(t, app, "GET", "/v1/me", "", nil)
	if status != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", status)
	}
}

// --- /v1/me/onboarding-status ---

func TestOnboardingStatus_FreshUser(t *testing.T) {
	app, _, svc := setupAuthTestApp(t)
	_, tok := createTestUser(t, svc.Pool, svc.JWTSecret, "fresh@example.com")

	status, body := doAuth(t, app, "GET", "/v1/me/onboarding-status", tok, nil)
	if status != 200 {
		t.Fatalf("status=%d body=%s", status, string(body))
	}
	var out struct {
		TeamsCount int  `json:"teams_count"`
		HasEvent   bool `json:"has_event"`
		Dismissed  bool `json:"dismissed"`
	}
	_ = json.Unmarshal(body, &out)
	if out.TeamsCount != 0 || out.HasEvent || out.Dismissed {
		t.Errorf("fresh user должен иметь все нули: %+v", out)
	}
}

func TestOnboardingDismiss_PersistsToDB(t *testing.T) {
	app, pool, svc := setupAuthTestApp(t)
	uid, tok := createTestUser(t, svc.Pool, svc.JWTSecret, "dismiss@example.com")

	status, _ := doAuth(t, app, "POST", "/v1/me/onboarding/dismiss", tok, nil)
	if status != 200 {
		t.Fatalf("dismiss status=%d", status)
	}

	var dismissedAt *string
	_ = pool.QueryRow(context.Background(),
		"SELECT onboarding_dismissed_at::text FROM users WHERE id=$1", uid).Scan(&dismissedAt)
	if dismissedAt == nil {
		t.Fatal("onboarding_dismissed_at не set")
	}

	// status теперь возвращает dismissed=true
	_, body := doAuth(t, app, "GET", "/v1/me/onboarding-status", tok, nil)
	var out struct {
		Dismissed bool `json:"dismissed"`
	}
	_ = json.Unmarshal(body, &out)
	if !out.Dismissed {
		t.Error("dismissed=false after explicit dismiss call")
	}
}

func TestOnboardingDismiss_Idempotent(t *testing.T) {
	app, pool, svc := setupAuthTestApp(t)
	uid, tok := createTestUser(t, svc.Pool, svc.JWTSecret, "idem@example.com")

	// Первый dismiss
	doAuth(t, app, "POST", "/v1/me/onboarding/dismiss", tok, nil)
	var first *time.Time
	_ = pool.QueryRow(context.Background(),
		"SELECT onboarding_dismissed_at FROM users WHERE id=$1", uid).Scan(&first)

	time.Sleep(50 * time.Millisecond) // чтобы видна была разница, если бы она была

	// Второй dismiss — не должен затереть первый timestamp
	status, _ := doAuth(t, app, "POST", "/v1/me/onboarding/dismiss", tok, nil)
	if status != 200 {
		t.Errorf("second dismiss status = %d", status)
	}
	var second *time.Time
	_ = pool.QueryRow(context.Background(),
		"SELECT onboarding_dismissed_at FROM users WHERE id=$1", uid).Scan(&second)
	if !first.Equal(*second) {
		t.Error("повторный dismiss переписал timestamp — не идемпотентно")
	}
}

// --- /v1/me/locale ---

func TestLocale_PatchesUserRow(t *testing.T) {
	app, pool, svc := setupAuthTestApp(t)
	uid, tok := createTestUser(t, svc.Pool, svc.JWTSecret, "locale@example.com")

	status, _ := doAuth(t, app, "PATCH", "/v1/me/locale", tok, map[string]string{"locale": "kk"})
	if status != 200 {
		t.Fatalf("status = %d", status)
	}

	var locale *string
	_ = pool.QueryRow(context.Background(), "SELECT locale FROM users WHERE id=$1", uid).Scan(&locale)
	if locale == nil || *locale != "kk" {
		t.Errorf("locale = %v, want 'kk'", locale)
	}
}

func TestLocale_RejectsUnsupported(t *testing.T) {
	app, _, svc := setupAuthTestApp(t)
	_, tok := createTestUser(t, svc.Pool, svc.JWTSecret, "bad-locale@example.com")

	status, body := doAuth(t, app, "PATCH", "/v1/me/locale", tok, map[string]string{"locale": "fr"})
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 body=%s", status, string(body))
	}
	var out map[string]string
	_ = json.Unmarshal(body, &out)
	if out["code"] != "invalid_locale" {
		t.Errorf("code = %v, want invalid_locale", out["code"])
	}
}

func TestLocale_RejectsAllSupportedExceptValid(t *testing.T) {
	app, _, svc := setupAuthTestApp(t)
	_, tok := createTestUser(t, svc.Pool, svc.JWTSecret, "all@example.com")

	for _, valid := range []string{"ru", "en", "kk", "es"} {
		status, _ := doAuth(t, app, "PATCH", "/v1/me/locale", tok, map[string]string{"locale": valid})
		if status != 200 {
			t.Errorf("locale %q rejected: status=%d", valid, status)
		}
	}
}
