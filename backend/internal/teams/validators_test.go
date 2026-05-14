package teams

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	cases := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{"user@example.com", "user@example.com", true},
		{"  user@example.com  ", "user@example.com", true},
		{"User@Example.COM", "user@example.com", true},
		{"first.last+tag@sub.example.com", "first.last+tag@sub.example.com", true},
		{"", "", false},
		{"not-an-email", "", false},
		{"@example.com", "", false},
		{"user@", "", false},
		{strings.Repeat("a", 250) + "@x.com", "", false},
	}
	for _, tc := range cases {
		got, ok := validateEmail(tc.in)
		if ok != tc.wantOK {
			t.Errorf("validateEmail(%q) ok = %v, want %v", tc.in, ok, tc.wantOK)
			continue
		}
		if ok && got != tc.want {
			t.Errorf("validateEmail(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"short", false},
		{"1234567", false},
		{"12345678", true},
		{strings.Repeat("a", 256), true},
		{strings.Repeat("a", 257), false},
	}
	for _, tc := range cases {
		if got := validatePassword(tc.in); got != tc.want {
			t.Errorf("validatePassword(len=%d) = %v, want %v", len(tc.in), got, tc.want)
		}
	}
}

func TestValidateDisplayName(t *testing.T) {
	cases := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{"Alice", "Alice", true},
		{"  Alice  ", "Alice", true},
		{"", "", false},
		{"   ", "", false},
		{"with\nnewline", "", false},
		{"with\ttab", "", false},
		{"with\rCR", "", false},
		{strings.Repeat("a", 64), strings.Repeat("a", 64), true},
		{strings.Repeat("a", 65), "", false},
		{"Юникод OK", "Юникод OK", true},
	}
	for _, tc := range cases {
		got, ok := validateDisplayName(tc.in)
		if ok != tc.wantOK {
			t.Errorf("validateDisplayName(%q) ok = %v, want %v", tc.in, ok, tc.wantOK)
			continue
		}
		if ok && got != tc.want {
			t.Errorf("validateDisplayName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
