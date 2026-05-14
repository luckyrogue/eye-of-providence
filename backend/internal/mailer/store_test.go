package mailer

import (
	"context"
	"strings"
	"testing"
)

func TestRender_BasicSubstitution(t *testing.T) {
	tpl := Template{
		Subject:  "Hi {{.Recipient}}",
		BodyHTML: "<p>Hello, {{.Recipient}}!</p>",
		BodyText: "Hello, {{.Recipient}}!",
	}
	out, err := Render(tpl, map[string]any{"Recipient": "Ada"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if out.Subject != "Hi Ada" {
		t.Errorf("subject = %q", out.Subject)
	}
	if !strings.Contains(out.HTML, "Hello, Ada!") {
		t.Errorf("html missing substitution: %q", out.HTML)
	}
	if out.Text != "Hello, Ada!" {
		t.Errorf("text = %q", out.Text)
	}
}

// TestRender_HTMLAutoEscape — критический guard: html/template должен
// автоматически escape'ить user-controlled value в HTML body, даже если
// admin-template содержит `{{.team_name}}` без явного `htmlEscape`.
func TestRender_HTMLAutoEscape(t *testing.T) {
	tpl := Template{
		Subject:  "Invite to {{.TeamName}}",
		BodyHTML: "<p>Welcome to {{.TeamName}}.</p>",
		BodyText: "Welcome to {{.TeamName}}.",
	}
	malicious := `<script>alert(1)</script>`
	out, err := Render(tpl, map[string]any{"TeamName": malicious})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(out.HTML, "<script>") {
		t.Errorf("html/template failed to escape <script>: %q", out.HTML)
	}
	if !strings.Contains(out.HTML, "&lt;script&gt;") {
		t.Errorf("html escape missing &lt;script&gt;: %q", out.HTML)
	}
	// text body — text/template НЕ escape'ит, raw passthrough.
	if !strings.Contains(out.Text, "<script>") {
		t.Errorf("text body should keep raw chars: %q", out.Text)
	}
}

// TestRender_SubjectStripsCRLF — anti header-injection.
func TestRender_SubjectStripsCRLF(t *testing.T) {
	tpl := Template{
		Subject:  "Subj\r\nBcc: attacker@evil.com\r\n",
		BodyHTML: "x",
		BodyText: "x",
	}
	out, err := Render(tpl, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.ContainsAny(out.Subject, "\r\n") {
		t.Errorf("subject still has CR/LF: %q", out.Subject)
	}
	if out.Subject != "Subj"+"Bcc: attacker@evil.com" {
		// Strip removes CR/LF only — content stays joined. Subject is fully
		// sanitized at storage level в admin endpoint; here we just ensure
		// the wire-level guard fires.
		t.Logf("subject after sanitize: %q (CR/LF stripped, content joined)", out.Subject)
	}
}

// TestRender_MissingVariable_ZeroValue — Option("missingkey=zero") ⇒
// неизвестная переменная рендерится как пустая строка, не error.
// Защищает send pipeline от падения если admin забыл какую-то переменную.
func TestRender_MissingVariable_ZeroValue(t *testing.T) {
	tpl := Template{
		Subject:  "X",
		BodyHTML: "<p>{{.Missing}}</p>",
		BodyText: "{{.Missing}}",
	}
	out, err := Render(tpl, map[string]any{})
	if err != nil {
		t.Fatalf("missingkey=zero should not error: %v", err)
	}
	if strings.Contains(out.HTML, "Missing") {
		t.Errorf("html should not contain raw 'Missing': %q", out.HTML)
	}
}

// TestRender_ParseError_BadSyntax — sanity: некорректный template syntax
// должен вернуть error, а не render заглушку.
func TestRender_ParseError_BadSyntax(t *testing.T) {
	tpl := Template{
		Subject:  "X",
		BodyHTML: "<p>{{ unclosed",
		BodyText: "x",
	}
	_, err := Render(tpl, nil)
	if err == nil {
		t.Fatal("expected parse error for unclosed template")
	}
}

func TestNilStore_Lookup(t *testing.T) {
	tpl, err := NilStore{}.Lookup(context.Background(), TemplateKeyTeamInvite, LocaleRU)
	if err != nil {
		t.Errorf("nil store err = %v, want nil", err)
	}
	if tpl != nil {
		t.Errorf("nil store template = %+v, want nil", tpl)
	}
}

func TestIsSupportedTemplateKey(t *testing.T) {
	if !IsSupportedTemplateKey(TemplateKeyTeamInvite) {
		t.Error("team_invite should be supported")
	}
	if IsSupportedTemplateKey("random_key") {
		t.Error("random key should not be supported")
	}
	if IsSupportedTemplateKey("") {
		t.Error("empty key should not be supported")
	}
}

func TestIsSupportedLocale(t *testing.T) {
	if !IsSupportedLocale("ru") {
		t.Error("ru should be supported")
	}
	if IsSupportedLocale("de") {
		t.Error("de should not be supported")
	}
}
