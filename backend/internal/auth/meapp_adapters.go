package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/auth/meapp"
)

type pgMeProfile struct {
	pool *pgxpool.Pool
}

func (p pgMeProfile) LoadExtras(ctx context.Context, userID uuid.UUID) (*meapp.ProfileExtras, error) {
	if p.pool == nil {
		return nil, nil
	}
	var ghLogin, globalRole, displayName, lastName, phone, locale *string
	var hasPassword bool
	var createdAt time.Time
	err := p.pool.QueryRow(ctx,
		"SELECT github_login, global_role, display_name, last_name, phone, locale, password_hash IS NOT NULL, created_at FROM users WHERE id = $1", userID,
	).Scan(&ghLogin, &globalRole, &displayName, &lastName, &phone, &locale, &hasPassword, &createdAt)
	if err != nil {
		// Совместимость с прежним handler: ошибки обогащения профиля не ломают GET /v1/me.
		return nil, nil
	}
	ex := &meapp.ProfileExtras{
		GithubLogin: ghLogin, GlobalRole: globalRole, DisplayName: displayName,
		LastName: lastName, Phone: phone, Locale: locale, HasPassword: hasPassword,
	}
	if !createdAt.IsZero() {
		ex.CreatedAtRFC = createdAt.UTC().Format(time.RFC3339)
	}
	return ex, nil
}

func (p pgMeProfile) UpdateLocale(ctx context.Context, userID uuid.UUID, locale string) error {
	if p.pool == nil {
		return nil
	}
	_, err := p.pool.Exec(ctx, `UPDATE users SET locale = $1 WHERE id = $2`, locale, userID)
	return err
}

type pgMeTokens struct {
	pool *pgxpool.Pool
}

func (t pgMeTokens) List(ctx context.Context, userID uuid.UUID) ([]meapp.TokenRow, error) {
	if t.pool == nil {
		return []meapp.TokenRow{}, nil
	}
	toks, err := ListAPITokens(ctx, t.pool, userID)
	if err != nil {
		return nil, err
	}
	out := make([]meapp.TokenRow, 0, len(toks))
	for _, a := range toks {
		out = append(out, apiTokenToRow(a))
	}
	return out, nil
}

func (t pgMeTokens) Create(ctx context.Context, userID uuid.UUID, name, scope string, ttl time.Duration) (string, meapp.TokenRow, error) {
	pt, row, err := CreateAPIToken(ctx, t.pool, userID, name, scope, ttl)
	if err != nil {
		return "", meapp.TokenRow{}, err
	}
	return pt, apiTokenToRow(row), nil
}

func (t pgMeTokens) Revoke(ctx context.Context, userID, tokenID uuid.UUID) (bool, error) {
	return RevokeAPIToken(ctx, t.pool, userID, tokenID)
}

func apiTokenToRow(a APIToken) meapp.TokenRow {
	return meapp.TokenRow{
		ID: a.ID, Name: a.Name, Scope: a.Scope, Prefix: a.Prefix,
		CreatedAt: a.CreatedAt, ExpiresAt: a.ExpiresAt, LastUsedAt: a.LastUsedAt,
	}
}

func newMeAppService(s MeService) *meapp.Service {
	return meapp.New(meapp.Deps{
		Profile: pgMeProfile{pool: s.Pool},
		Tokens:  pgMeTokens{pool: s.Pool},
	})
}
