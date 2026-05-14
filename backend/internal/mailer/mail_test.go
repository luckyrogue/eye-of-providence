package mailer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestResendSend_Success(t *testing.T) {
	var captured resendReq
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("auth header = %q, want Bearer test-key", got)
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &captured); err != nil {
			t.Fatalf("body decode: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"em_123"}`))
	}))
	defer srv.Close()

	prev := resendURL
	resendURL = srv.URL
	defer func() { resendURL = prev }()

	m := NewResend("test-key", "Test <test@x.dev>", zap.NewNop())
	if err := m.Send(context.Background(), "to@x.dev", "Hello", "<b>Hi</b>", "Hi"); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if captured.From != "Test <test@x.dev>" {
		t.Errorf("from = %q", captured.From)
	}
	if len(captured.To) != 1 || captured.To[0] != "to@x.dev" {
		t.Errorf("to = %v", captured.To)
	}
	if captured.Subject != "Hello" {
		t.Errorf("subject = %q", captured.Subject)
	}
	if captured.HTML != "<b>Hi</b>" {
		t.Errorf("html = %q", captured.HTML)
	}
}

func TestResendSend_4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(422)
		_, _ = w.Write([]byte(`{"name":"validation_error","message":"domain not verified"}`))
	}))
	defer srv.Close()
	prev := resendURL
	resendURL = srv.URL
	defer func() { resendURL = prev }()

	m := NewResend("k", "f@x.dev", zap.NewNop())
	err := m.Send(context.Background(), "to@x.dev", "S", "", "T")
	if err == nil {
		t.Fatal("expected error on 422")
	}
	if !strings.Contains(err.Error(), "domain not verified") {
		t.Errorf("error message lost: %v", err)
	}
}

func TestNoopMailer(t *testing.T) {
	m := Noop(zap.NewNop())
	if err := m.Send(context.Background(), "to@x.dev", "S", "h", "t"); err != nil {
		t.Errorf("Noop.Send returned error: %v", err)
	}
}

func TestInviteEmailTemplate(t *testing.T) {
	subject, html, text := InviteEmail("Acme", "https://app.dev/?invite=ABC", "Темирлан", LocaleRU)
	if !strings.Contains(subject, "Acme") {
		t.Errorf("subject missing team name: %q", subject)
	}
	if !strings.Contains(html, "https://app.dev/?invite=ABC") {
		t.Errorf("html missing invite URL")
	}
	if !strings.Contains(text, "Acme") {
		t.Errorf("text missing team name")
	}
	if !strings.Contains(html, "Темирлан") {
		t.Errorf("html missing inviter name")
	}
}

func TestHTMLEscape(t *testing.T) {
	cases := map[string]string{
		"plain":    "plain",
		"<script>": "&lt;script&gt;",
		`a"b`:      `a&quot;b`,
		"a&b":      "a&amp;b",
		"a'b":      "a&#39;b",
	}
	for in, want := range cases {
		if got := htmlEscape(in); got != want {
			t.Errorf("htmlEscape(%q) = %q, want %q", in, got, want)
		}
	}
}
