package ingest

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/ingest/domain"
	"github.com/eye-of-providence/backend/internal/store"
)

const testSecret = "test-secret-32-chars-or-longer-aaaa"

func setupApp(t *testing.T) (*fiber.App, store.EventStore, string) {
	t.Helper()
	st := store.NewMemory()
	logger, _ := zap.NewDevelopment()
	app := fiber.New()
	RegisterRoutes(app, st, logger, testSecret, nil)
	tok, err := auth.IssueJWT(testSecret, "test-user-uuid", "u@example.com", "test", 0, time.Hour)
	if err != nil {
		t.Fatalf("issue jwt: %v", err)
	}
	return app, st, tok
}

func post(t *testing.T, app *fiber.App, tok string, body any) (int, []byte) {
	t.Helper()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/ingest", strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", "application/json")
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, out
}

func TestValidEvent_HappyPath(t *testing.T) {
	e := domain.Event{AppBundle: "code", Source: "ide", Category: "manual", DurationMS: 1000}
	if !domain.ValidEvent(e) {
		t.Fatal("rejected valid event")
	}
}

func TestValidEvent_RejectsMissingFields(t *testing.T) {
	cases := []domain.Event{
		{Source: "ide", Category: "manual"},
		{AppBundle: "code", Category: "manual"},
		{AppBundle: "code", Source: "ide"},
		{AppBundle: "code", Source: "x", Category: "manual"},
		{AppBundle: "code", Source: "ide", Category: "spam"},
		{AppBundle: "code", Source: "ide", Category: "manual", DurationMS: 25 * 60 * 60 * 1000},
	}
	for i, e := range cases {
		if domain.ValidEvent(e) {
			t.Errorf("case %d: invalid event accepted: %+v", i, e)
		}
	}
}

func TestIngest_Unauthorized(t *testing.T) {
	app, _, _ := setupApp(t)
	status, _ := post(t, app, "", map[string]any{"events": []any{}})
	if status != 401 {
		t.Errorf("status = %d, want 401", status)
	}
}

func TestIngest_HappyPath(t *testing.T) {
	app, st, tok := setupApp(t)
	body := map[string]any{
		"events": []map[string]any{
			{"app_bundle": "code", "source": "ide", "category": "manual", "duration_ms": 1000},
			{"app_bundle": "browser", "source": "browser", "category": "ai", "duration_ms": 2000},
		},
	}
	status, raw := post(t, app, tok, body)
	if status != 200 {
		t.Fatalf("status=%d body=%s", status, string(raw))
	}
	var resp response
	_ = json.Unmarshal(raw, &resp)
	if resp.Accepted != 2 || resp.Rejected != 0 {
		t.Errorf("accepted=%d rejected=%d, want 2/0", resp.Accepted, resp.Rejected)
	}

	events, _ := st.ListRecent(t.Context(), "test-user-uuid", 10)
	if len(events) != 2 {
		t.Errorf("store has %d events, want 2", len(events))
	}
}

func TestIngest_RejectsBatchTooLarge(t *testing.T) {
	app, _, tok := setupApp(t)
	events := make([]map[string]any, maxEventsPerBatch+1)
	for i := range events {
		events[i] = map[string]any{"app_bundle": "code", "source": "ide", "category": "manual", "duration_ms": 1000}
	}
	status, _ := post(t, app, tok, map[string]any{"events": events})
	if status != 413 {
		t.Errorf("status = %d, want 413", status)
	}
}

func TestIngest_MixOfValidAndInvalid(t *testing.T) {
	app, _, tok := setupApp(t)
	body := map[string]any{
		"events": []map[string]any{
			{"app_bundle": "code", "source": "ide", "category": "manual", "duration_ms": 1000},
			{"app_bundle": "code", "source": "fake", "category": "manual"},
			{"app_bundle": "code", "source": "ide", "category": "manual", "duration_ms": 1000},
		},
	}
	status, raw := post(t, app, tok, body)
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
	var resp response
	_ = json.Unmarshal(raw, &resp)
	if resp.Accepted != 2 {
		t.Errorf("accepted = %d, want 2", resp.Accepted)
	}
	if resp.Rejected != 1 {
		t.Errorf("rejected = %d, want 1", resp.Rejected)
	}
}

func TestIngest_OverridesUserIDFromToken(t *testing.T) {

	app, st, tok := setupApp(t)
	body := map[string]any{
		"events": []map[string]any{
			{
				"app_bundle":  "code",
				"source":      "ide",
				"category":    "manual",
				"duration_ms": 1000,
				"user_id":     "EVIL-OTHER-USER",
			},
		},
	}
	post(t, app, tok, body)
	events, _ := st.ListRecent(t.Context(), "test-user-uuid", 10)
	if len(events) != 1 || events[0].UserID != "test-user-uuid" {
		t.Errorf("user_id не переписался от токена: events=%+v", events)
	}
}

func TestIngest_InvalidJSON_400(t *testing.T) {
	app, _, tok := setupApp(t)
	req := httptest.NewRequest("POST", "/v1/ingest", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, _ := app.Test(req, -1)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}
