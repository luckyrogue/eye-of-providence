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

const (
	TemplateKeyPasswordReset         = "password_reset"
	TemplateKeyTeamInvite            = "team_invite"
	TemplateKeySubscriptionActivated = "subscription_activated"
)

var SupportedTemplateKeys = []string{
	TemplateKeyPasswordReset,
	TemplateKeyTeamInvite,
	TemplateKeySubscriptionActivated,
}

var SupportedLocales = []string{
	string(LocaleRU),
	string(LocaleEN),
	string(LocaleKK),
	string(LocaleES),
}

func IsSupportedTemplateKey(s string) bool {
	for _, k := range SupportedTemplateKeys {
		if s == k {
			return true
		}
	}
	return false
}

func IsSupportedLocale(s string) bool {
	for _, l := range SupportedLocales {
		if s == l {
			return true
		}
	}
	return false
}

type Template struct {
	Key       string
	Locale    string
	Subject   string
	BodyHTML  string
	BodyText  string
	UpdatedAt time.Time
	UpdatedBy *uuid.UUID
}

type TemplateStore interface {
	Lookup(ctx context.Context, key string, locale Locale) (*Template, error)
}

var (
	ErrInvalidEmailTemplateKey    = errors.New("invalid email template key")
	ErrInvalidEmailTemplateLocale = errors.New("invalid email template locale")
)

type NilStore struct{}

func (NilStore) Lookup(_ context.Context, _ string, _ Locale) (*Template, error) {
	return nil, nil
}

type PGTemplateStore struct {
	Pool *pgxpool.Pool
}

func NewPGTemplateStore(pool *pgxpool.Pool) *PGTemplateStore {
	return &PGTemplateStore{Pool: pool}
}

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

func (s *PGTemplateStore) Upsert(ctx context.Context, t Template, actor uuid.UUID) (*Template, error) {
	if s == nil || s.Pool == nil {
		return nil, errors.New("template store: nil pool")
	}
	if !IsSupportedTemplateKey(t.Key) {
		return nil, ErrInvalidEmailTemplateKey
	}
	if !IsSupportedLocale(t.Locale) {
		return nil, ErrInvalidEmailTemplateLocale
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

type RenderedTemplate struct {
	Subject string
	HTML    string
	Text    string
}

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

func sanitizeSubject(s string) string {
	r := strings.NewReplacer("\r", "", "\n", "")
	return r.Replace(s)
}

func LookupOrFallback(ctx context.Context, store TemplateStore, key string, locale Locale) (*Template, error) {
	if store == nil {
		return nil, nil
	}
	return store.Lookup(ctx, key, locale)
}
