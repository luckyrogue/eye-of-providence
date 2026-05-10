package sso

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const stateTTL = 10 * time.Minute

// State — CSRF token + nonce + return URL для одной in-flight авторизации.
type State struct {
	Value    string
	TeamID   uuid.UUID
	Nonce    string
	ReturnTo string
}

var ErrStateInvalid = errors.New("sso state invalid or expired")

// CreateState — генерит state+nonce и сохраняет в sso_states. Caller вернёт
// state в IdP authorization URL. Возвращает оба, nonce нужен для OIDC ID
// token validation (claim "nonce" должен совпасть).
func CreateState(
	ctx context.Context,
	pool *pgxpool.Pool,
	teamID uuid.UUID,
	returnTo string,
) (*State, error) {
	state, err := randomToken(24)
	if err != nil {
		return nil, err
	}
	nonce, err := randomToken(16)
	if err != nil {
		return nil, err
	}
	expires := time.Now().Add(stateTTL)
	if _, err := pool.Exec(ctx, `
		INSERT INTO sso_states (state, team_id, nonce, return_to, expires_at)
		VALUES ($1, $2, $3, $4, $5)`,
		state, teamID, nonce, returnTo, expires,
	); err != nil {
		return nil, err
	}
	return &State{Value: state, TeamID: teamID, Nonce: nonce, ReturnTo: returnTo}, nil
}

// ConsumeState — atomically находит и удаляет state. Если expired или не
// существует — ErrStateInvalid. Удаление гарантирует one-time-use (защита
// от code-replay).
func ConsumeState(
	ctx context.Context,
	pool *pgxpool.Pool,
	stateValue string,
) (*State, error) {
	var s State
	s.Value = stateValue
	err := pool.QueryRow(ctx, `
		DELETE FROM sso_states
		WHERE state = $1 AND expires_at > now()
		RETURNING team_id, nonce, COALESCE(return_to, '')`,
		stateValue,
	).Scan(&s.TeamID, &s.Nonce, &s.ReturnTo)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrStateInvalid
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// CleanupExpired — удаляет stale state-rows (TTL прошёл). Вызывается из
// background goroutine в cmd/api startup (раз в час).
func CleanupExpired(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	tag, err := pool.Exec(ctx, `DELETE FROM sso_states WHERE expires_at <= now()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
