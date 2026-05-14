package emailtemplates_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/emailtemplates"
)

type fakeRepo struct {
	rows    []emailtemplates.OverrideRow
	lookup  *emailtemplates.OverrideRow
	upsert  *emailtemplates.OverrideRow
	del     *emailtemplates.OverrideRow
	err     error
	lastUps *emailtemplates.OverrideRow
}

func (f *fakeRepo) ListOverrides(ctx context.Context) ([]emailtemplates.OverrideRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]emailtemplates.OverrideRow(nil), f.rows...), nil
}

func (f *fakeRepo) Lookup(ctx context.Context, key, locale string) (*emailtemplates.OverrideRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.lookup != nil {
		return f.lookup, nil
	}
	return nil, nil
}

func (f *fakeRepo) Upsert(ctx context.Context, row emailtemplates.OverrideRow, actorID uuid.UUID) (*emailtemplates.OverrideRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.lastUps = &row
	if f.upsert != nil {
		return f.upsert, nil
	}
	now := time.Now().UTC()
	out := row
	out.UpdatedAt = now
	out.UpdatedBy = &actorID
	return &out, nil
}

func (f *fakeRepo) Delete(ctx context.Context, key, locale string) (*emailtemplates.OverrideRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.del, nil
}

type fakeBaseline struct{ tpl *emailtemplates.OverrideRow }

func (b fakeBaseline) Template(key, locale string) *emailtemplates.OverrideRow {
	return b.tpl
}

type fakeValidator struct{ err error }

func (v fakeValidator) Validate(subject, bodyHTML, bodyText, key, locale string) error {
	return v.err
}

type fakeAudit struct{ events []emailtemplates.AuditEvent }

func (a *fakeAudit) Log(ctx context.Context, e emailtemplates.AuditEvent) {
	a.events = append(a.events, e)
}

func testKeysLocales() ([]string, []string) {
	return []string{"password_reset", "team_invite"}, []string{"en", "ru"}
}

func TestService_ListMatrix(t *testing.T) {
	repo := &fakeRepo{
		rows: []emailtemplates.OverrideRow{
			{Key: "password_reset", Locale: "en", Subject: "s"},
		},
	}
	keys, locs := testKeysLocales()
	svc := emailtemplates.New(emailtemplates.Deps{
		Repo:     repo,
		Baseline: fakeBaseline{},
		Keys:     keys,
		Locales:  locs,
	})
	entries, err := svc.ListMatrix(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(keys)*len(locs) {
		t.Fatalf("entries=%d want %d", len(entries), len(keys)*len(locs))
	}
	var found bool
	for _, e := range entries {
		if e.Key == "password_reset" && e.Locale == "en" {
			if !e.HasOverride {
				t.Error("password_reset:en should have override")
			}
			found = true
		}
	}
	if !found {
		t.Error("expected password_reset:en cell")
	}
}

func TestService_Get_Baseline(t *testing.T) {
	base := &emailtemplates.OverrideRow{Subject: "subj", BodyHTML: "<p>x</p>", BodyText: "x"}
	keys, locs := testKeysLocales()
	svc := emailtemplates.New(emailtemplates.Deps{
		Repo:     &fakeRepo{},
		Baseline: fakeBaseline{tpl: base},
		Keys:     keys,
		Locales:  locs,
	})
	v, err := svc.Get(context.Background(), "password_reset", "en")
	if err != nil {
		t.Fatal(err)
	}
	if !v.IsDefault || v.Subject != "subj" {
		t.Fatalf("view=%+v", v)
	}
}

func TestService_Upsert_Validation(t *testing.T) {
	repo := &fakeRepo{}
	audit := &fakeAudit{}
	keys, locs := testKeysLocales()
	svc := emailtemplates.New(emailtemplates.Deps{
		Repo:      repo,
		Baseline:  fakeBaseline{},
		Validator: fakeValidator{err: nil},
		Audit:     audit,
		Keys:      keys,
		Locales:   locs,
	})
	meta := emailtemplates.RequestMeta{IP: "127.0.0.1", UserAgent: "test"}
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	_, err := svc.Upsert(context.Background(), meta, uid, "a@b.c", "bad_key", "en", emailtemplates.UpsertCommand{
		Subject: "s", BodyHTML: "<p>x</p>", BodyText: "",
	})
	if !errors.Is(err, emailtemplates.ErrInvalidKey) {
		t.Fatalf("want ErrInvalidKey got %v", err)
	}

	_, err = svc.Upsert(context.Background(), meta, uid, "a@b.c", "password_reset", "en", emailtemplates.UpsertCommand{
		Subject: "", BodyHTML: "<p>x</p>", BodyText: "",
	})
	if !errors.Is(err, emailtemplates.ErrMissingField) {
		t.Fatalf("want ErrMissingField got %v", err)
	}
}

func TestService_Upsert_SyntaxRejected_Audit(t *testing.T) {
	repo := &fakeRepo{}
	audit := &fakeAudit{}
	keys, locs := testKeysLocales()
	svc := emailtemplates.New(emailtemplates.Deps{
		Repo:      repo,
		Baseline:  fakeBaseline{},
		Validator: fakeValidator{err: context.Canceled},
		Audit:     audit,
		Keys:      keys,
		Locales:   locs,
	})
	meta := emailtemplates.RequestMeta{}
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	_, err := svc.Upsert(context.Background(), meta, uid, "a@b.c", "password_reset", "en", emailtemplates.UpsertCommand{
		Subject: "s", BodyHTML: "<p>x</p>", BodyText: "",
	})
	if _, ok := emailtemplates.IsInvalidSyntax(err); !ok {
		t.Fatalf("want TemplateSyntaxError got %v", err)
	}
	if len(audit.events) != 1 || audit.events[0].Action != "email_template.update_rejected" {
		t.Fatalf("audit events=%v", audit.events)
	}
	if repo.lastUps != nil {
		t.Error("repo should not upsert on syntax error")
	}
}

func TestService_Upsert_Success_Audit(t *testing.T) {
	repo := &fakeRepo{}
	audit := &fakeAudit{}
	keys, locs := testKeysLocales()
	svc := emailtemplates.New(emailtemplates.Deps{
		Repo:      repo,
		Baseline:  fakeBaseline{},
		Validator: fakeValidator{err: nil},
		Audit:     audit,
		Keys:      keys,
		Locales:   locs,
	})
	meta := emailtemplates.RequestMeta{}
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	out, err := svc.Upsert(context.Background(), meta, uid, "a@b.c", "password_reset", "en", emailtemplates.UpsertCommand{
		Subject: "s", BodyHTML: "<p>{{.ResetURL}}</p>", BodyText: "t",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Subject != "s" || out.IsDefault {
		t.Fatalf("bad out %+v", out)
	}
	if len(audit.events) != 1 || audit.events[0].Action != "email_template.updated" {
		t.Fatalf("want updated audit, got %v", audit.events)
	}
}
