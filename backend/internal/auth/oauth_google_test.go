package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

func newGoogleWithMockEndpoint(t *testing.T, tokenURL string) *GoogleOAuth {
	t.Helper()
	g := NewGoogleOAuth("test-client-id", "test-client-secret", "https://eop.example/cb")
	g.cfg.Endpoint = oauth2.Endpoint{
		AuthURL:  "https://example/auth",
		TokenURL: tokenURL,
	}
	return g
}

func TestGoogle_NameIsGoogle(t *testing.T) {
	g := NewGoogleOAuth("c", "s", "https://x.example/cb")
	if got := g.Name(); got != "google" {
		t.Fatalf("Name() = %q, want google", got)
	}
}

func TestGoogle_ImplementsOAuthProvider(t *testing.T) {
	var p OAuthProvider = NewGoogleOAuth("c", "s", "https://x.example/cb")
	if p.Name() != "google" {
		t.Errorf("Name via interface = %q", p.Name())
	}
}

func TestGoogle_AuthCodeURL_IncludesState(t *testing.T) {
	g := NewGoogleOAuth("test-client-id", "test-secret", "https://eop.example/cb")
	u, err := url.Parse(g.AuthCodeURL("xyz-state"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if u.Query().Get("state") != "xyz-state" {
		t.Errorf("state = %q", u.Query().Get("state"))
	}
	if u.Query().Get("client_id") != "test-client-id" {
		t.Errorf("client_id = %q", u.Query().Get("client_id"))
	}
	if u.Query().Get("redirect_uri") != "https://eop.example/cb" {
		t.Errorf("redirect_uri = %q", u.Query().Get("redirect_uri"))
	}
	if !strings.Contains(u.Query().Get("scope"), "email") {
		t.Errorf("scope missing email: %q", u.Query().Get("scope"))
	}
	if !strings.Contains(u.Query().Get("scope"), "openid") {
		t.Errorf("scope missing openid: %q", u.Query().Get("scope"))
	}
}

func TestGoogle_Exchange_NotConfigured(t *testing.T) {
	g := NewGoogleOAuth("", "secret", "https://eop.example/cb")
	_, err := g.Exchange(context.Background(), "any-code")
	if err == nil {
		t.Fatal("expected error for empty client id, got nil")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("error should mention 'not configured'; got %q", err.Error())
	}
}

func TestGoogle_Exchange_TokenEndpoint5xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal", http.StatusInternalServerError)
	}))
	defer srv.Close()

	g := newGoogleWithMockEndpoint(t, srv.URL+"/token")
	_, err := g.Exchange(context.Background(), "any-code")
	if err == nil {
		t.Fatal("expected error on 5xx token endpoint, got nil")
	}

	if !strings.Contains(strings.ToLower(err.Error()), "token") &&
		!strings.Contains(strings.ToLower(err.Error()), "exchange") {
		t.Errorf("error should reference token exchange; got %q", err.Error())
	}
}

func TestGoogle_Exchange_MissingIDToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "ya29.fake",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer srv.Close()

	g := newGoogleWithMockEndpoint(t, srv.URL+"/token")
	_, err := g.Exchange(context.Background(), "code")
	if err == nil {
		t.Fatal("expected error on missing id_token, got nil")
	}
	if !strings.Contains(err.Error(), "id_token") {
		t.Errorf("error should mention id_token; got %q", err.Error())
	}
}

func TestGoogle_Exchange_BadIDTokenSignature(t *testing.T) {
	if testing.Short() {
		t.Skip("requires Google JWKS fetch")
	}

	hdr, err := json.Marshal(map[string]string{"alg": "none", "typ": "JWT"})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{
		"sub":            "123",
		"email":          "a@b.com",
		"email_verified": true,
		"aud":            "test-client-id",
	})
	if err != nil {
		t.Fatal(err)
	}
	badIDToken := base64.RawURLEncoding.EncodeToString(hdr) + "." +
		base64.RawURLEncoding.EncodeToString(payload) + "."

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "fake",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"id_token":     badIDToken,
		})
	}))
	defer srv.Close()

	g := newGoogleWithMockEndpoint(t, srv.URL+"/token")
	_, err = g.Exchange(context.Background(), "code")
	if err == nil {
		t.Fatal("expected validate error on alg=none id_token, got nil")
	}
}

func newGoogleWithMockTokenAndValidator(t *testing.T, tokenURL string, validator GoogleIDTokenValidator) *GoogleOAuth {
	t.Helper()
	g := NewGoogleOAuth("test-client-id", "test-client-secret", "https://eop.example/cb", WithGoogleValidator(validator))
	g.cfg.Endpoint = oauth2.Endpoint{
		AuthURL:  "https://example/auth",
		TokenURL: tokenURL,
	}
	return g
}

func mockTokenServer(t *testing.T, idToken string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "ya29.fake",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"id_token":     idToken,
		})
	}))
}

func TestGoogle_Exchange_HappyPath(t *testing.T) {
	srv := mockTokenServer(t, "fake-id-token")
	defer srv.Close()

	validator := func(_ context.Context, raw, aud string) (*idtoken.Payload, error) {
		if raw != "fake-id-token" {
			t.Errorf("validator got raw=%q", raw)
		}
		if aud != "test-client-id" {
			t.Errorf("validator got aud=%q", aud)
		}
		return &idtoken.Payload{
			Subject: "108765432109876543210",
			Claims: map[string]any{
				"email":          "alice@example.com",
				"email_verified": true,
				"name":           "Alice Example",
			},
		}, nil
	}

	g := newGoogleWithMockTokenAndValidator(t, srv.URL+"/token", validator)
	user, err := g.Exchange(context.Background(), "code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Subject != "108765432109876543210" {
		t.Errorf("Subject = %q, want google sub", user.Subject)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("Email = %q", user.Email)
	}
	if user.Name != "Alice Example" {
		t.Errorf("Name = %q", user.Name)
	}
	if user.Login != "" {
		t.Errorf("Login should be empty for google, got %q", user.Login)
	}
}

func TestGoogle_Exchange_EmailNotVerified(t *testing.T) {
	srv := mockTokenServer(t, "fake-id-token")
	defer srv.Close()

	validator := func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
		return &idtoken.Payload{
			Subject: "108765432109876543210",
			Claims: map[string]any{
				"email":          "alice@example.com",
				"email_verified": false,
				"name":           "Alice Example",
			},
		}, nil
	}

	g := newGoogleWithMockTokenAndValidator(t, srv.URL+"/token", validator)
	_, err := g.Exchange(context.Background(), "code")
	if err == nil {
		t.Fatal("expected rejection when email_verified=false")
	}
	if !strings.Contains(err.Error(), "not verified") {
		t.Errorf("error should mention 'not verified'; got %q", err.Error())
	}
}

func TestGoogle_Exchange_AudMismatch(t *testing.T) {
	srv := mockTokenServer(t, "fake-id-token")
	defer srv.Close()

	validator := func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
		return nil, errors.New("idtoken: audience provided does not match aud claim in the JWT")
	}

	g := newGoogleWithMockTokenAndValidator(t, srv.URL+"/token", validator)
	_, err := g.Exchange(context.Background(), "code")
	if err == nil {
		t.Fatal("expected aud-mismatch error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "validate") &&
		!strings.Contains(strings.ToLower(err.Error()), "aud") {
		t.Errorf("error should reference validate/aud; got %q", err.Error())
	}
}

func TestGoogle_Exchange_MissingEmail(t *testing.T) {
	srv := mockTokenServer(t, "fake-id-token")
	defer srv.Close()

	validator := func(_ context.Context, _, _ string) (*idtoken.Payload, error) {
		return &idtoken.Payload{
			Subject: "12345",
			Claims: map[string]any{
				"email_verified": true,
			},
		}, nil
	}

	g := newGoogleWithMockTokenAndValidator(t, srv.URL+"/token", validator)
	_, err := g.Exchange(context.Background(), "code")
	if err == nil {
		t.Fatal("expected error when email claim missing")
	}
	if !strings.Contains(err.Error(), "missing email") {
		t.Errorf("error should mention 'missing email'; got %q", err.Error())
	}
}
