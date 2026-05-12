//go:build integration

// Запуск:
//   EOP_TEST_PG_DSN=postgres://eop:eop_dev@localhost:5432/eop_test \
//     go test -tags=integration ./internal/auth/...
//
// Phase 2 endpoint tests — /v1/me/identities GET/DELETE.
//
// Backend status (2026-05-12): endpoints не зарегистрированы. Тесты ниже —
// scaffold; они скипаются с понятным сообщением, пока backend-channel не
// добавит routes в RegisterMeRoutes (см. .team/backend-progress.md "Phase 2 in
// flight").
//
// Источник правды: .team/qa-testplans/identity-linking.md.

package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/migrate"
	"github.com/eye-of-providence/backend/internal/store"
)

func setupIdentitiesApp(t *testing.T) (*fiber.App, *pgxpool.Pool, MeService) {
	t.Helper()
	dsn := os.Getenv("EOP_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("EOP_TEST_PG_DSN not set; skipping identities integration test")
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
	_, _ = pool.Exec(ctx, "TRUNCATE users, user_identities, webauthn_credentials CASCADE")
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "TRUNCATE users, user_identities, webauthn_credentials CASCADE")
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

func doIdent(t *testing.T, app *fiber.App, method, path, token string) (int, []byte) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
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

func makeUserAndToken(t *testing.T, pool *pgxpool.Pool, email, secret string, hasPassword bool) (uuid.UUID, string) {
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
	tv, _ := TokenVersion(context.Background(), pool, id)
	tok, err := IssueJWT(secret, id.String(), email, "test", tv, time.Hour)
	if err != nil {
		t.Fatalf("issue jwt: %v", err)
	}
	return id, tok
}

// ---- Scaffold tests: they probe the route ahead of implementation. ----

// TestListIdentities_OwnUserOnly — пользователь видит только свои identities,
// не чужие. Базовый authz check.
func TestListIdentities_OwnUserOnly(t *testing.T) {
	app, pool, svc := setupIdentitiesApp(t)
	uidA, tokA := makeUserAndToken(t, pool, "a@example.com", svc.JWTSecret, false)
	uidB, _ := makeUserAndToken(t, pool, "b@example.com", svc.JWTSecret, false)
	insertIdentity(t, pool, uidA, "github", "gh-A")
	insertIdentity(t, pool, uidB, "google", "go-B")

	status, body := doIdent(t, app, "GET", "/v1/me/identities", tokA)
	if status == http.StatusNotFound {
		t.Skipf("GET /v1/me/identities not registered yet (404). Implementation pending. Body=%s", string(body))
	}
	if status != http.StatusOK {
		t.Fatalf("status=%d body=%s", status, string(body))
	}
	var out struct {
		Identities []struct {
			Provider string `json:"provider"`
			Subject  string `json:"subject"`
		} `json:"identities"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode: %v body=%s", err, string(body))
	}
	if len(out.Identities) != 1 {
		t.Fatalf("count=%d, want 1 (user A own)", len(out.Identities))
	}
	if out.Identities[0].Provider != "github" || out.Identities[0].Subject != "gh-A" {
		t.Errorf("unexpected identity: %+v", out.Identities[0])
	}
}

// TestDeleteIdentity_LastFactorRejected — 4a из identity-linking.md: 1 identity,
// нет password, нет passkey → DELETE отдаёт 400 last_auth_factor (или 409 по
// договорённости — handler выбирает code, тест допускает оба).
func TestDeleteIdentity_LastFactorRejected(t *testing.T) {
	app, pool, svc := setupIdentitiesApp(t)
	uid, tok := makeUserAndToken(t, pool, "lock@example.com", svc.JWTSecret, false)
	identID := insertIdentity(t, pool, uid, "github", "only")

	status, body := doIdent(t, app, "DELETE", "/v1/me/identities/"+identID.String(), tok)
	if status == http.StatusNotFound {
		t.Skipf("DELETE /v1/me/identities/:id not registered yet. Body=%s", string(body))
	}
	if status != http.StatusBadRequest && status != http.StatusConflict {
		t.Fatalf("status=%d body=%s, want 400 or 409 (last_auth_factor)", status, string(body))
	}
	// Code должен быть стабильным машинно-парсимым строкой — frontend по нему
	// показывает CTA "set a password first".
	var out map[string]string
	if err := json.Unmarshal(body, &out); err == nil {
		if out["code"] != "" && out["code"] != "last_auth_factor" {
			t.Errorf("error code = %q, want last_auth_factor", out["code"])
		}
	}

	// Row не должна быть удалена.
	var cnt int
	_ = pool.QueryRow(context.Background(),
		"SELECT count(*) FROM user_identities WHERE id=$1", identID).Scan(&cnt)
	if cnt != 1 {
		t.Errorf("identity row gone (%d remain); should have been protected", cnt)
	}
}

// TestDeleteIdentity_OtherFactorsExist_Succeeds — у юзера есть password +
// 1 identity. DELETE identity → 204, row исчезла, password остаётся.
func TestDeleteIdentity_OtherFactorsExist_Succeeds(t *testing.T) {
	app, pool, svc := setupIdentitiesApp(t)
	uid, tok := makeUserAndToken(t, pool, "ok@example.com", svc.JWTSecret, true)
	identID := insertIdentity(t, pool, uid, "google", "g-ok")

	status, body := doIdent(t, app, "DELETE", "/v1/me/identities/"+identID.String(), tok)
	if status == http.StatusNotFound {
		t.Skipf("DELETE /v1/me/identities/:id not registered yet. Body=%s", string(body))
	}
	if status != http.StatusNoContent && status != http.StatusOK {
		t.Fatalf("status=%d body=%s, want 204 or 200", status, string(body))
	}
	var cnt int
	_ = pool.QueryRow(context.Background(),
		"SELECT count(*) FROM user_identities WHERE id=$1", identID).Scan(&cnt)
	if cnt != 0 {
		t.Errorf("row remains (%d), expected deletion", cnt)
	}
}

// TestDeleteIdentity_NotOwn — попытка удалить чужую identity → 404 или 403,
// не 200. Защита от IDOR.
func TestDeleteIdentity_NotOwn(t *testing.T) {
	app, pool, svc := setupIdentitiesApp(t)
	uidA, tokA := makeUserAndToken(t, pool, "a@example.com", svc.JWTSecret, true)
	uidB, _ := makeUserAndToken(t, pool, "b@example.com", svc.JWTSecret, true)
	_ = uidA
	identB := insertIdentity(t, pool, uidB, "github", "b-only")

	status, body := doIdent(t, app, "DELETE", "/v1/me/identities/"+identB.String(), tokA)
	if status == http.StatusNotFound {
		// 404 — корректный negative ответ (не leak'аем существование).
		return
	}
	if status == http.StatusForbidden {
		// 403 — тоже корректно.
		return
	}
	if status >= 200 && status < 300 {
		t.Fatalf("IDOR: deleted other user's identity (status=%d). body=%s", status, string(body))
	}
	// Любой другой код — flag for review.
	t.Logf("unexpected status %d (acceptable: 403/404). body=%s", status, string(body))
}

// TestDeleteIdentity_Unauth — без JWT → 401.
func TestDeleteIdentity_Unauth(t *testing.T) {
	app, _, _ := setupIdentitiesApp(t)
	status, _ := doIdent(t, app, "DELETE", "/v1/me/identities/"+uuid.New().String(), "")
	if status != http.StatusUnauthorized {
		t.Errorf("status=%d, want 401", status)
	}
}
