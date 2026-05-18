package authapp_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/authapp"
)

type fakeAuth struct{}

func (fakeAuth) VerifyLogin(context.Context, string, string) (authapp.LoginUser, error) {
	return authapp.LoginUser{ID: uuid.New(), Email: "a@b.c"}, nil
}

func TestLogin(t *testing.T) {
	u, err := authapp.New(authapp.Deps{Auth: fakeAuth{}}).Login(context.Background(), "a@b.c", "secret")
	if err != nil || u.Email != "a@b.c" {
		t.Fatalf("u=%+v err=%v", u, err)
	}
}
