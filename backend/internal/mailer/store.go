// store.go — DB-backed override layer для transactional email templates.
//
// Дизайн (Phase 3 admin):
//   * Embedded baseline в templates.go — source of truth для всех новых
//     инсталляций; код-deploys обновляют baseline без миграций.
//   * `email_templates` table (migration 022) — super_admin overrides per
//     (key, locale). Row exists ⇒ rendering use её subject+body; row missing
//     или DB unreachable ⇒ fallback на embedded baseline (см.
//     .team/product-acceptance/admin-email-templates.md Scenarios 2, 3).
//   * `TemplateStore` interface — позволяет тестам подменить хранилище без
//     запуска Postgres. `NilStore` всегда возвращает (nil, nil) ⇒ caller
//     знает, что нужно fallback.
//
// Render: `subject` через `text/template`, `body_html` через `html/template`
// (auto-escapes user input — критично для recipient-controlled vars типа
// {{team_name}} в team_invite, см. QA bug review). `body_text` тоже через
// `text/template`. Variable substitution синтаксис стандартный Go template:
// `{{.Name}}`, `{{.URL}}`. Mailer-layer caller передаёт map[string]any с
// уже подготовленными ключами.

package mailer

import (
	"bytes"
	"context"
	"errors"
	htmltpl "html/template"
	"strings"
	texttpl "text/template"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TemplateKey — известные имена шаблонов. Они же значения колонки `key` в
// email_templates table (migration 022). Список фиксирован — adding a new
// kind email требует backend changes (см. admin-email-templates.md scope).
const (
	TemplateKeyPasswordReset          = "password_reset"
	TemplateKeyTeamInvite             = "team_invite"
	TemplateKeySubscriptionActivated  = "subscription_activated"
)

// SupportedTemplateKeys — для валидации входа в admin endpoints. Hardcoded
// allowlist; модификация только через миграцию + backend code.
var SupportedTemplateKeys = []string{
	TemplateKeyPasswordReset,
	TemplateKeyTeamInvite,
	TemplateKeySubscriptionActivated,
}

// SupportedLocales — те же 4 локали, что в Locale enum. Returning string
// чтобы caller'у было удобно сравнивать без NormalizeLocale.
var SupportedLocales = []string{
	string(LocaleRU),
	string(LocaleEN),
	string(LocaleKK),
	string(LocaleES),
}

// IsSupportedTemplateKey — для admin handler validation. Returns false для
// произвольных строк типа "; DROP TABLE".
func IsSupportedTemplateKey(s string) bool {
	for _, k := range SupportedTemplateKeys {
		if s == k {
			return true
		}
	}
	return false
}

// IsSupportedLocale — same, для locale ввода.
func IsSupportedLocale(s string) bool {
	for _, l := range SupportedLocales {
		if s == l {
			return true
		}
	}
	return false
}

// Template — overrideROW, как хранится в DB и возвращается админ-UI.
type Template struct {
	Key       string
	Locale    string
	Subject   string
	BodyHTML  string
	BodyText  string
	UpdatedAt time.Time
	UpdatedBy *uuid.UUID
}

// TemplateStore — интерфейс для DB layer. Mailer вызывает Lookup на каждом
// Send (с graceful fallback) — реализация должна быть быстрой и tolerant к
// DB ошибкам (возвращать nil, nil лучше чем error для cache-miss path).
type TemplateStore interface {
	Lookup(ctx context.Context, key string, locale Locale) (*Template, error)
}

// NilStore — для тестов / dev mode без DB. Возвращает (nil, nil) на всё,
// заставляя caller'а fallback на embedded baseline.
type NilStore struct{}

func (NilStore) Lookup(_ context.Context, _ string, _ Locale) (*Template, error) {
	return nil, nil
}

// PGTemplateStore — Postgres implementation. Использует тот же pool, что и
// teams/auth — нет отдельного connection.
type PGTemplateStore struct {
	Pool *pgxpool.Pool
}

// NewPGTemplateStore — конструктор. Если Pool nil, методы no-op возвращают
// (nil, nil).
func NewPGTemplateStore(pool *pgxpool.Pool) *PGTemplateStore {
	return &PGTemplateStore{Pool: pool}
}

// Lookup — single row read. Не возвращает error на pgx.ErrNoRows — это
// expected (нет override ⇒ baseline). DB error возвращается, caller решает
// fallback.
func (s *PGTemplateStore) Lookup(ctx context.Context, key string, locale Locale) (*Template, error) {
	if s == nil || s.Pool == nil {
		return nil, nil
	}
	var t Template
	t.Key = key
	t.Locale = string(locale)
	err := s.Pool.QueryRow(ctx, `
		SELECT subject, body_html, body_text, updated_at, updated_by
		FROM email_templates
		WHERE key = $1 AND locale = $2`, key, string(locale),
	).Scan(&t.Subject, &t.BodyHTML, &t.BodyText, &t.UpdatedAt, &t.UpdatedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Upsert — INSERT ... ON CONFLICT для admin save endpoint.
func (s *PGTemplateStore) Upsert(ctx context.Context, t Template, actor uuid.UUID) (*Template, error) {
	if s == nil || s.Pool == nil {
		return nil, errors.New("template store: nil pool")
	}
	var actorPtr *uuid.UUID
	if actor != uuid.Nil {
		actorPtr = &actor
	}
	var out Template
	out.Key = t.Key
	out.Locale = t.Locale
	out.Subject = t.Subject
	out.BodyHTML = t.BodyHTML
	out.BodyText = t.BodyText
	err := s.Pool.QueryRow(ctx, `
		INSERT INTO email_templates (key, locale, subject, body_html, body_text, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, now(), $6)
		ON CONFLICT (key, locale) DO UPDATE SET
		  subject = EXCLUDED.subject,
		  body_html = EXCLUDED.body_html,
		  body_text = EXCLUDED.body_text,
		  updated_at = now(),
		  updated_by = EXCLUDED.updated_by
		RETURNING updated_at, updated_by`,
		t.Key, t.Locale, t.Subject, t.BodyHTML, t.BodyText, actorPtr,
	).Scan(&out.UpdatedAt, &out.UpdatedBy)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete — revert override ⇒ fallback to embedded baseline. Возвращает
// предыдущий row для audit log payload'а (per admin-email-templates.md
// Scenario 6).
func (s *PGTemplateStore) Delete(ctx context.Context, key, locale string) (*Template, error) {
	if s == nil || s.Pool == nil {
		return nil, errors.New("template store: nil pool")
	}
	var t Template
	t.Key = key
	t.Locale = locale
	err := s.Pool.QueryRow(ctx, `
		DELETE FROM email_templates
		WHERE key = $1 AND locale = $2
		RETURNING subject, body_html, body_text, updated_at, updated_by`,
		key, locale,
	).Scan(&t.Subject, &t.BodyHTML, &t.BodyText, &t.UpdatedAt, &t.UpdatedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ListOverrides — все existing rows. Используется admin matrix endpoint
// для определения `has_override` per (key, locale) пары.
func (s *PGTemplateStore) ListOverrides(ctx context.Context) ([]Template, error) {
	if s == nil || s.Pool == nil {
		return nil, nil
	}
	rows, err := s.Pool.Query(ctx, `
		SELECT key, locale, subject, body_html, body_text, updated_at, updated_by
		FROM email_templates
		ORDER BY key, locale`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Template{}
	for rows.Next() {
		var t Template
		if err := rows.Scan(&t.Key, &t.Locale, &t.Subject, &t.BodyHTML, &t.BodyText, &t.UpdatedAt, &t.UpdatedBy); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// --- Rendering ---

// RenderedTemplate — финальный output после substitution. Caller передаёт
// этот struct в mailer.Send().
type RenderedTemplate struct {
	Subject string
	HTML    string
	Text    string
}

// Render — substitution + auto-escape. `subject` и `body_text` через
// text/template (no escape, line-safe), `body_html` через html/template
// (auto-escapes user input). Если кто-то напишет `{{.team_name}}` со
// значением `<script>`, html/template вернёт `&lt;script&gt;` в HTML body
// автоматически — критично т.к. admin-edited templates могут содержать
// recipient-controlled vars типа team_name.
//
// Subject sanitized отдельно: CR/LF удаляется (anti header-injection).
func Render(t Template, vars map[string]any) (RenderedTemplate, error) {
	if vars == nil {
		vars = map[string]any{}
	}
	subject, err := renderText(t.Subject, vars)
	if err != nil {
		return RenderedTemplate{}, err
	}
	subject = sanitizeSubject(subject)

	html, err := renderHTML(t.BodyHTML, vars)
	if err != nil {
		return RenderedTemplate{}, err
	}
	text, err := renderText(t.BodyText, vars)
	if err != nil {
		return RenderedTemplate{}, err
	}
	return RenderedTemplate{
		Subject: subject,
		HTML:    html,
		Text:    text,
	}, nil
}

// renderText — text/template execution. Имя шаблона `tpl` фиксировано
// (Parse требует unique name; для one-shot render это не важно).
func renderText(src string, vars map[string]any) (string, error) {
	tpl, err := texttpl.New("tpl").Option("missingkey=zero").Parse(src)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, vars); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// renderHTML — html/template (auto-escape).
func renderHTML(src string, vars map[string]any) (string, error) {
	tpl, err := htmltpl.New("tpl").Option("missingkey=zero").Parse(src)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, vars); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// sanitizeSubject — strip CR/LF (anti header-injection). Subject поле в
// SMTP не должно содержать переводов строк; Resend API игнорирует, но
// мы strict.
func sanitizeSubject(s string) string {
	r := strings.NewReplacer("\r", "", "\n", "")
	return r.Replace(s)
}

// LookupOrFallback — convenience helper для caller'ов (mailer.SendTemplate).
// Если store == nil, возвращает nil без ошибки — caller use embedded
// baseline. На DB error возвращает (nil, err), но caller рекомендуется
// просто log + fallback.
func LookupOrFallback(ctx context.Context, store TemplateStore, key string, locale Locale) (*Template, error) {
	if store == nil {
		return nil, nil
	}
	return store.Lookup(ctx, key, locale)
}
