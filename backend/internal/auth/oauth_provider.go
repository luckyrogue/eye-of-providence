package auth

import "context"

type ExternalUser struct {
	Subject string
	Email   string
	Name    string
	Login   string
}

type OAuthProvider interface {
	Name() string

	AuthCodeURL(state string) string

	Exchange(ctx context.Context, code string) (*ExternalUser, error)
}
