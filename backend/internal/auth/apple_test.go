package auth

import (
	"context"
	"net/url"
	"strings"
	"testing"
)

// Apple Sign-in — Phase 1 stub.
//
// Per `.team/product-decisions-confirmed.md` §1: Apple deferred to v0.6 because
// the founder is not enrolled in Apple Developer Program. The stub must:
//
//   1. Satisfy the OAuthProvider interface so `cmd/api/main.go` can wire it conditionally.
//   2. Form a syntactically correct AuthCodeURL (so we don't ship a half-working button).
//   3. Refuse to perform Exchange — return an explicit "stub" error so handler.go can
//      return 502/503 instead of falling through and minting a JWT for a fake user.
//
// These tests are regression guards — they fail loudly if anyone "completes" Exchange
// without finishing the rest of the flow (JWKS verification, client_secret JWT signing,
// email_verified guard).

func newAppleStub() *AppleOAuth {
	return NewAppleOAuth(
		"team-id-XYZ",
		"key-id-ABC",
		"com.eop.web",
		"-----BEGIN PRIVATE KEY-----\nDUMMY\n-----END PRIVATE KEY-----\n",
		"https://eop.example/v1/auth/apple/callback",
	)
}

func TestApple_NameIsApple(t *testing.T) {
	a := newAppleStub()
	if got := a.Name(); got != "apple" {
		t.Fatalf("Name() = %q, want %q", got, "apple")
	}
}

func TestApple_AuthCodeURL_IncludesRequiredParams(t *testing.T) {
	a := newAppleStub()
	raw := a.AuthCodeURL("xyz-state-token")

	if !strings.HasPrefix(raw, "https://appleid.apple.com/auth/authorize?") {
		t.Fatalf("unexpected URL host: %s", raw)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	q := u.Query()
	// Apple-specific params (form_post + scope=name email + response_type=code).
	mustEq := map[string]string{
		"client_id":     "com.eop.web",
		"redirect_uri":  "https://eop.example/v1/auth/apple/callback",
		"response_type": "code",
		"scope":         "name email",
		"response_mode": "form_post",
		"state":         "xyz-state-token",
	}
	for k, want := range mustEq {
		if got := q.Get(k); got != want {
			t.Errorf("query %q = %q, want %q", k, got, want)
		}
	}
}

// TestApple_Exchange_IsStub — критическая регрессия. Если кто-то имплементит
// Exchange до того, как добавит JWKS verification + email_verified guard +
// signed client_secret — этот тест должен поломаться, чтобы PR не прошёл.
//
// Когда Apple flow будет ready, удалить этот тест и заменить на полноценную
// suite (mock JWKS + sign id_token тестовым ключом).
func TestApple_Exchange_IsStub(t *testing.T) {
	a := newAppleStub()
	ext, err := a.Exchange(context.Background(), "fake-code")
	if err == nil {
		t.Fatal("expected stub error, got nil — Apple Exchange must NOT mint identities until Phase 2 complete")
	}
	if ext != nil {
		t.Errorf("ExternalUser should be nil for stub; got %+v", ext)
	}
	// Сообщение должно явно намекать на stub, чтобы handler/log могли отличить
	// "real auth error" от "feature not enabled".
	msg := err.Error()
	if !strings.Contains(msg, "stub") && !strings.Contains(msg, "not implemented") {
		t.Errorf("error message should mention 'stub' or 'not implemented'; got %q", msg)
	}
}

// TestApple_ImplementsOAuthProvider — compile-time guard, дополнительно к
// `var _ OAuthProvider = (*AppleOAuth)(nil)` в apple.go.
func TestApple_ImplementsOAuthProvider(t *testing.T) {
	var p OAuthProvider = newAppleStub()
	if p.Name() != "apple" {
		t.Errorf("via interface: Name() = %q", p.Name())
	}
}
