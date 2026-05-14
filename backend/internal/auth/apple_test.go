package auth

import (
	"context"
	"net/url"
	"strings"
	"testing"
)

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

func TestApple_Exchange_IsStub(t *testing.T) {
	a := newAppleStub()
	ext, err := a.Exchange(context.Background(), "fake-code")
	if err == nil {
		t.Fatal("expected stub error, got nil — Apple Exchange must NOT mint identities until Phase 2 complete")
	}
	if ext != nil {
		t.Errorf("ExternalUser should be nil for stub; got %+v", ext)
	}

	msg := err.Error()
	if !strings.Contains(msg, "stub") && !strings.Contains(msg, "not implemented") {
		t.Errorf("error message should mention 'stub' or 'not implemented'; got %q", msg)
	}
}

func TestApple_ImplementsOAuthProvider(t *testing.T) {
	var p OAuthProvider = newAppleStub()
	if p.Name() != "apple" {
		t.Errorf("via interface: Name() = %q", p.Name())
	}
}
