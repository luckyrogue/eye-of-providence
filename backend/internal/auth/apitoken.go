package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tokenPrefix    = "eop_"
	tokenBytesRand = 24
	tokenPrefixLen = 8
)

type APIToken struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Scope      string     `json:"scope"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

func CreateAPIToken(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, name, scope string, ttl time.Duration) (string, APIToken, error) {
	if name = strings.TrimSpace(name); name == "" {
		name = "token"
	}
	if !validScope(scope) {
		return "", APIToken{}, fmt.Errorf("invalid scope: %s", scope)
	}

	plaintext, err := generateToken()
	if err != nil {
		return "", APIToken{}, err
	}
	hashed := hashAPIToken(plaintext)
	prefix := plaintext[:tokenPrefixLen]

	var expires *time.Time
	if ttl > 0 {
		t := time.Now().UTC().Add(ttl)
		expires = &t
	}

	out := APIToken{
		ID:        uuid.New(),
		Name:      name,
		Scope:     scope,
		Prefix:    prefix,
		ExpiresAt: expires,
	}
	err = pool.QueryRow(ctx, `
		INSERT INTO api_tokens (id, user_id, name, scope, hashed_token, prefix, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at`,
		out.ID, userID, out.Name, out.Scope, hashed, out.Prefix, out.ExpiresAt,
	).Scan(&out.CreatedAt)
	if err != nil {
		return "", APIToken{}, err
	}
	return plaintext, out, nil
}

var ErrTokenInvalid = errors.New("invalid or expired api token")

func VerifyAPIToken(ctx context.Context, pool *pgxpool.Pool, plaintext string) (uuid.UUID, string, error) {
	if !strings.HasPrefix(plaintext, tokenPrefix) {
		return uuid.Nil, "", ErrTokenInvalid
	}
	hashed := hashAPIToken(plaintext)
	var (
		id     uuid.UUID
		userID uuid.UUID
		scope  string
		dbHash string
	)
	err := pool.QueryRow(ctx, `
		SELECT id, user_id, scope, hashed_token
		FROM api_tokens
		WHERE hashed_token = $1
		  AND revoked_at IS NULL
		  AND (expires_at IS NULL OR expires_at > now())`,
		hashed,
	).Scan(&id, &userID, &scope, &dbHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, "", ErrTokenInvalid
	}
	if err != nil {
		return uuid.Nil, "", err
	}

	if subtle.ConstantTimeCompare([]byte(hashed), []byte(dbHash)) != 1 {
		return uuid.Nil, "", ErrTokenInvalid
	}

	_, _ = pool.Exec(ctx, `UPDATE api_tokens SET last_used_at = now() WHERE id = $1`, id)
	return userID, scope, nil
}

func ListAPITokens(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]APIToken, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, name, scope, prefix, created_at, expires_at, last_used_at
		FROM api_tokens
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []APIToken{}
	for rows.Next() {
		var t APIToken
		if err := rows.Scan(&t.ID, &t.Name, &t.Scope, &t.Prefix, &t.CreatedAt, &t.ExpiresAt, &t.LastUsedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

func RevokeAPIToken(ctx context.Context, pool *pgxpool.Pool, userID, tokenID uuid.UUID) (bool, error) {
	tag, err := pool.Exec(ctx, `
		UPDATE api_tokens SET revoked_at = now()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`, tokenID, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func hashAPIToken(plaintext string) string {
	h := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(h[:])
}

func generateToken() (string, error) {
	b := make([]byte, tokenBytesRand)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return tokenPrefix + hex.EncodeToString(b), nil
}

func validScope(s string) bool {
	switch s {
	case "read", "write:ingest", "admin":
		return true
	}
	return false
}
