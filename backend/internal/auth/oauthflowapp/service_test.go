package oauthflowapp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth/oauthapp"
	"github.com/eye-of-providence/backend/internal/auth/oauthflowapp"
)

type fakeProv struct{}

func (fakeProv) AuthCodeURL(string) string { return "https://idp" }

func (fakeProv) Exchange(context.Context, string) (oauthapp.ExternalUser, error) {
	return oauthapp.ExternalUser{Subject: "s", Email: "a@b.c"}, nil
}

type fakeLinker struct{}

func (fakeLinker) UpsertOAuthUser(context.Context, string, oauthapp.ExternalUser) (uuid.UUID, error) {
	return uuid.New(), nil
}

type fakeSession struct{}

func (fakeSession) IssueHandoff(context.Context, uuid.UUID, string, string) (string, error) {
	return "jwt", nil
}

func TestCompleteCallback(t *testing.T) {
	svc := oauthflowapp.New(oauthflowapp.Deps{Linker: fakeLinker{}, Session: fakeSession{}})
	st := svc.StateCookieValue("abc", "/dash")
	res, err := svc.CompleteCallback(context.Background(), "github", fakeProv{}, oauthflowapp.CallbackInput{
		GotState: "abc", StoredStateCookie: st, Code: "code",
	})
	if err != nil || res.Token != "jwt" || res.ReturnTo != "/dash" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestStateMismatch(t *testing.T) {
	svc := oauthflowapp.New(oauthflowapp.Deps{Linker: fakeLinker{}, Session: fakeSession{}})
	_, err := svc.CompleteCallback(context.Background(), "github", fakeProv{}, oauthflowapp.CallbackInput{
		GotState: "x", StoredStateCookie: "y", Code: "c",
	})
	if !errors.Is(err, oauthflowapp.ErrStateMismatch) {
		t.Fatalf("err=%v", err)
	}
}
