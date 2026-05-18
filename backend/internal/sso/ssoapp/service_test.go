package ssoapp_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/sso/ssoapp"
)

type fakeProv struct{}

func (fakeProv) AuthCodeURL(string, string) string { return "https://idp/authorize" }

type fakeReg struct{}

func (fakeReg) GetOIDC(context.Context, uuid.UUID) (ssoapp.OIDCProvider, error) {
	return fakeProv{}, nil
}

type fakeStates struct{}

func (fakeStates) CreateState(context.Context, uuid.UUID, string) (string, string, error) {
	return "st", "nonce", nil
}

func TestAuthorizeURL(t *testing.T) {
	svc := ssoapp.New(ssoapp.Deps{Registry: fakeReg{}, States: fakeStates{}})
	url, err := svc.AuthorizeURL(context.Background(), uuid.New(), "/")
	if err != nil || url == "" {
		t.Fatalf("url=%q err=%v", url, err)
	}
}
