package mailer

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.uber.org/zap"
)

type captureMailer struct {
	gotTo, gotSubject, gotHTML, gotText string
	failNext                            error
}

func (c *captureMailer) Send(_ context.Context, to, subject, html, text string) error {
	if c.failNext != nil {
		err := c.failNext
		c.failNext = nil
		return err
	}
	c.gotTo, c.gotSubject, c.gotHTML, c.gotText = to, subject, html, text
	return nil
}

type errStore struct{}

func (errStore) Lookup(_ context.Context, _ string, _ Locale) (*Template, error) {
	return nil, errors.New("db down")
}

type staticStore struct {
	tpl *Template
}

func (s staticStore) Lookup(_ context.Context, _ string, _ Locale) (*Template, error) {
	return s.tpl, nil
}

func TestDispatcher_FallbackToEmbedded_OnNoOverride(t *testing.T) {
	m := &captureMailer{}
	d := Dispatcher{Mailer: m, Store: NilStore{}, Logger: zap.NewNop()}

	err := d.SendTemplate(context.Background(), "user@ex.dev", TemplateKeyPasswordReset, LocaleEN,
		map[string]any{"ResetURL": "https://app/reset"})
	if err != nil {
		t.Fatalf("SendTemplate: %v", err)
	}
	if m.gotTo != "user@ex.dev" {
		t.Errorf("to = %q", m.gotTo)
	}
	if !strings.Contains(strings.ToLower(m.gotSubject), "password") &&
		!strings.Contains(strings.ToLower(m.gotSubject), "reset") {
		t.Errorf("subject lacks password/reset: %q", m.gotSubject)
	}
	if !strings.Contains(m.gotHTML, "https://app/reset") {
		t.Errorf("html missing URL: %q", m.gotHTML)
	}
}

func TestDispatcher_FallbackOnDBError(t *testing.T) {
	m := &captureMailer{}
	d := Dispatcher{Mailer: m, Store: errStore{}, Logger: zap.NewNop()}
	err := d.SendTemplate(context.Background(), "u@ex.dev", TemplateKeyPasswordReset, LocaleRU,
		map[string]any{"ResetURL": "https://app/r"})
	if err != nil {
		t.Fatalf("SendTemplate (DB outage path): %v", err)
	}
	if m.gotSubject == "" {
		t.Error("subject empty after fallback")
	}
}

func TestDispatcher_OverrideWins(t *testing.T) {
	override := &Template{
		Key:      TemplateKeyPasswordReset,
		Locale:   "en",
		Subject:  "Custom subject",
		BodyHTML: "<p>custom {{.ResetURL}}</p>",
		BodyText: "custom {{.ResetURL}}",
	}
	m := &captureMailer{}
	d := Dispatcher{Mailer: m, Store: staticStore{tpl: override}, Logger: zap.NewNop()}
	err := d.SendTemplate(context.Background(), "u@ex.dev", TemplateKeyPasswordReset, LocaleEN,
		map[string]any{"ResetURL": "https://app/r"})
	if err != nil {
		t.Fatalf("SendTemplate: %v", err)
	}
	if m.gotSubject != "Custom subject" {
		t.Errorf("subject = %q, want Custom subject", m.gotSubject)
	}
	if !strings.Contains(m.gotHTML, "https://app/r") {
		t.Errorf("html missing URL: %q", m.gotHTML)
	}
}

func TestDispatcher_RejectsUnknownKey(t *testing.T) {
	m := &captureMailer{}
	d := Dispatcher{Mailer: m, Store: NilStore{}, Logger: zap.NewNop()}
	err := d.SendTemplate(context.Background(), "u@ex.dev", "magic", LocaleEN, nil)
	if err == nil {
		t.Fatal("expected error for unknown template key")
	}
}

func TestDispatcher_NoMailer(t *testing.T) {
	d := Dispatcher{Mailer: nil, Store: NilStore{}, Logger: zap.NewNop()}
	err := d.SendTemplate(context.Background(), "u@ex.dev", TemplateKeyPasswordReset, LocaleEN, nil)
	if !errors.Is(err, ErrNoMailer) {
		t.Errorf("want ErrNoMailer, got %v", err)
	}
}
