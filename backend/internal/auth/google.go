package auth

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

type GoogleIDTokenValidator func(ctx context.Context, rawIDToken, audience string) (*idtoken.Payload, error)

type GoogleOAuth struct {
	cfg       *oauth2.Config
	validator GoogleIDTokenValidator
}

var _ OAuthProvider = (*GoogleOAuth)(nil)

func NewGoogleOAuth(clientID, clientSecret, redirectURL string, opts ...GoogleOption) *GoogleOAuth {
	g := &GoogleOAuth{
		cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

type GoogleOption func(*GoogleOAuth)

func WithGoogleValidator(fn GoogleIDTokenValidator) GoogleOption {
	return func(g *GoogleOAuth) { g.validator = fn }
}

func (g *GoogleOAuth) Name() string { return "google" }

func (g *GoogleOAuth) AuthCodeURL(state string) string {
	return g.cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (g *GoogleOAuth) Exchange(ctx context.Context, code string) (*ExternalUser, error) {
	if g.cfg.ClientID == "" {
		return nil, errors.New("google oauth not configured (set EOP_GOOGLE_CLIENT_ID)")
	}
	tok, err := g.cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google token exchange: %w", err)
	}

	rawID, ok := tok.Extra("id_token").(string)
	if !ok || rawID == "" {
		return nil, errors.New("google response missing id_token")
	}

	validate := g.validator
	if validate == nil {
		validate = idtoken.Validate
	}
	payload, err := validate(ctx, rawID, g.cfg.ClientID)
	if err != nil {
		return nil, fmt.Errorf("google id_token validate: %w", err)
	}

	emailVerified, _ := payload.Claims["email_verified"].(bool)
	if !emailVerified {
		return nil, errors.New("google account email not verified")
	}
	email, _ := payload.Claims["email"].(string)
	if email == "" {
		return nil, errors.New("google id_token missing email")
	}
	name, _ := payload.Claims["name"].(string)

	return &ExternalUser{
		Subject: payload.Subject,
		Email:   email,
		Name:    name,
		Login:   "",
	}, nil
}
