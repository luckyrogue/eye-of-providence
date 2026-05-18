package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/auth/passkeysapp"
)

func newPasskeysApp(w *WebAuthnService, pool *pgxpool.Pool) *passkeysapp.Service {
	if w == nil {
		return passkeysapp.New(passkeysapp.Deps{})
	}
	return passkeysapp.New(passkeysapp.Deps{
		RP:      passkeyRPAdapter{inner: w},
		Factors: authFactorCounterAdapter{pool: pool},
	})
}

type passkeyRPAdapter struct {
	inner *WebAuthnService
}

func (a passkeyRPAdapter) ListPasskeys(ctx context.Context, userID uuid.UUID) ([]passkeysapp.PasskeyRow, error) {
	rows, err := a.inner.ListPasskeys(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]passkeysapp.PasskeyRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, passkeysapp.PasskeyRow{
			ID: r.ID, Nickname: r.Nickname, AAGUID: r.AAGUID,
			Transports: r.Transports, CreatedAt: r.CreatedAt, LastUsedAt: r.LastUsedAt,
		})
	}
	return out, nil
}

func (a passkeyRPAdapter) PasskeyCredentialIDForUser(ctx context.Context, userID, passkeyID uuid.UUID) ([]byte, error) {
	id, err := a.inner.PasskeyCredentialIDForUser(ctx, userID, passkeyID)
	if errors.Is(err, ErrPasskeyNotFound) {
		return nil, passkeysapp.ErrPasskeyNotFound
	}
	return id, err
}

func (a passkeyRPAdapter) DeletePasskey(ctx context.Context, userID, passkeyID uuid.UUID) error {
	err := a.inner.DeletePasskey(ctx, userID, passkeyID)
	if errors.Is(err, ErrPasskeyNotFound) {
		return passkeysapp.ErrPasskeyNotFound
	}
	return err
}
