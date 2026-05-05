package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type PasswordUser struct {
	ID           uuid.UUID
	Email        string
	DisplayName  string
	PasswordHash string
	TeamID       *uuid.UUID
	Role         string
}

func HashPassword(p string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(h), err
}

func VerifyPassword(hash, p string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(p)) == nil
}

// CreateUser — регистрация: email уникален, password захеширован.
func CreateUser(ctx context.Context, pool *pgxpool.Pool, email, displayName, password string, teamID *uuid.UUID) (*PasswordUser, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	id := uuid.New()
	role := "member"
	if teamID == nil {
		role = "owner" // первый пользователь, без команды — может потом создать
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, display_name, password_hash, team_id, role)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, email, displayName, hash, teamID, role)
	if err != nil {
		return nil, err
	}
	return &PasswordUser{ID: id, Email: email, DisplayName: displayName, TeamID: teamID, Role: role}, nil
}

// FindUserByEmail — для login.
func FindUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (*PasswordUser, error) {
	var u PasswordUser
	var teamID *uuid.UUID
	var name *string
	err := pool.QueryRow(ctx, `
		SELECT id, email, display_name, password_hash, team_id, COALESCE(role, 'member')
		FROM users WHERE email = $1 LIMIT 1
	`, email).Scan(&u.ID, &u.Email, &name, &u.PasswordHash, &teamID, &u.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if name != nil {
		u.DisplayName = *name
	}
	u.TeamID = teamID
	return &u, nil
}

var ErrUserNotFound = errors.New("user not found")
