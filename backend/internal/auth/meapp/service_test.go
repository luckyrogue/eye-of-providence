package meapp_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth/meapp"
)

type fakeProfile struct {
	extras *meapp.ProfileExtras
	err    error
	updErr error
}

func (f fakeProfile) LoadExtras(ctx context.Context, userID uuid.UUID) (*meapp.ProfileExtras, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.extras, nil
}

func (f fakeProfile) UpdateLocale(ctx context.Context, userID uuid.UUID, locale string) error {
	return f.updErr
}

type fakeTokens struct {
	list   []meapp.TokenRow
	listErr error
	createPlain string
	createRow   meapp.TokenRow
	createErr   error
	revokeOK    bool
	revokeErr   error
}

func (f fakeTokens) List(ctx context.Context, userID uuid.UUID) ([]meapp.TokenRow, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return append([]meapp.TokenRow(nil), f.list...), nil
}

func (f fakeTokens) Create(ctx context.Context, userID uuid.UUID, name, scope string, ttl time.Duration) (string, meapp.TokenRow, error) {
	if f.createErr != nil {
		return "", meapp.TokenRow{}, f.createErr
	}
	return f.createPlain, f.createRow, nil
}

func (f fakeTokens) Revoke(ctx context.Context, userID, tokenID uuid.UUID) (bool, error) {
	if f.revokeErr != nil {
		return false, f.revokeErr
	}
	return f.revokeOK, nil
}

func TestGetProfile_JWTOnly(t *testing.T) {
	s := meapp.New(meapp.Deps{})
	out, err := s.GetProfile(context.Background(), meapp.SessionClaims{
		UserID: "u1", Email: "a@b.c", Provider: "github",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out["user_id"] != "u1" || out["email"] != "a@b.c" || out["provider"] != "github" {
		t.Fatalf("unexpected base: %#v", out)
	}
	if _, ok := out["github_login"]; ok {
		t.Fatal("expected no github_login without profile")
	}
}

func TestGetProfile_WithExtras(t *testing.T) {
	gh := "octocat"
	s := meapp.New(meapp.Deps{
		Profile: fakeProfile{extras: &meapp.ProfileExtras{GithubLogin: &gh, HasPassword: true}},
	})
	out, err := s.GetProfile(context.Background(), meapp.SessionClaims{UserID: uuid.New().String()})
	if err != nil {
		t.Fatal(err)
	}
	if out["github_login"] != "octocat" || out["has_password"] != true {
		t.Fatalf("unexpected: %#v", out)
	}
}

func TestGetProfile_InvalidSubject(t *testing.T) {
	s := meapp.New(meapp.Deps{Profile: fakeProfile{}})
	_, err := s.GetProfile(context.Background(), meapp.SessionClaims{UserID: "not-a-uuid"})
	if err == nil {
		t.Fatal("expected error")
	}
	if err != meapp.ErrInvalidSubject {
		t.Fatalf("got %v", err)
	}
}

func TestPatchLocale_Unsupported(t *testing.T) {
	s := meapp.New(meapp.Deps{Profile: fakeProfile{}})
	_, err := s.PatchLocale(context.Background(), uuid.New(), "xx")
	if err != meapp.ErrUnsupportedLocale {
		t.Fatalf("got %v", err)
	}
}

func TestPatchLocale_OK(t *testing.T) {
	s := meapp.New(meapp.Deps{Profile: fakeProfile{}})
	loc, err := s.PatchLocale(context.Background(), uuid.New(), "en")
	if err != nil || loc != "en" {
		t.Fatalf("loc=%q err=%v", loc, err)
	}
}

func TestListAPITokens_NilRepo(t *testing.T) {
	s := meapp.New(meapp.Deps{})
	got, err := s.ListAPITokens(context.Background(), uuid.New())
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v err=%v", got, err)
	}
}

func TestCreateAPIToken_Validation(t *testing.T) {
	s := meapp.New(meapp.Deps{Tokens: fakeTokens{}})
	uid := uuid.New()
	_, _, err := s.CreateAPIToken(context.Background(), uid, meapp.CreateAPITokenInput{
		Name: string(make([]byte, 65)), Scope: "read", TTLDays: 0,
	})
	if err != meapp.ErrNameTooLong {
		t.Fatalf("name: got %v", err)
	}
	_, _, err = s.CreateAPIToken(context.Background(), uid, meapp.CreateAPITokenInput{Name: "x", Scope: "read", TTLDays: 366})
	if err != meapp.ErrTTLOutOfRange {
		t.Fatalf("ttl: got %v", err)
	}
}

func TestCreateAPIToken_NoDB(t *testing.T) {
	s := meapp.New(meapp.Deps{})
	_, _, err := s.CreateAPIToken(context.Background(), uuid.New(), meapp.CreateAPITokenInput{Name: "n", Scope: "read"})
	if err != meapp.ErrDBNotConfigured {
		t.Fatalf("got %v", err)
	}
}

func TestCreateAPIToken_Delegates(t *testing.T) {
	row := meapp.TokenRow{ID: uuid.New(), Name: "n", Scope: "read"}
	toks := fakeTokens{createPlain: "eop_deadbeef", createRow: row}
	s := meapp.New(meapp.Deps{Tokens: toks})
	pt, meta, err := s.CreateAPIToken(context.Background(), uuid.New(), meapp.CreateAPITokenInput{Name: "n", Scope: "read", TTLDays: 1})
	if err != nil || pt != "eop_deadbeef" || meta.ID != row.ID {
		t.Fatalf("pt=%q meta=%+v err=%v", pt, meta, err)
	}
}

func TestRevoke_NoDB(t *testing.T) {
	s := meapp.New(meapp.Deps{})
	_, err := s.RevokeAPIToken(context.Background(), uuid.New(), uuid.New())
	if err != meapp.ErrDBNotConfigured {
		t.Fatalf("got %v", err)
	}
}
