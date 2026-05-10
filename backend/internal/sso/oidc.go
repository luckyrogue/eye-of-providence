package sso

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCProvider — обёртка вокруг go-oidc/v3 provider'а + oauth2.Config.
// Кешируется per-team в OIDCRegistry; повторный AuthCodeURL/Exchange не
// делают discovery.
type OIDCProvider struct {
	cfg       *Config
	provider  *oidc.Provider
	verifier  *oidc.IDTokenVerifier
	oauth2cfg *oauth2.Config
}

// OIDCIdentity — то что мы извлекаем из ID token. Минимум, достаточный
// для JIT provisioning. Все поля validated через verifier.
type OIDCIdentity struct {
	Subject       string // "sub" claim — stable IdP identifier
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// defaultOIDCScopes — минимальный набор. Email обязателен для domain check
// и для display. Profile — для name/picture. Открытый scope "openid" required.
var defaultOIDCScopes = []string{oidc.ScopeOpenID, "email", "profile"}

// NewOIDCProvider — создаёт provider через discovery (well-known endpoint).
// Полный round-trip к IdP — caller должен закешировать результат.
//
// redirectURL — должен ТОЧНО совпадать с тем, что зарегистрирован в IdP
// (за наследничество case-sensitive). Обычно — `<public_url>/v1/sso/oidc/callback`.
func NewOIDCProvider(ctx context.Context, cfg *Config, redirectURL string) (*OIDCProvider, error) {
	if cfg.Provider != ProviderOIDC {
		return nil, fmt.Errorf("config provider is %s, not oidc", cfg.Provider)
	}
	if cfg.OIDCIssuer == "" || cfg.OIDCClientID == "" || cfg.OIDCClientSecret == "" {
		return nil, errors.New("oidc config incomplete: issuer/client_id/client_secret required")
	}

	dctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	prov, err := oidc.NewProvider(dctx, cfg.OIDCIssuer)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery failed: %w", err)
	}

	scopes := cfg.OIDCScopes
	if len(scopes) == 0 {
		scopes = defaultOIDCScopes
	}

	o2cfg := &oauth2.Config{
		ClientID:     cfg.OIDCClientID,
		ClientSecret: cfg.OIDCClientSecret,
		Endpoint:     prov.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       scopes,
	}

	verifier := prov.Verifier(&oidc.Config{ClientID: cfg.OIDCClientID})

	return &OIDCProvider{
		cfg:       cfg,
		provider:  prov,
		verifier:  verifier,
		oauth2cfg: o2cfg,
	}, nil
}

// AuthCodeURL — IdP authorization endpoint с state + nonce. Nonce binding'ит
// state с ID-token'ом — после callback'а проверяется что claim "nonce" в
// ID token == nonce, который мы выдали.
func (p *OIDCProvider) AuthCodeURL(state, nonce string) string {
	return p.oauth2cfg.AuthCodeURL(state, oidc.Nonce(nonce))
}

// Exchange — code → token + ID token validation. Проверяет:
//   - signature (через JWKS от провайдера)
//   - issuer
//   - audience (client_id)
//   - nonce (если passed) — anti-replay
//
// Возвращает identity или error если что-то не сошлось.
func (p *OIDCProvider) Exchange(ctx context.Context, code, expectedNonce string) (*OIDCIdentity, error) {
	ectx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tok, err := p.oauth2cfg.Exchange(ectx, code)
	if err != nil {
		return nil, fmt.Errorf("oidc code exchange failed: %w", err)
	}
	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, errors.New("oidc response missing id_token (IdP должен возвращать ID token при openid scope)")
	}

	idToken, err := p.verifier.Verify(ectx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("oidc id_token verify failed: %w", err)
	}

	if expectedNonce != "" && idToken.Nonce != expectedNonce {
		return nil, errors.New("oidc id_token nonce mismatch (replay attempt?)")
	}

	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("oidc claims parse failed: %w", err)
	}
	if claims.Sub == "" {
		return nil, errors.New("oidc id_token missing 'sub' claim")
	}
	if claims.Email == "" {
		return nil, errors.New("oidc id_token missing 'email' claim — IdP должен выдавать email scope")
	}

	return &OIDCIdentity{
		Subject:       claims.Sub,
		Email:         strings.ToLower(strings.TrimSpace(claims.Email)),
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		Picture:       claims.Picture,
	}, nil
}

// CheckEmailDomain — если allowed_domains настроен, email DOMAIN ДОЛЖЕН быть
// в списке. Защита от misconfigured IdP'а или скомпрометированного аккаунта
// в external IdP'е (e.g. Google personal вместо Google Workspace).
func (p *OIDCProvider) CheckEmailDomain(email string) error {
	if len(p.cfg.AllowedDomains) == 0 {
		return nil
	}
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return errors.New("email has no domain part")
	}
	domain := strings.ToLower(email[at+1:])
	for _, allowed := range p.cfg.AllowedDomains {
		if strings.EqualFold(allowed, domain) {
			return nil
		}
	}
	return fmt.Errorf("email domain %q not in allowed list", domain)
}
