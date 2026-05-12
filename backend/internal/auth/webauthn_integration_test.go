//go:build integration

// Запуск:
//   EOP_TEST_PG_DSN=postgres://eop:eop_dev@localhost:5432/eop_test \
//     go test -tags=integration ./internal/auth/...
//
// Полные WebAuthn ceremony-тесты — register / login / replay — против
// virtualwebauthn-authenticator'а descope/virtualwebauthn v1.0.5.
//
// Источник правды: webauthn.go + .team/qa-testplans/webauthn-register.md +
// webauthn-login.md.
//
// Архитектура fake authenticator'а:
//   - RelyingParty{ID: localhost, Origin: http://localhost:5173} — точно
//     совпадает с newTestWebAuthn() config'ом, иначе RP-validation в
//     go-webauthn отвергнет attestation/assertion.
//   - virtualwebauthn.NewCredential(KeyTypeEC2) — EC2/P-256 keypair (типичный
//     для платформенных passkey'ев: Apple/Google).
//   - UserHandle = userID.Bytes — мы храним 16-byte raw UUID в WebAuthnID,
//     go-webauthn проверит совпадение в FinishLogin.

package auth

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/descope/virtualwebauthn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap/zaptest"

	"github.com/eye-of-providence/backend/internal/migrate"
)

const (
	testWebAuthnRPID     = "localhost"
	testWebAuthnRPName   = "EOP Test"
	testWebAuthnOrigin   = "http://localhost:5173"
	testWebAuthnUserName = "wa-test@example.com"
)

// setupWebAuthnIntegration — DB + Redis (miniredis) + WebAuthnService готовый
// к полному ceremony. Inserts a user row и возвращает её id.
func setupWebAuthnIntegration(t *testing.T) (*WebAuthnService, *pgxpool.Pool, uuid.UUID) {
	t.Helper()
	dsn := os.Getenv("EOP_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("EOP_TEST_PG_DSN not set; skipping WebAuthn integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	if err := migrate.RunPostgres(ctx, dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := pool.Exec(ctx, "TRUNCATE users, user_identities, webauthn_credentials CASCADE"); err != nil {
		t.Fatalf("truncate pre: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			"TRUNCATE users, user_identities, webauthn_credentials CASCADE")
		pool.Close()
	})

	rds := newTestRedis(t)
	wa, err := NewWebAuthnService(testWebAuthnRPID, testWebAuthnRPName, testWebAuthnOrigin, pool, rds, zaptest.NewLogger(t))
	if err != nil {
		t.Fatalf("NewWebAuthnService: %v", err)
	}
	if wa == nil {
		t.Fatal("nil WebAuthnService")
	}

	userID := uuid.New()
	_, err = pool.Exec(ctx,
		"INSERT INTO users (id, email, display_name) VALUES ($1, $2, $3)",
		userID, testWebAuthnUserName, "WA Tester")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return wa, pool, userID
}

// newVirtualRP — config совпадает с newTestWebAuthn() — иначе RP/Origin checks
// в go-webauthn fail'ятся.
func newVirtualRP() virtualwebauthn.RelyingParty {
	return virtualwebauthn.RelyingParty{
		Name:   testWebAuthnRPName,
		ID:     testWebAuthnRPID,
		Origin: testWebAuthnOrigin,
	}
}

// registerCredential — выполняет полный BeginRegistration → FinishRegistration
// flow и возвращает виртуальные credential/authenticator, готовые к login.
//
// nickname опционален; пустая строка → NULL в DB.
func registerCredential(t *testing.T, wa *WebAuthnService, userID uuid.UUID, nickname string) (virtualwebauthn.Authenticator, virtualwebauthn.Credential) {
	t.Helper()
	ctx := context.Background()

	creation, sid, err := wa.BeginRegistration(ctx, userID)
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	if sid == "" {
		t.Fatal("empty session id")
	}

	optionsJSON, err := json.Marshal(creation.Response)
	if err != nil {
		t.Fatalf("marshal options: %v", err)
	}
	attestationOpts, err := virtualwebauthn.ParseAttestationOptions(string(optionsJSON))
	if err != nil {
		t.Fatalf("ParseAttestationOptions: %v", err)
	}

	rp := newVirtualRP()
	authenticator := virtualwebauthn.NewAuthenticator()
	cred := virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)

	attestationResponse := virtualwebauthn.CreateAttestationResponse(rp, authenticator, cred, *attestationOpts)

	if err := wa.FinishRegistration(ctx, userID, sid, []byte(attestationResponse), nickname); err != nil {
		t.Fatalf("FinishRegistration: %v", err)
	}

	// UserHandle == raw UUID bytes — webauthn.go использует userID.Bytes как
	// WebAuthnID, go-webauthn проверяет matching при FinishLogin/DiscoverableLogin.
	uidBytes := userID
	authenticator.Options.UserHandle = uidBytes[:]
	authenticator.AddCredential(cred)
	return authenticator, cred
}

// TestWebAuthnRegister_FullFlow — happy-path register: BeginRegistration →
// virtualwebauthn attestation → FinishRegistration → row в DB.
//
// Источник: webauthn-register.md §1 (happy path).
func TestWebAuthnRegister_FullFlow(t *testing.T) {
	wa, pool, userID := setupWebAuthnIntegration(t)
	ctx := context.Background()

	_, cred := registerCredential(t, wa, userID, "Virtual EC2")

	// Verify row persisted с правильными полями.
	var (
		credentialID []byte
		signCount    int64
		nickname     *string
		transports   string
	)
	err := pool.QueryRow(ctx, `
		SELECT credential_id, sign_count, nickname, transports
		FROM webauthn_credentials WHERE user_id = $1
	`, userID).Scan(&credentialID, &signCount, &nickname, &transports)
	if err != nil {
		t.Fatalf("select credential: %v", err)
	}
	if string(credentialID) != string(cred.ID) {
		t.Errorf("stored credential_id mismatch (len got=%d want=%d)", len(credentialID), len(cred.ID))
	}
	if signCount < 0 {
		t.Errorf("sign_count negative: %d", signCount)
	}
	if nickname == nil || *nickname != "Virtual EC2" {
		got := "<nil>"
		if nickname != nil {
			got = *nickname
		}
		t.Errorf("nickname=%q, want %q", got, "Virtual EC2")
	}
	if transports == "" {
		t.Errorf("transports empty; virtualwebauthn defaults to internal — expected non-empty")
	}
}

// TestWebAuthnLogin_ReplayAttack — переиспользование одного assertion дважды
// отвергается. Защита: webauthn.go сохраняет sessions в Redis SETEX → GETDEL,
// первый FinishLogin потребляет session, второй — fails ("session not found").
//
// Дополнительно проверяем sign_count regression: после успешного login DB
// хранит обновлённый sign_count; повторный assertion с тем же counter (без
// bump'а виртуального authenticator'а) приведёт к CloneWarning в go-webauthn
// (см. authenticator.go:60). Здесь это вторичная защита — primary — single-use
// session — уже отвергает replay.
//
// Источник: webauthn-login.md §5.
func TestWebAuthnLogin_ReplayAttack(t *testing.T) {
	wa, pool, userID := setupWebAuthnIntegration(t)
	ctx := context.Background()

	authenticator, cred := registerCredential(t, wa, userID, "replay-test")

	email := testWebAuthnUserName

	// === Login attempt #1: должен пройти. ===
	assertion1, sid1, err := wa.BeginLogin(ctx, &email)
	if err != nil {
		t.Fatalf("BeginLogin #1: %v", err)
	}
	options1JSON, err := json.Marshal(assertion1.Response)
	if err != nil {
		t.Fatalf("marshal options #1: %v", err)
	}
	parsedOpts1, err := virtualwebauthn.ParseAssertionOptions(string(options1JSON))
	if err != nil {
		t.Fatalf("ParseAssertionOptions #1: %v", err)
	}
	rp := newVirtualRP()
	body1 := virtualwebauthn.CreateAssertionResponse(rp, authenticator, cred, *parsedOpts1)

	gotUserID, err := wa.FinishLogin(ctx, sid1, []byte(body1))
	if err != nil {
		t.Fatalf("FinishLogin #1: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("FinishLogin #1 returned userID=%s, want %s", gotUserID, userID)
	}

	// === Replay attempt: same (sid1, body1) → session уже GETDEL'd. ===
	if _, err := wa.FinishLogin(ctx, sid1, []byte(body1)); err == nil {
		t.Fatal("replay with reused session_id succeeded; single-use protection broken")
	}

	// === Sign-count regression: новый BeginLogin, но cred.Counter оставлен
	// без bump'а — virtualwebauthn пишет тот же sign_count в authData. После
	// previous success DB хранит sign_count=0 (initial). go-webauthn fires
	// CloneWarning при authDataCount <= storedCount (см. authenticator.go:60).
	// FinishLogin не возвращает error по CloneWarning, но мы убеждаемся, что
	// замёт sign_count в DB не двинулся назад. ===
	var storedSignCount int64
	if err := pool.QueryRow(ctx, `SELECT sign_count FROM webauthn_credentials WHERE user_id = $1`, userID).Scan(&storedSignCount); err != nil {
		t.Fatalf("select sign_count: %v", err)
	}

	assertion2, sid2, err := wa.BeginLogin(ctx, &email)
	if err != nil {
		t.Fatalf("BeginLogin #2: %v", err)
	}
	options2JSON, _ := json.Marshal(assertion2.Response)
	parsedOpts2, err := virtualwebauthn.ParseAssertionOptions(string(options2JSON))
	if err != nil {
		t.Fatalf("ParseAssertionOptions #2: %v", err)
	}
	// КРИТИЧЕСКОЕ: не bump'аем cred.Counter — симулируем cloned authenticator.
	body2 := virtualwebauthn.CreateAssertionResponse(rp, authenticator, cred, *parsedOpts2)

	// FinishLogin успешен (go-webauthn не fail'ит — только flag CloneWarning).
	// Но stored sign_count в DB не должен регрессировать.
	if _, err := wa.FinishLogin(ctx, sid2, []byte(body2)); err != nil {
		// Если когда-нибудь webauthn.go начнёт propagate CloneWarning →
		// принимаем error. Тест не должен false-fail'ить в этом случае.
		t.Logf("FinishLogin #2 errored (likely CloneWarning surfacing): %v", err)
	}

	var afterSignCount int64
	if err := pool.QueryRow(ctx, `SELECT sign_count FROM webauthn_credentials WHERE user_id = $1`, userID).Scan(&afterSignCount); err != nil {
		t.Fatalf("select sign_count after: %v", err)
	}
	if afterSignCount < storedSignCount {
		t.Errorf("sign_count regressed: %d → %d (anti-clone protection failed)", storedSignCount, afterSignCount)
	}
}
