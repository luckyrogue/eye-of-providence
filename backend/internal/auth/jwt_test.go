package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

const testSecret = "test-secret-32-chars-or-longer-aaaa"

func TestIssueAndParse_HappyPath(t *testing.T) {
	tok, err := IssueJWT(testSecret, "user-1", "user@example.com", "password", 7, time.Hour)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	c, err := ParseJWT(testSecret, tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if c.UserID != "user-1" {
		t.Errorf("UserID = %q, want user-1", c.UserID)
	}
	if c.Email != "user@example.com" {
		t.Errorf("Email = %q", c.Email)
	}
	if c.Provider != "password" {
		t.Errorf("Provider = %q", c.Provider)
	}
	if c.TokenVersion != 7 {
		t.Errorf("TokenVersion = %d, want 7", c.TokenVersion)
	}
}

func TestParseJWT_WrongSecret(t *testing.T) {
	tok, _ := IssueJWT(testSecret, "user-1", "", "", 0, time.Hour)
	if _, err := ParseJWT("different-secret-32-chars-aaaaaaa", tok); err == nil {
		t.Fatal("expected error with wrong secret, got nil")
	}
}

func TestParseJWT_Expired(t *testing.T) {
	tok, _ := IssueJWT(testSecret, "user-1", "", "", 0, -time.Hour)
	if _, err := ParseJWT(testSecret, tok); err == nil {
		t.Fatal("expected error on expired token, got nil")
	}
}

func TestParseJWT_Garbage(t *testing.T) {
	if _, err := ParseJWT(testSecret, "not.a.jwt"); err == nil {
		t.Fatal("expected error on garbage, got nil")
	}
}

func TestParseJWT_RejectsAlgNone(t *testing.T) {
	hdr, err := json.Marshal(map[string]string{"alg": "none", "typ": "JWT"})
	if err != nil {
		t.Fatal(err)
	}
	pay, err := json.Marshal(map[string]string{"sub": "hack"})
	if err != nil {
		t.Fatal(err)
	}
	noneToken := base64.RawURLEncoding.EncodeToString(hdr) + "." +
		base64.RawURLEncoding.EncodeToString(pay) + "."
	if _, err := ParseJWT(testSecret, noneToken); err == nil {
		t.Fatal("expected reject of alg=none token, got nil")
	}
}

func TestIssueJWT_TokenVersionPropagated(t *testing.T) {
	for _, tv := range []int{0, 1, 42, 9999} {
		tok, err := IssueJWT(testSecret, "u", "", "", tv, time.Hour)
		if err != nil {
			t.Fatalf("tv=%d issue failed: %v", tv, err)
		}
		c, err := ParseJWT(testSecret, tok)
		if err != nil {
			t.Fatalf("tv=%d parse failed: %v", tv, err)
		}
		if c.TokenVersion != tv {
			t.Errorf("tv=%d round-trip got %d", tv, c.TokenVersion)
		}
	}
}

func TestIssueJWT_DotSeparated(t *testing.T) {

	tok, _ := IssueJWT(testSecret, "u", "", "", 0, time.Hour)
	if strings.Count(tok, ".") != 2 {
		t.Errorf("expected 2 dots, got %d in %q", strings.Count(tok, "."), tok)
	}
}
