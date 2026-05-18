package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth/oauthapp"
	"github.com/eye-of-providence/backend/internal/auth/oauthflowapp"
)

func newOAuthFlowApp(s Service) *oauthflowapp.Service {
	return oauthflowapp.New(oauthflowapp.Deps{
		Linker:  oauthLinkerAdapter{inner: newOAuthAppService(s)},
		Session: newSessionApp(s.JWTSecret, s.Pool),
	})
}

type oauthLinkerAdapter struct {
	inner *oauthapp.Service
}

func (a oauthLinkerAdapter) UpsertOAuthUser(ctx context.Context, provider string, ext oauthapp.ExternalUser) (uuid.UUID, error) {
	return a.inner.UpsertOAuthUser(ctx, provider, ext)
}

type oauthProviderAdapter struct {
	p OAuthProvider
}

func (a oauthProviderAdapter) AuthCodeURL(state string) string {
	return a.p.AuthCodeURL(state)
}

func (a oauthProviderAdapter) Exchange(ctx context.Context, code string) (oauthapp.ExternalUser, error) {
	u, err := a.p.Exchange(ctx, code)
	if err != nil {
		return oauthapp.ExternalUser{}, err
	}
	return oauthapp.ExternalUser{Subject: u.Subject, Email: u.Email, Name: u.Name, Login: u.Login}, nil
}
