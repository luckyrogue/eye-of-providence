package auth

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

// GoogleOAuth — Google OIDC provider. Использует стандартный OAuth2 code-flow
// + id_token verification через google.golang.org/api/idtoken (валидирует
// signature через Google JWKS).
//
// Scopes: openid + email + profile (минимально достаточный набор для
// получения sub, email, email_verified, name).
//
// Security guard: callback rejects token если email_verified != true.
// Иначе чужак, подняв Google-аккаунт с известной чужой почтой, мог бы
// присвоить чужую identity в нашей системе.
// GoogleIDTokenValidator — pluggable id_token verifier. По умолчанию указывает
// на idtoken.Validate (Google's live JWKS). Тесты могут заменить функцию,
// чтобы валидировать токены, подписанные тестовым keypair'ом, без сетевых
// вызовов в Google.
type GoogleIDTokenValidator func(ctx context.Context, rawIDToken, audience string) (*idtoken.Payload, error)

type GoogleOAuth struct {
	cfg       *oauth2.Config
	validator GoogleIDTokenValidator
}

// статический check — *GoogleOAuth удовлетворяет OAuthProvider.
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

// GoogleOption — функциональная опция для NewGoogleOAuth.
type GoogleOption func(*GoogleOAuth)

// WithGoogleValidator — переопределяет id_token validator (для тестов).
// В production не используется — nil validator означает дефолт idtoken.Validate.
func WithGoogleValidator(fn GoogleIDTokenValidator) GoogleOption {
	return func(g *GoogleOAuth) { g.validator = fn }
}

// Name — provider key для interface OAuthProvider.
func (g *GoogleOAuth) Name() string { return "google" }

// AuthCodeURL — redirect URL на Google consent screen. state — anti-CSRF
// nonce; passed back в callback. AccessTypeOnline (без refresh_token) —
// нам hотя бы один обмен достаточен, мы не делаем offline access.
func (g *GoogleOAuth) AuthCodeURL(state string) string {
	return g.cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// Exchange — обмен authorization code на ExternalUser. Шаги:
//  1. POST на token endpoint Google → access_token + id_token
//  2. Validate id_token через idtoken.Validate (проверка подписи + aud + exp)
//  3. Проверка email_verified=true (см. security guard выше)
//
// Subject = `sub` claim (стабильный numeric string, не меняется при смене
// email). Email = `email` claim (если verified). Name = `name` claim.
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

	// idtoken.Validate fetches Google JWKS (cached) и проверяет подпись + iss + aud + exp.
	validate := g.validator
	if validate == nil {
		validate = idtoken.Validate
	}
	payload, err := validate(ctx, rawID, g.cfg.ClientID)
	if err != nil {
		return nil, fmt.Errorf("google id_token validate: %w", err)
	}

	// Security: email_verified=true обязателен. payload.Claims содержит decoded JSON.
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
		Login:   "", // google не имеет username concept
	}, nil
}
