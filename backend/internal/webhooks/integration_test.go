//go:build integration

package webhooks

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/migrate"
)

func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("EOP_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("EOP_TEST_PG_DSN not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate.RunPostgres(ctx, dsn); err != nil {
		t.Fatal(err)
	}
	_, _ = pool.Exec(ctx, "TRUNCATE users, webhooks RESTART IDENTITY CASCADE")
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "TRUNCATE users, webhooks RESTART IDENTITY CASCADE")
		pool.Close()
	})
	return pool
}

func makeUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	hash, _ := auth.HashPassword("dev123456")
	_, err := pool.Exec(context.Background(),
		"INSERT INTO users (id, email, display_name, password_hash) VALUES ($1, $2, $3, $4)",
		id, id.String()+"@example.com", "T", hash)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestCreateWebhook_BadURL(t *testing.T) {
	pool := setupDB(t)
	uid := makeUser(t, pool)
	logger, _ := zap.NewDevelopment()
	svc := New(pool, logger)
	_, _, err := svc.Create(context.Background(), uid, "ftp://x", []string{EventCommitIngested}, "")
	if err == nil {
		t.Error("expected error for non-http URL")
	}
}

func TestCreateWebhook_InvalidEvent(t *testing.T) {
	pool := setupDB(t)
	uid := makeUser(t, pool)
	svc := New(pool, zap.NewNop())
	_, _, err := svc.Create(context.Background(), uid, "https://example.com/h", []string{"random.event"}, "")
	if err == nil {
		t.Error("expected error for unknown event")
	}
}

func TestCreateAndList(t *testing.T) {
	pool := setupDB(t)
	uid := makeUser(t, pool)
	svc := New(pool, zap.NewNop())
	secret, hook, err := svc.Create(context.Background(), uid, "https://example.com/h",
		[]string{EventCommitIngested, EventReportGenerated}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(secret) < 32 {
		t.Errorf("secret too short: %s", secret)
	}
	if hook.URL != "https://example.com/h" || len(hook.Events) != 2 {
		t.Errorf("hook=%+v", hook)
	}
	list, err := svc.List(context.Background(), uid)
	if err != nil || len(list) != 1 {
		t.Fatalf("list err=%v, len=%d", err, len(list))
	}
}

func TestDispatch_HappyPath(t *testing.T) {
	pool := setupDB(t)
	uid := makeUser(t, pool)

	var (
		gotMu       sync.Mutex
		gotBody     []byte
		gotSig      string
		gotEvent    string
		hits        atomic.Int32
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMu.Lock()
		defer gotMu.Unlock()
		body, _ := io.ReadAll(r.Body)
		gotBody = body
		gotSig = r.Header.Get("X-EoP-Signature")
		gotEvent = r.Header.Get("X-EoP-Event")
		hits.Add(1)
		w.WriteHeader(200)
	}))
	defer server.Close()

	svc := New(pool, zap.NewNop())
	secret, _, err := svc.Create(context.Background(), uid, server.URL, []string{EventCommitIngested}, "")
	if err != nil {
		t.Fatal(err)
	}

	// Synchronous-ish: вызываем dispatchSync вместо Dispatch чтобы не race.
	svc.dispatchSync(uid, EventCommitIngested, map[string]any{"sha": "abc"})

	// Wait short for goroutine; dispatchSync блокирует do all sends.
	for i := 0; i < 20 && hits.Load() == 0; i++ {
		time.Sleep(50 * time.Millisecond)
	}
	if hits.Load() != 1 {
		t.Fatalf("expected 1 delivery, got %d", hits.Load())
	}

	gotMu.Lock()
	defer gotMu.Unlock()
	if gotEvent != EventCommitIngested {
		t.Errorf("event header=%q", gotEvent)
	}
	if !VerifySignature(secret, gotBody, gotSig) {
		t.Errorf("signature verification failed: sig=%s body=%s", gotSig, string(gotBody))
	}

	var payload map[string]any
	_ = json.Unmarshal(gotBody, &payload)
	if payload["event"] != EventCommitIngested {
		t.Errorf("payload.event=%v", payload["event"])
	}
	data, _ := payload["data"].(map[string]any)
	if data["sha"] != "abc" {
		t.Errorf("payload.data.sha=%v", data["sha"])
	}

	// last_delivery_at + last_status должны обновиться.
	var lastStatus *int
	_ = pool.QueryRow(context.Background(),
		"SELECT last_status FROM webhooks WHERE user_id=$1", uid).Scan(&lastStatus)
	if lastStatus == nil || *lastStatus != 200 {
		t.Errorf("last_status=%v, want 200", lastStatus)
	}
}

func TestDispatch_OnlyMatchingEvents(t *testing.T) {
	pool := setupDB(t)
	uid := makeUser(t, pool)

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(200)
	}))
	defer server.Close()

	svc := New(pool, zap.NewNop())
	_, _, err := svc.Create(context.Background(), uid, server.URL, []string{EventReportGenerated}, "")
	if err != nil {
		t.Fatal(err)
	}
	svc.dispatchSync(uid, EventCommitIngested, map[string]any{}) // не matching event
	time.Sleep(100 * time.Millisecond)
	if hits.Load() != 0 {
		t.Errorf("delivered %d times, want 0 (event not subscribed)", hits.Load())
	}
}

func TestDispatch_5xxRetry(t *testing.T) {
	pool := setupDB(t)
	uid := makeUser(t, pool)

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(500) // всегда fail
	}))
	defer server.Close()

	svc := New(pool, zap.NewNop())
	// fast HTTP timeout чтобы тест не висел
	svc.HTTPClient = &http.Client{Timeout: 1 * time.Second}
	_, _, err := svc.Create(context.Background(), uid, server.URL, []string{EventCommitIngested}, "")
	if err != nil {
		t.Fatal(err)
	}

	// dispatchSync включает 4 попытки (delays 0/1s/3s/9s). Ждать 13s + buffer.
	done := make(chan struct{})
	go func() {
		svc.dispatchSync(uid, EventCommitIngested, map[string]any{})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("dispatchSync did not return")
	}

	if hits.Load() != 4 {
		t.Errorf("expected 4 attempts (1 + 3 retries), got %d", hits.Load())
	}

	var lastStatus *int
	_ = pool.QueryRow(context.Background(),
		"SELECT last_status FROM webhooks WHERE user_id=$1", uid).Scan(&lastStatus)
	if lastStatus == nil || *lastStatus != 500 {
		t.Errorf("last_status=%v, want 500", lastStatus)
	}
}

func TestDispatch_SlackFormat(t *testing.T) {
	pool := setupDB(t)
	uid := makeUser(t, pool)

	var (
		gotMu   sync.Mutex
		gotBody []byte
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMu.Lock()
		defer gotMu.Unlock()
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(200)
	}))
	defer server.Close()

	svc := New(pool, zap.NewNop())
	_, _, err := svc.Create(context.Background(), uid, server.URL, []string{EventCommitIngested}, "slack")
	if err != nil {
		t.Fatal(err)
	}

	svc.dispatchSync(uid, EventCommitIngested, map[string]any{
		"sha":     "abc1234567",
		"message": "fix bug",
		"branch":  "main",
	})

	gotMu.Lock()
	defer gotMu.Unlock()
	var payload map[string]any
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("body not JSON: %s", string(gotBody))
	}
	// Slack contract: верхний уровень — text + blocks, без event/data.
	if _, ok := payload["text"]; !ok {
		t.Error("Slack payload missing text")
	}
	if _, ok := payload["blocks"]; !ok {
		t.Error("Slack payload missing blocks")
	}
	if _, ok := payload["event"]; ok {
		t.Error("Slack payload should not have event field (raw shape leaked)")
	}
}

func TestDelete_OtherUser(t *testing.T) {
	pool := setupDB(t)
	uid1 := makeUser(t, pool)
	uid2 := makeUser(t, pool)
	svc := New(pool, zap.NewNop())
	_, hook, _ := svc.Create(context.Background(), uid1, "https://example.com/h", []string{EventCommitIngested}, "")
	ok, err := svc.Delete(context.Background(), uid2, hook.ID)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("delete succeeded across users")
	}
}
