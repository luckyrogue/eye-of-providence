//go:build integration

package teams

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/migrate"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
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
	truncateAll(t, pool)
	t.Cleanup(func() {
		truncateAll(t, pool)
		pool.Close()
	})
	return pool
}

func truncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"TRUNCATE users, teams, team_members, team_invites, team_payments, projects, commits, reports, devices, consent, api_tokens RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func newTestApp(t *testing.T, pool *pgxpool.Pool) (*fiber.App, Service) {
	t.Helper()
	app := fiber.New()
	logger, _ := zap.NewDevelopment()
	svc := Service{
		Pool:          pool,
		JWTSecret:     "test-secret-32-chars-or-longer-aaaa",
		Logger:        logger,
		InviteOnly:    true,
		BetaTeamLimit: 3,
	}
	RegisterRoutes(app, svc)
	// /v1/auth/{forgot,reset}-password переехали из teams в auth
	// после refactor'а — wire'им вручную чтобы integration suite
	// password_reset_test.go не получал 401 на свои POST'ы.
	auth.RegisterPasswordResetRoutes(app, auth.PasswordResetService{
		Pool:      pool,
		Logger:    logger,
		PublicURL: "http://test.local",
	})
	return app, svc
}

func createUser(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	id := uuid.New()
	_, err = pool.Exec(context.Background(),
		"INSERT INTO users (id, email, display_name, password_hash) VALUES ($1, $2, $3, $4)",
		id, email, strings.Split(email, "@")[0], hash)
	if err != nil {
		t.Fatalf("createUser: %v", err)
	}
	return id
}

func makeSuperAdmin(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"UPDATE users SET global_role='super_admin' WHERE id=$1", userID)
	if err != nil {
		t.Fatalf("makeSuperAdmin: %v", err)
	}
}

func loginAs(t *testing.T, pool *pgxpool.Pool, secret string, userID uuid.UUID, email string) string {
	t.Helper()
	tv, err := auth.TokenVersion(context.Background(), pool, userID)
	if err != nil {
		t.Fatalf("token version: %v", err)
	}
	tok, err := auth.IssueJWT(secret, userID.String(), email, "test", tv, time.Hour)
	if err != nil {
		t.Fatalf("issue jwt: %v", err)
	}
	return tok
}

func createTeamDirect(t *testing.T, pool *pgxpool.Pool, name string, ownerID uuid.UUID) uuid.UUID {
	t.Helper()
	teamID := uuid.New()
	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer tx.Rollback(context.Background())
	_, err = tx.Exec(context.Background(),
		"INSERT INTO teams (id, name, plan, created_by) VALUES ($1, $2, 'free', $3)",
		teamID, name, ownerID)
	if err != nil {
		t.Fatalf("insert team: %v", err)
	}
	_, err = tx.Exec(context.Background(),
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'owner')",
		teamID, ownerID)
	if err != nil {
		t.Fatalf("insert member: %v", err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("commit: %v", err)
	}
	return teamID
}

func do(t *testing.T, app *fiber.App, method, path, token string, body any) (int, []byte) {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		bodyReader = strings.NewReader(string(raw))
	}
	req := httptest.NewRequest(method, path, bodyReader)
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
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, raw
}

func jsonField(t *testing.T, raw []byte, key string) any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal %q: %v", string(raw), err)
	}
	return m[key]
}
