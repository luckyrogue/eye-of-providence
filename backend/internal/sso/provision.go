package sso

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User — то что provisioner возвращает caller'у для issue'инга JWT.
type User struct {
	ID           uuid.UUID
	Email        string
	DisplayName  string
	TokenVersion int
}

// ErrJITDisabled — SSO успешно, но юзера в нашей системе нет, и JIT
// провижининг выключен (jit_provision=false). Caller отдаёт 403 с понятным
// сообщением "ask admin to invite you".
var ErrJITDisabled = errors.New("user not found and JIT provisioning disabled")

// ProvisionUser — атомарно lookup-or-create user + добавление в team.
//
// Resolution order:
//  1. Найти по (sso_team_id, sso_subject) — стабильный link, email мог поменяться.
//  2. Если не нашли по subject — найти по email. Если есть юзер, привязать
//     к этому team_id/subject (claims his identity).
//  3. Если нет — JIT create (если разрешено), добавить в team с jit_role.
//
// Возвращает user + bool isNew (true если только что создан, false если
// existing). isNew пригодится для analytics / welcome email'а.
func ProvisionUser(
	ctx context.Context,
	pool *pgxpool.Pool,
	teamID uuid.UUID,
	cfg *Config,
	ident *OIDCIdentity,
) (*User, bool, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1) Lookup by (sso_team_id, sso_subject).
	var u User
	err = tx.QueryRow(ctx, `
		SELECT id, email, COALESCE(display_name, email), token_version
		FROM users WHERE sso_team_id = $1 AND sso_subject = $2`,
		teamID, ident.Subject,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.TokenVersion)
	if err == nil {
		// Existing SSO user → ensure team membership (idempotent).
		if err := ensureTeamMember(ctx, tx, teamID, u.ID, cfg.JITRole); err != nil {
			return nil, false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, false, err
		}
		return &u, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	// 2) Lookup by email (existing password-юзер, который теперь идёт через SSO).
	err = tx.QueryRow(ctx, `
		SELECT id, email, COALESCE(display_name, email), token_version
		FROM users WHERE email = $1`,
		ident.Email,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.TokenVersion)
	if err == nil {
		// Link existing user к этой team + subject.
		if _, err := tx.Exec(ctx, `
			UPDATE users SET sso_team_id = $1, sso_subject = $2 WHERE id = $3`,
			teamID, ident.Subject, u.ID,
		); err != nil {
			return nil, false, err
		}
		if err := ensureTeamMember(ctx, tx, teamID, u.ID, cfg.JITRole); err != nil {
			return nil, false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, false, err
		}
		return &u, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	// 3) JIT create. Если не разрешено — отказ.
	if !cfg.JITProvision {
		return nil, false, ErrJITDisabled
	}

	uid := uuid.New()
	displayName := ident.Name
	if displayName == "" {
		// Fallback: email local part. Не пустое имя — db CHECK не любит.
		displayName = emailLocalPart(ident.Email)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (id, email, display_name, sso_team_id, sso_subject, token_version)
		VALUES ($1, $2, $3, $4, $5, 1)`,
		uid, ident.Email, displayName, teamID, ident.Subject,
	); err != nil {
		return nil, false, err
	}
	if err := ensureTeamMember(ctx, tx, teamID, uid, cfg.JITRole); err != nil {
		return nil, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	return &User{
		ID:           uid,
		Email:        ident.Email,
		DisplayName:  displayName,
		TokenVersion: 1,
	}, true, nil
}

// ensureTeamMember — INSERT ... ON CONFLICT, чтобы повторный SSO-login не
// падал и не апгрейдил role у existing member'а (admin/owner случайно не
// downgrade'ятся до 'member').
func ensureTeamMember(ctx context.Context, tx pgx.Tx, teamID, userID uuid.UUID, role string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (team_id, user_id) DO NOTHING`,
		teamID, userID, role,
	)
	return err
}

func emailLocalPart(email string) string {
	at := strings.LastIndex(email, "@")
	if at <= 0 {
		return email
	}
	return email[:at]
}

// Touch — обновляет email/display_name если поменялись в IdP. Не критично
// для auth, но keeps display sync'ed. Вызывается опционально, errors ignored
// (audit only).
func Touch(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, ident *OIDCIdentity) {
	if ident.Name == "" {
		_, _ = pool.Exec(ctx, `UPDATE users SET email = $1 WHERE id = $2`,
			ident.Email, userID)
		return
	}
	_, _ = pool.Exec(ctx, `UPDATE users SET email = $1, display_name = $2 WHERE id = $3`,
		ident.Email, ident.Name, userID)
}
