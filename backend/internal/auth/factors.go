package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthFactors — все способы аутентификации, доступные юзеру на момент проверки.
// Используется как guard "last auth factor" при удалении:
//   - DELETE /v1/me/password         — нельзя если Total() == 1 и это password
//   - DELETE /v1/me/identities/:id   — нельзя если после удаления Total() == 0
//   - DELETE /v1/me/passkeys/:id     — нельзя если после удаления Total() == 0
//
// Total() = PasswordSet ? 1 : 0  +  Identities  +  Passkeys.
//
// Каждое поле считается ОТДЕЛЬНО: даже если юзер удаляет одну identity, мы
// передаём excludeIdentity и в подсчёт identity-count она не попадает —
// представляет "состояние ПОСЛЕ удаления". Так же для passkey credential_id.
type AuthFactors struct {
	PasswordSet bool
	Identities  int
	Passkeys    int
}

// Total — общее число доступных способов входа. 0 = lockout (юзер не сможет
// войти ни одним способом).
func (f AuthFactors) Total() int {
	n := f.Identities + f.Passkeys
	if f.PasswordSet {
		n++
	}
	return n
}

// CountAuthFactors — считает все факторы для userID, опционально исключая
// конкретную identity (по id) и/или passkey (по credential_id raw bytes).
//
// excludeIdentity == nil → не исключаем; то же для excludePasskey.
//
// Pool == nil — возвращает пустую структуру (in-memory dev mode не имеет
// understanding of factors).
func CountAuthFactors(
	ctx context.Context,
	pool *pgxpool.Pool,
	userID uuid.UUID,
	excludeIdentity *uuid.UUID,
	excludePasskey []byte,
) (AuthFactors, error) {
	out := AuthFactors{}
	if pool == nil {
		return out, nil
	}
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 1. password_hash IS NOT NULL → PasswordSet=true.
	var hasPassword bool
	if err := pool.QueryRow(c,
		`SELECT password_hash IS NOT NULL FROM users WHERE id = $1`, userID,
	).Scan(&hasPassword); err != nil {
		return out, err
	}
	out.PasswordSet = hasPassword

	// 2. count(user_identities) минус опционально исключённая.
	idQuery := `SELECT count(*) FROM user_identities WHERE user_id = $1`
	idArgs := []any{userID}
	if excludeIdentity != nil {
		idQuery += ` AND id <> $2`
		idArgs = append(idArgs, *excludeIdentity)
	}
	if err := pool.QueryRow(c, idQuery, idArgs...).Scan(&out.Identities); err != nil {
		return out, err
	}

	// 3. count(webauthn_credentials) минус опционально исключённая.
	pkQuery := `SELECT count(*) FROM webauthn_credentials WHERE user_id = $1`
	pkArgs := []any{userID}
	if len(excludePasskey) > 0 {
		pkQuery += ` AND credential_id <> $2`
		pkArgs = append(pkArgs, excludePasskey)
	}
	if err := pool.QueryRow(c, pkQuery, pkArgs...).Scan(&out.Passkeys); err != nil {
		return out, err
	}

	return out, nil
}
