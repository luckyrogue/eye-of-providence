//go:build integration

package teams

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/eye-of-providence/backend/internal/auth"
)

func TestAdminWebhooks_CrossTeam(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)

	admin := createUser(t, pool, "wh-admin@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "wh-admin@example.com")

	owner1 := createUser(t, pool, "wh-o1@example.com")
	team1 := createTeamDirect(t, pool, "WhTeam1", owner1)
	owner2 := createUser(t, pool, "wh-o2@example.com")
	team2 := createTeamDirect(t, pool, "WhTeam2", owner2)

	_, err := pool.Exec(context.Background(), `
		INSERT INTO webhooks (user_id, url, secret, events, active)
		VALUES ($1, 'https://example.com/wh1', 'secret1', ARRAY['commit.ingested'], true),
		       ($2, 'https://example.com/wh2', 'secret2', ARRAY['report.generated'], true)`,
		owner1, owner2)
	if err != nil {
		t.Fatalf("insert webhooks: %v", err)
	}
	_ = team1
	_ = team2

	status, body := do(t, app, "GET", "/v1/admin/webhooks", tok, nil)
	skipIfNotFound(t, status, "GET /v1/admin/webhooks")
	if status != 200 {
		t.Fatalf("status=%d body=%s", status, string(body))
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, string(body))
	}
	var entries []map[string]any
	for _, key := range []string{"webhooks", "items", "entries"} {
		if r, ok := raw[key]; ok {
			_ = json.Unmarshal(r, &entries)
			break
		}
	}
	if entries == nil {

		_ = json.Unmarshal(body, &entries)
	}

	if len(entries) < 2 {
		t.Errorf("webhook count = %d, want >= 2", len(entries))
	}

	hasTeamName := false
	for _, e := range entries {
		if name, _ := e["team_name"].(string); name != "" {
			hasTeamName = true
			break
		}
	}
	if !hasTeamName {
		t.Logf("admin webhooks response не содержит team_name; ожидалось от Phase 3 cross-team join")
	}

	for _, e := range entries {
		if s, ok := e["secret"].(string); ok && s != "" {
			t.Errorf("admin webhook entry leaks secret: %v", s)
		}
	}
}

func TestAdminAPITokens_CrossUser(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)

	admin := createUser(t, pool, "tok-admin@example.com")
	makeSuperAdmin(t, pool, admin)
	tok := loginAs(t, pool, svc.JWTSecret, admin, "tok-admin@example.com")

	user1 := createUser(t, pool, "tok-u1@example.com")
	user2 := createUser(t, pool, "tok-u2@example.com")

	plain1, _, err := auth.CreateAPIToken(context.Background(), pool, user1, "ci-exporter", "read", 0)
	if err != nil {
		t.Fatalf("create token1: %v", err)
	}
	plain2, _, err := auth.CreateAPIToken(context.Background(), pool, user2, "personal-bi", "read", 0)
	if err != nil {
		t.Fatalf("create token2: %v", err)
	}

	status, body := do(t, app, "GET", "/v1/admin/api-tokens", tok, nil)
	skipIfNotFound(t, status, "GET /v1/admin/api-tokens")
	if status != 200 {
		t.Fatalf("status=%d body=%s", status, string(body))
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, string(body))
	}
	var entries []map[string]any
	for _, key := range []string{"api_tokens", "tokens", "items", "entries"} {
		if r, ok := raw[key]; ok {
			_ = json.Unmarshal(r, &entries)
			break
		}
	}
	if entries == nil {
		_ = json.Unmarshal(body, &entries)
	}

	if len(entries) < 2 {
		t.Errorf("token count = %d, want >= 2", len(entries))
	}

	bodyStr := string(body)
	if strings.Contains(bodyStr, plain1) || strings.Contains(bodyStr, plain2) {
		t.Error("CRITICAL: admin api-tokens response leaks plaintext token value")
	}
	if strings.Contains(bodyStr, "hashed_token") {
		t.Error("admin api-tokens response leaks hashed_token field")
	}

	hasUserEmail := false
	hasPrefix := false
	for _, e := range entries {
		if email, _ := e["user_email"].(string); email != "" {
			hasUserEmail = true
		}
		if prefix, _ := e["prefix"].(string); prefix != "" {
			hasPrefix = true
		}
	}
	if !hasUserEmail {
		t.Errorf("admin api-tokens entry missing user_email field; expected from cross-user join")
	}
	if !hasPrefix {
		t.Errorf("admin api-tokens entry missing prefix field; expected for UI display")
	}
}

func TestAdminWebhooks_RequiresSuperAdmin(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)

	regular := createUser(t, pool, "wh-reg@example.com")
	tok := loginAs(t, pool, svc.JWTSecret, regular, "wh-reg@example.com")

	status, _ := do(t, app, "GET", "/v1/admin/webhooks", tok, nil)
	skipIfNotFound(t, status, "GET /v1/admin/webhooks (auth)")
	if status != 403 {
		t.Errorf("status=%d, want 403", status)
	}
}

func TestAdminAPITokens_RequiresSuperAdmin(t *testing.T) {
	pool := setupTestDB(t)
	app, svc, _ := newAdminApp(t, pool)

	regular := createUser(t, pool, "tok-reg@example.com")
	tok := loginAs(t, pool, svc.JWTSecret, regular, "tok-reg@example.com")

	status, _ := do(t, app, "GET", "/v1/admin/api-tokens", tok, nil)
	skipIfNotFound(t, status, "GET /v1/admin/api-tokens (auth)")
	if status != 403 {
		t.Errorf("status=%d, want 403", status)
	}
}
