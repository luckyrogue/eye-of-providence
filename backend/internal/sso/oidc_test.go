package sso

import (
	"testing"
)

func TestOIDCProvider_CheckEmailDomain(t *testing.T) {
	tests := []struct {
		name           string
		allowedDomains []string
		email          string
		wantErr        bool
	}{
		{"no restriction", nil, "any@example.com", false},
		{"empty list", []string{}, "any@example.com", false},
		{"match", []string{"acme.com"}, "alice@acme.com", false},
		{"case insensitive", []string{"ACME.com"}, "alice@acme.com", false},
		{"no match", []string{"acme.com"}, "bob@evil.com", true},
		{"no @ — invalid", []string{"acme.com"}, "noatsign", true},
		{"multi-domain match second", []string{"acme.com", "good.io"}, "x@good.io", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &OIDCProvider{cfg: &Config{AllowedDomains: tt.allowedDomains}}
			err := p.CheckEmailDomain(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckEmailDomain(%q) error = %v, wantErr %v",
					tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestEmailLocalPart(t *testing.T) {
	tests := map[string]string{
		"alice@acme.com":      "alice",
		"a.b+c@example.org":   "a.b+c",
		"":                    "",
		"noatsign":            "noatsign",
		"@only":               "@only", // edge: at < 0 в LastIndex был бы 0
	}
	for in, want := range tests {
		got := emailLocalPart(in)
		if got != want {
			t.Errorf("emailLocalPart(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEscapePath(t *testing.T) {
	tests := map[string]string{
		"/dashboard":            "/dashboard",
		"/path#fragment":        "/path%23fragment",
		"/path&query":           "/path%26query",
		"/with spaces":          "/with%20spaces",
		"/clean?keep=ok":        "/clean?keep=ok", // ? not escaped
	}
	for in, want := range tests {
		got := escapePath(in)
		if got != want {
			t.Errorf("escapePath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeScopes(t *testing.T) {
	got := normalizeScopes([]string{" openid ", "email", "", "  ", "profile"})
	want := []string{"openid", "email", "profile"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("scope[%d] = %q, want %q", i, got[i], v)
		}
	}
}

func TestNormalizeDomains(t *testing.T) {
	got := normalizeDomains([]string{" ACME.com ", "  ", "good.IO"})
	want := []string{"acme.com", "good.io"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("domain[%d] = %q, want %q", i, got[i], v)
		}
	}
}
