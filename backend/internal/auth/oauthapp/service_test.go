package oauthapp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth/oauthapp"
)

type fakeStore struct {
	byIdent     map[string]uuid.UUID
	byEmail     map[string]uuid.UUID
	linkErr     error
	createErr   error
	lastLink    *linkCall
	lastCreate  *createCall
	updatedMail []uuid.UUID
}

type linkCall struct {
	userID                      uuid.UUID
	provider, subject, email string
}

type createCall struct {
	newID    uuid.UUID
	provider string
	ext      oauthapp.ExternalUser
}

func (f *fakeStore) FindUserIDByIdentity(ctx context.Context, provider, subject string) (uuid.UUID, bool, error) {
	if f.byIdent == nil {
		return uuid.Nil, false, nil
	}
	k := provider + "|" + subject
	uid, ok := f.byIdent[k]
	return uid, ok, nil
}

func (f *fakeStore) UpdateUserEmailIfEmpty(ctx context.Context, userID uuid.UUID, email string) error {
	f.updatedMail = append(f.updatedMail, userID)
	return nil
}

func (f *fakeStore) FindUserIDByEmail(ctx context.Context, email string) (uuid.UUID, bool, error) {
	if f.byEmail == nil {
		return uuid.Nil, false, nil
	}
	uid, ok := f.byEmail[email]
	return uid, ok, nil
}

func (f *fakeStore) LinkIdentity(ctx context.Context, userID uuid.UUID, provider, subject, email string) error {
	f.lastLink = &linkCall{userID: userID, provider: provider, subject: subject, email: email}
	return f.linkErr
}

func (f *fakeStore) CreateUserWithIdentity(ctx context.Context, newID uuid.UUID, provider string, ext oauthapp.ExternalUser) error {
	f.lastCreate = &createCall{newID: newID, provider: provider, ext: ext}
	return f.createErr
}

func TestUpsertOAuthUser_NilStore_Deterministic(t *testing.T) {
	s := oauthapp.New(nil)
	uid, err := s.UpsertOAuthUser(context.Background(), "google", oauthapp.ExternalUser{Subject: "sub1", Email: "a@b.c"})
	if err != nil {
		t.Fatal(err)
	}
	uid2, _ := s.UpsertOAuthUser(context.Background(), "google", oauthapp.ExternalUser{Subject: "sub1", Email: "a@b.c"})
	if uid != uid2 {
		t.Fatalf("expected stable id")
	}
}

func TestUpsertOAuthUser_ExistingIdentity(t *testing.T) {
	ex := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	st := &fakeStore{byIdent: map[string]uuid.UUID{"gh|u1": ex}}
	s := oauthapp.New(st)
	uid, err := s.UpsertOAuthUser(context.Background(), "gh", oauthapp.ExternalUser{Subject: "u1", Email: "x@y.z"})
	if err != nil || uid != ex {
		t.Fatalf("uid=%v err=%v", uid, err)
	}
	if len(st.updatedMail) != 1 || st.updatedMail[0] != ex {
		t.Fatalf("expected email sync: %+v", st.updatedMail)
	}
}

func TestUpsertOAuthUser_LinkByEmail(t *testing.T) {
	link := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	st := &fakeStore{byEmail: map[string]uuid.UUID{"a@b.c": link}}
	s := oauthapp.New(st)
	uid, err := s.UpsertOAuthUser(context.Background(), "google", oauthapp.ExternalUser{Subject: "subx", Email: "a@b.c"})
	if err != nil || uid != link {
		t.Fatalf("uid=%v err=%v", uid, err)
	}
	if st.lastLink == nil || st.lastLink.userID != link {
		t.Fatalf("link: %+v", st.lastLink)
	}
}

func TestUpsertOAuthUser_Create(t *testing.T) {
	st := &fakeStore{}
	s := oauthapp.New(st)
	uid, err := s.UpsertOAuthUser(context.Background(), "github", oauthapp.ExternalUser{
		Subject: "s", Email: "new@x.com", Name: "", Login: "octo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if st.lastCreate == nil || st.lastCreate.provider != "github" || st.lastCreate.ext.Login != "octo" {
		t.Fatalf("create: %+v", st.lastCreate)
	}
	if uid != st.lastCreate.newID {
		t.Fatalf("expected returned new id")
	}
}

func TestUpsertOAuthUser_StoreError(t *testing.T) {
	st := &fakeStore{createErr: errors.New("boom")}
	s := oauthapp.New(st)
	_, err := s.UpsertOAuthUser(context.Background(), "google", oauthapp.ExternalUser{Subject: "s", Email: "e@e.e"})
	if err == nil {
		t.Fatal("expected error")
	}
}
