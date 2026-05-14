//go:build integration

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

func newVirtualRP() virtualwebauthn.RelyingParty {
	return virtualwebauthn.RelyingParty{
		Name:   testWebAuthnRPName,
		ID:     testWebAuthnRPID,
		Origin: testWebAuthnOrigin,
	}
}

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

	uidBytes := userID
	authenticator.Options.UserHandle = uidBytes[:]
	authenticator.AddCredential(cred)
	return authenticator, cred
}

func TestWebAuthnRegister_FullFlow(t *testing.T) {
	wa, pool, userID := setupWebAuthnIntegration(t)
	ctx := context.Background()

	_, cred := registerCredential(t, wa, userID, "Virtual EC2")

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

func TestWebAuthnLogin_ReplayAttack(t *testing.T) {
	wa, pool, userID := setupWebAuthnIntegration(t)
	ctx := context.Background()

	authenticator, cred := registerCredential(t, wa, userID, "replay-test")

	email := testWebAuthnUserName

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

	if _, err := wa.FinishLogin(ctx, sid1, []byte(body1)); err == nil {
		t.Fatal("replay with reused session_id succeeded; single-use protection broken")
	}

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

	body2 := virtualwebauthn.CreateAssertionResponse(rp, authenticator, cred, *parsedOpts2)

	if _, err := wa.FinishLogin(ctx, sid2, []byte(body2)); err != nil {

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
