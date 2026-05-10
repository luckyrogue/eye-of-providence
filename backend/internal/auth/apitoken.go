// API tokens — long-lived auth для public API (export, BI, integrations).
//
// Формат plaintext: "eop_<48 hex chars>" (24 random bytes), prefix "eop_a4f3"
// сохраняется в plaintext колонке для UI; full-token хешируется sha256 и
// сравнивается constant-time. Plaintext возвращается ровно один раз — при
// создании.
//
// Scopes (TEXT):
//
//	read         — read-only: events list, summary, insights
//	write:ingest — POST /v1/ingest (для CI batchers / git hooks)
//	admin        — всё что user сам делает (super_admin scope = global_role)
//
// Revocation: soft-delete revoked_at IS NOT NULL. Middleware игнорирует
// revoked tokens; user через UI может revoke и/или create new.
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
	tokenBytesRand = 24 // 48 hex chars
	tokenPrefixLen = 8  // первые "eop_a4f3" сохраняем в БД для UI
)

// APIToken — DB-row, безопасная для отдачи пользователю (без hashed_token).
type APIToken struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Scope      string     `json:"scope"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// CreateAPIToken — генерит plaintext + сохраняет hash. Возвращает токен +
// row. Plaintext доступен ТОЛЬКО в момент создания.
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

// VerifyAPIToken — ищет active token по hash, обновляет last_used_at.
// Возвращает user_id + scope. ErrTokenInvalid если нет совпадения / revoked /
// expired.
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
	// Constant-time compare защищает от timing attacks даже хотя hashed_token
	// найден по equality в SQL — defense in depth.
	if subtle.ConstantTimeCompare([]byte(hashed), []byte(dbHash)) != 1 {
		return uuid.Nil, "", ErrTokenInvalid
	}
	// Best-effort обновление last_used_at — не блокируем request если fail.
	_, _ = pool.Exec(ctx, `UPDATE api_tokens SET last_used_at = now() WHERE id = $1`, id)
	return userID, scope, nil
}

// ListAPITokens — возвращает active tokens пользователя для UI.
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

// RevokeAPIToken — soft-delete. Возвращает true если row существовал и
// принадлежал юзеру.
func RevokeAPIToken(ctx context.Context, pool *pgxpool.Pool, userID, tokenID uuid.UUID) (bool, error) {
	tag, err := pool.Exec(ctx, `
		UPDATE api_tokens SET revoked_at = now()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`, tokenID, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// hashAPIToken — sha256 plaintext'а в hex. Не bcrypt: api token = high-entropy
// (24 random bytes), bcrypt не даёт security gain, только slow-down (мы хешируем
// at request time на каждом API call).
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
