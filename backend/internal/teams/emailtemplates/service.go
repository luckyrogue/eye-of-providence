package emailtemplates

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const maxBodyBytes = 256 * 1024

type TemplateSyntaxError struct {
	Detail string
}

func (e *TemplateSyntaxError) Error() string { return e.Detail }

type Service struct {
	repo      OverrideRepository
	baseline  BaselineProvider
	validator TemplateSyntaxValidator
	audit     AuditSink
	keys      []string
	locales   []string
}

type Deps struct {
	Repo      OverrideRepository
	Baseline  BaselineProvider
	Validator TemplateSyntaxValidator
	Audit     AuditSink
	Keys      []string
	Locales   []string
}

func New(d Deps) *Service {
	return &Service{
		repo:      d.Repo,
		baseline:  d.Baseline,
		validator: d.Validator,
		audit:     d.Audit,
		keys:      append([]string(nil), d.Keys...),
		locales:   append([]string(nil), d.Locales...),
	}
}

func (s *Service) validateKeyLocale(key, locale string) error {
	if !contains(s.keys, key) {
		return ErrInvalidKey
	}
	if !contains(s.locales, locale) {
		return ErrInvalidLocale
	}
	return nil
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func (s *Service) ListMatrix(ctx context.Context) ([]MatrixEntry, error) {
	overrides := map[string]OverrideRow{}
	if s.repo != nil {
		rows, err := s.repo.ListOverrides(ctx)
		if err != nil {
			return nil, err
		}
		for _, t := range rows {
			overrides[t.Key+":"+t.Locale] = t
		}
	}
	out := make([]MatrixEntry, 0, len(s.keys)*len(s.locales))
	for _, key := range s.keys {
		for _, loc := range s.locales {
			entry := MatrixEntry{Key: key, Locale: loc}
			if row, ok := overrides[key+":"+loc]; ok {
				entry.HasOverride = true
				ts := row.UpdatedAt
				entry.UpdatedAt = &ts
				entry.UpdatedBy = row.UpdatedBy
			}
			out = append(out, entry)
		}
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, key, locale string) (*View, error) {
	if err := s.validateKeyLocale(key, locale); err != nil {
		return nil, err
	}
	v := &View{Key: key, Locale: locale}
	if s.repo != nil {
		row, err := s.repo.Lookup(ctx, key, locale)
		if err != nil {
			return nil, err
		}
		if row != nil {
			v.Subject = row.Subject
			v.BodyHTML = row.BodyHTML
			v.BodyText = row.BodyText
			v.IsDefault = false
			ts := row.UpdatedAt
			v.UpdatedAt = &ts
			v.UpdatedBy = row.UpdatedBy
			return v, nil
		}
	}
	if s.baseline == nil {
		return nil, fmt.Errorf("%w: no baseline for %s:%s", ErrNoBaseline, key, locale)
	}
	base := s.baseline.Template(key, locale)
	if base == nil {
		return nil, fmt.Errorf("%w: no baseline for %s:%s", ErrNoBaseline, key, locale)
	}
	v.Subject = base.Subject
	v.BodyHTML = base.BodyHTML
	v.BodyText = base.BodyText
	v.IsDefault = true
	return v, nil
}

func (s *Service) Upsert(ctx context.Context, meta RequestMeta, actorID uuid.UUID, actorEmail, key, locale string, cmd UpsertCommand) (*UpsertResult, error) {
	if s.repo == nil {
		return nil, ErrStoreUnavailable
	}
	if err := s.validateKeyLocale(key, locale); err != nil {
		return nil, err
	}
	if strings.TrimSpace(cmd.Subject) == "" || strings.TrimSpace(cmd.BodyHTML) == "" {
		return nil, ErrMissingField
	}
	if len(cmd.BodyHTML) > maxBodyBytes || len(cmd.BodyText) > maxBodyBytes {
		return nil, ErrBodyTooLarge
	}
	if s.validator != nil {
		if err := s.validator.Validate(cmd.Subject, cmd.BodyHTML, cmd.BodyText, key, locale); err != nil {
			s.logAudit(ctx, meta, actorID, actorEmail, "email_template.update_rejected", "email_template", key+":"+locale, map[string]any{
				"error_code":   "invalid_template_syntax",
				"error_detail": err.Error(),
			})
			return nil, &TemplateSyntaxError{Detail: err.Error()}
		}
	}
	row := OverrideRow{
		Key:      key,
		Locale:   locale,
		Subject:  cmd.Subject,
		BodyHTML: cmd.BodyHTML,
		BodyText: cmd.BodyText,
	}
	out, err := s.repo.Upsert(ctx, row, actorID)
	if err != nil {
		return nil, err
	}
	s.logAudit(ctx, meta, actorID, actorEmail, "email_template.updated", "email_template", key+":"+locale, map[string]any{
		"subject":         cmd.Subject,
		"body_html_bytes": len(cmd.BodyHTML),
		"body_text_bytes": len(cmd.BodyText),
	})
	return &UpsertResult{
		Key:       out.Key,
		Locale:    out.Locale,
		Subject:   out.Subject,
		BodyHTML:  out.BodyHTML,
		BodyText:  out.BodyText,
		UpdatedAt: out.UpdatedAt,
		UpdatedBy: out.UpdatedBy,
		IsDefault: false,
	}, nil
}

func (s *Service) Delete(ctx context.Context, meta RequestMeta, actorID uuid.UUID, actorEmail, key, locale string) (*OverrideRow, error) {
	if s.repo == nil {
		return nil, ErrStoreUnavailable
	}
	if err := s.validateKeyLocale(key, locale); err != nil {
		return nil, err
	}
	prev, err := s.repo.Delete(ctx, key, locale)
	if err != nil {
		return nil, err
	}
	if prev == nil {
		return nil, fmt.Errorf("%w: no override exists for %s:%s", ErrNoOverride, key, locale)
	}
	s.logAudit(ctx, meta, actorID, actorEmail, "email_template.reverted", "email_template", key+":"+locale, map[string]any{
		"previous_subject":    prev.Subject,
		"previous_body_html":  prev.BodyHTML,
		"previous_body_text":  prev.BodyText,
		"previous_updated_at": prev.UpdatedAt,
		"previous_updated_by": prev.UpdatedBy,
	})
	return prev, nil
}

func (s *Service) logAudit(ctx context.Context, meta RequestMeta, actorID uuid.UUID, actorEmail, action, targetType, targetID string, md map[string]any) {
	if s.audit == nil {
		return
	}
	s.audit.Log(ctx, AuditEvent{
		ActorID:    actorID,
		ActorEmail: actorEmail,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Metadata:   md,
		IP:         meta.IP,
		UserAgent:  meta.UserAgent,
	})
}

func JoinSupported(items []string) string {
	var b strings.Builder
	for i, x := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(x)
	}
	return b.String()
}

func IsInvalidSyntax(err error) (detail string, ok bool) {
	var syn *TemplateSyntaxError
	if errors.As(err, &syn) {
		return syn.Detail, true
	}
	return "", false
}
