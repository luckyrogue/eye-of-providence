// Package sso — single sign-on (OIDC сейчас, SAML позже).
//
// Архитектура:
//
//	teams ── 1:1 ── sso_configs (per-team IdP setup, owner/admin manage)
//	users.sso_team_id + users.sso_subject — стабильный link к IdP-identity.
//
// Flow (OIDC SP-initiated):
//
//	1. User → POST /v1/sso/start → backend генерит state+nonce, кладёт в
//	   sso_states (TTL 10min), редиректит на IdP authorization URL.
//	2. IdP → редирект на /v1/sso/oidc/callback?code=...&state=...
//	3. Backend проверяет state, exchange code → ID token, валидирует nonce.
//	4. JIT: если sso_subject не найден — создаём user, добавляем в team
//	   с jit_role. Иначе — login existing user.
//	5. Выпускаем JWT, редиректим на return_to.
//
// SAML (Phase 2): отдельный package sso/saml_provider.go с crewjam/saml.

package sso

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Provider — тип SSO для одной team. Расширяется когда добавится SAML.
type Provider string

const (
	ProviderOIDC Provider = "oidc"
	ProviderSAML Provider = "saml"
)

// Config — per-team SSO config row. Encoded из sso_configs table.
type Config struct {
	TeamID    uuid.UUID
	Provider  Provider
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time

	// OIDC fields. Все nullable — таблица multi-provider.
	OIDCIssuer       string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCScopes       []string

	// SAML fields (заглушки до Phase 2).
	SAMLIdPMetadata string
	SAMLSPEntityID  string
	SAMLACSURL      string

	// Common.
	AllowedDomains []string
	JITProvision   bool
	JITRole        string // "owner" | "admin" | "member"
}

// ErrConfigNotFound — нет sso_config для team_id'а. Caller должен fallback
// на standard password/GitHub auth.
var ErrConfigNotFound = errors.New("sso config not found")

// LoadConfig — читает sso_configs по team_id. Возвращает ErrConfigNotFound
// если не настроен.
func LoadConfig(ctx context.Context, pool *pgxpool.Pool, teamID uuid.UUID) (*Config, error) {
	var c Config
	c.TeamID = teamID
	err := pool.QueryRow(ctx, `
		SELECT provider, enabled,
		       COALESCE(oidc_issuer, ''),
		       COALESCE(oidc_client_id, ''),
		       COALESCE(oidc_client_secret, ''),
		       COALESCE(oidc_scopes, ARRAY[]::TEXT[]),
		       COALESCE(saml_idp_metadata, ''),
		       COALESCE(saml_sp_entity_id, ''),
		       COALESCE(saml_acs_url, ''),
		       COALESCE(allowed_domains, ARRAY[]::TEXT[]),
		       jit_provision, jit_role, created_at, updated_at
		FROM sso_configs WHERE team_id = $1`, teamID).Scan(
		&c.Provider, &c.Enabled,
		&c.OIDCIssuer, &c.OIDCClientID, &c.OIDCClientSecret, &c.OIDCScopes,
		&c.SAMLIdPMetadata, &c.SAMLSPEntityID, &c.SAMLACSURL,
		&c.AllowedDomains, &c.JITProvision, &c.JITRole,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConfigNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// SaveConfig — upsert (one row per team). created_at не трогаем при update.
func SaveConfig(ctx context.Context, pool *pgxpool.Pool, c *Config) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO sso_configs (
			team_id, provider, enabled,
			oidc_issuer, oidc_client_id, oidc_client_secret, oidc_scopes,
			saml_idp_metadata, saml_sp_entity_id, saml_acs_url,
			allowed_domains, jit_provision, jit_role
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (team_id) DO UPDATE SET
			provider = EXCLUDED.provider,
			enabled = EXCLUDED.enabled,
			oidc_issuer = EXCLUDED.oidc_issuer,
			oidc_client_id = EXCLUDED.oidc_client_id,
			oidc_client_secret = EXCLUDED.oidc_client_secret,
			oidc_scopes = EXCLUDED.oidc_scopes,
			saml_idp_metadata = EXCLUDED.saml_idp_metadata,
			saml_sp_entity_id = EXCLUDED.saml_sp_entity_id,
			saml_acs_url = EXCLUDED.saml_acs_url,
			allowed_domains = EXCLUDED.allowed_domains,
			jit_provision = EXCLUDED.jit_provision,
			jit_role = EXCLUDED.jit_role,
			updated_at = now()`,
		c.TeamID, string(c.Provider), c.Enabled,
		nullStr(c.OIDCIssuer), nullStr(c.OIDCClientID), nullStr(c.OIDCClientSecret), c.OIDCScopes,
		nullStr(c.SAMLIdPMetadata), nullStr(c.SAMLSPEntityID), nullStr(c.SAMLACSURL),
		c.AllowedDomains, c.JITProvision, c.JITRole,
	)
	return err
}

// DeleteConfig — disable SSO для team (превращает её в обычную команду).
func DeleteConfig(ctx context.Context, pool *pgxpool.Pool, teamID uuid.UUID) error {
	_, err := pool.Exec(ctx, `DELETE FROM sso_configs WHERE team_id = $1`, teamID)
	return err
}

// nullStr — convert empty string → SQL NULL (column nullable).
func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
