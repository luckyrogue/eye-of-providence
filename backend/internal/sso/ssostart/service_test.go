package ssostart_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/sso/ssostart"
)

type fakeReg struct {
	p   ssostart.OIDCProvider
	err error
}

func (f fakeReg) GetOIDC(ctx context.Context, teamID uuid.UUID) (ssostart.OIDCProvider, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.p, nil
}

type fakeProv struct{ url string }

func (f fakeProv) AuthCodeURL(stateValue, nonce string) string {
	return f.url + "?s=" + stateValue + "&n=" + nonce
}

type fakeStates struct {
	sv, n string
	err   error
}

func (f fakeStates) CreateState(ctx context.Context, teamID uuid.UUID, returnTo string) (string, string, error) {
	if f.err != nil {
		return "", "", f.err
	}
	return f.sv, f.n, nil
}

func TestAuthorizeURL_OK(t *testing.T) {
	s := ssostart.New(fakeReg{p: fakeProv{url: "https://idp"}}, fakeStates{sv: "st", n: "no"})
	u, err := s.AuthorizeURL(context.Background(), uuid.New(), "/dash")
	if err != nil || u != "https://idp?s=st&n=no" {
		t.Fatalf("%q err=%v", u, err)
	}
}

func TestAuthorizeURL_RegErr(t *testing.T) {
	s := ssostart.New(fakeReg{err: errors.New("x")}, fakeStates{})
	_, err := s.AuthorizeURL(context.Background(), uuid.New(), "")
	if err == nil {
		t.Fatal("expected err")
	}
}
