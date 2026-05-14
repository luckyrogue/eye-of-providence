package teams

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/mailer"
	"github.com/eye-of-providence/backend/internal/teams/emailtemplates"
)

type pgOverrideRepo struct {
	store *mailer.PGTemplateStore
}

func (r pgOverrideRepo) ListOverrides(ctx context.Context) ([]emailtemplates.OverrideRow, error) {
	if r.store == nil || r.store.Pool == nil {
		return nil, nil
	}
	rows, err := r.store.ListOverrides(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]emailtemplates.OverrideRow, 0, len(rows))
	for _, t := range rows {
		out = append(out, mailerTemplateToRow(t))
	}
	return out, nil
}

func (r pgOverrideRepo) Lookup(ctx context.Context, key, locale string) (*emailtemplates.OverrideRow, error) {
	if r.store == nil || r.store.Pool == nil {
		return nil, nil
	}
	t, err := r.store.Lookup(ctx, key, mailer.Locale(locale))
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, nil
	}
	row := mailerTemplateToRow(*t)
	return &row, nil
}

func (r pgOverrideRepo) Upsert(ctx context.Context, row emailtemplates.OverrideRow, actorID uuid.UUID) (*emailtemplates.OverrideRow, error) {
	if r.store == nil || r.store.Pool == nil {
		return nil, emailtemplates.ErrStoreUnavailable
	}
	mt := mailer.Template{
		Key:      row.Key,
		Locale:   row.Locale,
		Subject:  row.Subject,
		BodyHTML: row.BodyHTML,
		BodyText: row.BodyText,
	}
	out, err := r.store.Upsert(ctx, mt, actorID)
	if err != nil {
		switch {
		case errors.Is(err, mailer.ErrInvalidEmailTemplateKey):
			return nil, emailtemplates.ErrInvalidKey
		case errors.Is(err, mailer.ErrInvalidEmailTemplateLocale):
			return nil, emailtemplates.ErrInvalidLocale
		default:
			return nil, err
		}
	}
	res := mailerTemplateToRow(*out)
	return &res, nil
}

func (r pgOverrideRepo) Delete(ctx context.Context, key, locale string) (*emailtemplates.OverrideRow, error) {
	if r.store == nil || r.store.Pool == nil {
		return nil, emailtemplates.ErrStoreUnavailable
	}
	prev, err := r.store.Delete(ctx, key, locale)
	if err != nil {
		return nil, err
	}
	if prev == nil {
		return nil, nil
	}
	row := mailerTemplateToRow(*prev)
	return &row, nil
}

func mailerTemplateToRow(t mailer.Template) emailtemplates.OverrideRow {
	return emailtemplates.OverrideRow{
		Key:       t.Key,
		Locale:    t.Locale,
		Subject:   t.Subject,
		BodyHTML:  t.BodyHTML,
		BodyText:  t.BodyText,
		UpdatedAt: t.UpdatedAt,
		UpdatedBy: t.UpdatedBy,
	}
}

type baselineAdapter struct{}

func (baselineAdapter) Template(key, locale string) *emailtemplates.OverrideRow {
	b := mailer.BaselineTemplate(key, mailer.Locale(locale))
	if b == nil {
		return nil
	}
	return &emailtemplates.OverrideRow{
		Key:       b.Key,
		Locale:    b.Locale,
		Subject:   b.Subject,
		BodyHTML:  b.BodyHTML,
		BodyText:  b.BodyText,
		UpdatedAt: b.UpdatedAt,
		UpdatedBy: b.UpdatedBy,
	}
}

type renderValidator struct{}

func (renderValidator) Validate(subject, bodyHTML, bodyText, key, locale string) error {
	_, err := mailer.Render(mailer.Template{
		Key:      key,
		Locale:   locale,
		Subject:  subject,
		BodyHTML: bodyHTML,
		BodyText: bodyText,
	}, nil)
	return err
}

type auditSinkAdapter struct {
	svc audit.Service
}

func (a auditSinkAdapter) Log(ctx context.Context, e emailtemplates.AuditEvent) {
	if a.svc.Pool == nil {
		return
	}
	a.svc.Log(ctx, audit.Entry{
		ActorID:    e.ActorID,
		ActorEmail: e.ActorEmail,
		Action:     audit.Action(e.Action),
		TargetType: e.TargetType,
		TargetID:   e.TargetID,
		Metadata:   e.Metadata,
		IP:         e.IP,
		UserAgent:  e.UserAgent,
	})
}

func (s Service) newEmailTemplatesService() *emailtemplates.Service {
	return emailtemplates.New(emailtemplates.Deps{
		Repo:      pgOverrideRepo{store: s.TemplateStore},
		Baseline:  baselineAdapter{},
		Validator: renderValidator{},
		Audit:     auditSinkAdapter{svc: s.Audit},
		Keys:      mailer.SupportedTemplateKeys,
		Locales:   mailer.SupportedLocales,
	})
}
