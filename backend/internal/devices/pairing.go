package devices

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
	codeLen       = 6
	codeAlphabet  = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	secretBytes   = 32
	codeTTL       = 10 * time.Minute
	maxNameLen    = 64
	devicesScope  = "write:ingest"
	tokenPrefix   = "eop_"
	tokenRandHex  = 48
	tokenPrefHash = 8
)

var (
	ErrPairingNotFound = errors.New("pairing not found or expired")

	ErrSecretMismatch = errors.New("pairing secret mismatch")

	ErrCodeNotFound = errors.New("pairing code invalid or expired")

	ErrAlreadyClaimed = errors.New("pairing code already claimed")

	ErrInvalidKind = errors.New("invalid device kind")
)

var validKinds = map[string]struct{}{
	"ext":   {},
	"agent": {},
	"ide":   {},
}

type PairBeginResult struct {
	PairID    uuid.UUID `json:"pair_id"`
	Secret    string    `json:"secret"`
	Code      string    `json:"code"`
	ExpiresIn int       `json:"expires_in"`
}

type PollResult struct {
	Status  string  `json:"status"`
	Token   *string `json:"token,omitempty"`
	UserID  *string `json:"user_id,omitempty"`
	DevName *string `json:"device_name,omitempty"`
}

type Device struct {
	ID         uuid.UUID  `json:"id"`
	Kind       string     `json:"kind"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

func PairBegin(ctx context.Context, pool *pgxpool.Pool, kind, nameHint string) (PairBeginResult, error) {
	if _, ok := validKinds[kind]; !ok {
		return PairBeginResult{}, ErrInvalidKind
	}
	nameHint = strings.TrimSpace(nameHint)
	if len(nameHint) > maxNameLen {
		nameHint = nameHint[:maxNameLen]
	}
	secret, err := randomHex(secretBytes)
	if err != nil {
		return PairBeginResult{}, err
	}
	secretHash := sha256Hex(secret)
	expiresAt := time.Now().UTC().Add(codeTTL)

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		code, err := randomCode(codeLen)
		if err != nil {
			return PairBeginResult{}, err
		}
		id := uuid.New()
		_, err = pool.Exec(ctx, `
			INSERT INTO pairing_codes (id, code, secret_hash, kind, name_hint, code_expires_at)
			VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6)`,
			id, code, secretHash, kind, nameHint, expiresAt,
		)
		if err != nil {

			if isUniqueViolation(err) {
				lastErr = err
				continue
			}
			return PairBeginResult{}, err
		}
		return PairBeginResult{
			PairID:    id,
			Secret:    secret,
			Code:      code,
			ExpiresIn: int(codeTTL.Seconds()),
		}, nil
	}
	return PairBeginResult{}, fmt.Errorf("could not generate unique code after 3 attempts: %w", lastErr)
}

func Poll(ctx context.Context, pool *pgxpool.Pool, pairID uuid.UUID, secret string) (PollResult, error) {
	if secret == "" {
		return PollResult{}, ErrSecretMismatch
	}
	var (
		secretHashDB string
		expiresAt    time.Time
		tokenID      *uuid.UUID
		plainText    *string
		claimedAt    *time.Time
	)
	err := pool.QueryRow(ctx, `
		SELECT secret_hash, code_expires_at, claimed_token_id, claimed_plaintext, claimed_at
		FROM pairing_codes
		WHERE id = $1`, pairID,
	).Scan(&secretHashDB, &expiresAt, &tokenID, &plainText, &claimedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PollResult{}, ErrPairingNotFound
	}
	if err != nil {
		return PollResult{}, err
	}
	if subtle.ConstantTimeCompare([]byte(sha256Hex(secret)), []byte(secretHashDB)) != 1 {
		return PollResult{}, ErrSecretMismatch
	}

	if claimedAt == nil && time.Now().UTC().After(expiresAt) {
		return PollResult{Status: "expired"}, nil
	}
	if claimedAt == nil || tokenID == nil {
		return PollResult{Status: "pending"}, nil
	}

	var userID uuid.UUID
	var name string
	err = pool.QueryRow(ctx, `
		SELECT user_id, name FROM api_tokens WHERE id = $1`, *tokenID,
	).Scan(&userID, &name)
	if err != nil {
		return PollResult{}, err
	}
	out := PollResult{Status: "claimed"}
	uid := userID.String()
	out.UserID = &uid
	out.DevName = &name
	if plainText != nil && *plainText != "" {

		_, _ = pool.Exec(ctx, `
			UPDATE pairing_codes SET claimed_plaintext = NULL WHERE id = $1`, pairID)
		out.Token = plainText
	}
	return out, nil
}

func Claim(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, code, name string) (Device, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != codeLen {
		return Device{}, ErrCodeNotFound
	}
	name = strings.TrimSpace(name)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return Device{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		pairID    uuid.UUID
		kind      string
		nameHint  *string
		expiresAt time.Time
		claimedAt *time.Time
	)
	err = tx.QueryRow(ctx, `
		SELECT id, kind, name_hint, code_expires_at, claimed_at
		FROM pairing_codes
		WHERE code = $1
		FOR UPDATE`, code,
	).Scan(&pairID, &kind, &nameHint, &expiresAt, &claimedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Device{}, ErrCodeNotFound
	}
	if err != nil {
		return Device{}, err
	}
	if claimedAt != nil {
		return Device{}, ErrAlreadyClaimed
	}
	if time.Now().UTC().After(expiresAt) {
		return Device{}, ErrCodeNotFound
	}

	if name == "" {
		if nameHint != nil && *nameHint != "" {
			name = *nameHint
		} else {
			name = defaultName(kind)
		}
	}
	if len(name) > maxNameLen {
		name = name[:maxNameLen]
	}

	plaintext, err := randomToken()
	if err != nil {
		return Device{}, err
	}
	hashedTok := sha256Hex(plaintext)
	prefix := plaintext[:tokenPrefHash]

	out := Device{
		ID:   uuid.New(),
		Kind: kind,
		Name: name,
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO api_tokens (id, user_id, name, scope, hashed_token, prefix, kind)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, prefix`,
		out.ID, userID, name, devicesScope, hashedTok, prefix, kind,
	).Scan(&out.CreatedAt, &out.Prefix)
	if err != nil {
		return Device{}, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE pairing_codes
		   SET claimed_token_id  = $1,
		       claimed_plaintext = $2,
		       claimed_at        = now()
		 WHERE id = $3`, out.ID, plaintext, pairID)
	if err != nil {
		return Device{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Device{}, err
	}
	return out, nil
}

func List(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]Device, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, kind, name, prefix, created_at, last_used_at
		FROM api_tokens
		WHERE user_id = $1
		  AND kind IS NOT NULL
		  AND revoked_at IS NULL
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Device{}
	for rows.Next() {
		var d Device
		var kind *string
		if err := rows.Scan(&d.ID, &kind, &d.Name, &d.Prefix, &d.CreatedAt, &d.LastUsedAt); err != nil {
			return nil, err
		}
		if kind != nil {
			d.Kind = *kind
		}
		out = append(out, d)
	}
	return out, nil
}

func Revoke(ctx context.Context, pool *pgxpool.Pool, userID, deviceID uuid.UUID) (bool, error) {
	tag, err := pool.Exec(ctx, `
		UPDATE api_tokens SET revoked_at = now()
		WHERE id = $1 AND user_id = $2 AND kind IS NOT NULL AND revoked_at IS NULL`,
		deviceID, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func GCExpired(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	tag, err := pool.Exec(ctx, `
		DELETE FROM pairing_codes
		WHERE (claimed_at IS NULL AND code_expires_at < now() - interval '1 hour')
		   OR (claimed_at IS NOT NULL AND claimed_at < now() - interval '24 hours')`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func randomCode(n int) (string, error) {
	buf := make([]byte, n)
	rnd := make([]byte, n)
	if _, err := rand.Read(rnd); err != nil {
		return "", err
	}
	alphaLen := byte(len(codeAlphabet))
	for i := 0; i < n; i++ {
		buf[i] = codeAlphabet[rnd[i]%alphaLen]
	}
	return string(buf), nil
}

func randomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func randomToken() (string, error) {
	b := make([]byte, tokenRandHex/2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return tokenPrefix + hex.EncodeToString(b), nil
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func defaultName(kind string) string {
	switch kind {
	case "ext":
		return "Browser extension"
	case "agent":
		return "Desktop agent"
	case "ide":
		return "VS Code"
	default:
		return "Device"
	}
}

func isUniqueViolation(err error) bool {

	return err != nil && strings.Contains(err.Error(), "23505")
}
