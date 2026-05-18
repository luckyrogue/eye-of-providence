package invitesapp

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/teams/domain"
)

type pgInvites struct {
	pool *pgxpool.Pool
}

func NewPGInvites(pool *pgxpool.Pool) InviteRepository {
	return pgInvites{pool: pool}
}

func (p pgInvites) FindByCode(ctx context.Context, code string) (*domain.Invite, error) {
	if p.pool == nil {
		return nil, domain.ErrInviteInvalid
	}
	var inv domain.Invite
	err := p.pool.QueryRow(ctx, `
		SELECT id, team_id, code, max_uses, use_count, expires_at
		FROM team_invites
		WHERE code = $1
		  AND (expires_at IS NULL OR expires_at > now())
		  AND use_count < max_uses
		LIMIT 1`, code).Scan(&inv.ID, &inv.TeamID, &inv.Code, &inv.MaxUses, &inv.UseCount, &inv.Expires)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInviteInvalid
		}
		return nil, err
	}
	return &inv, nil
}

func (p pgInvites) Consume(ctx context.Context, code string, userID uuid.UUID) (uuid.UUID, error) {
	if p.pool == nil {
		return uuid.Nil, domain.ErrInviteInvalid
	}
	var teamID uuid.UUID
	err := p.pool.QueryRow(ctx, `
		UPDATE team_invites
		SET use_count = use_count + 1, used_by = $2, used_at = now()
		WHERE code = $1
		  AND use_count < max_uses
		  AND (expires_at IS NULL OR expires_at > now())
		RETURNING team_id`, code, userID).Scan(&teamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, domain.ErrInviteInvalid
		}
		return uuid.Nil, err
	}
	return teamID, nil
}

func (p pgInvites) Create(ctx context.Context, teamID, createdBy uuid.UUID, code string, maxUses int, expires time.Time, email *string) error {
	if p.pool == nil {
		return nil
	}
	_, err := p.pool.Exec(ctx, `
		INSERT INTO team_invites (team_id, code, created_by, max_uses, expires_at, email)
		VALUES ($1, $2, $3, $4, $5, $6)`, teamID, code, createdBy, maxUses, expires, email)
	return err
}

func (p pgInvites) MarkSent(ctx context.Context, code string, sentAt time.Time) error {
	if p.pool == nil {
		return nil
	}
	_, err := p.pool.Exec(ctx, `UPDATE team_invites SET sent_at = $1 WHERE code = $2`, sentAt, code)
	return err
}

type pgMembers struct {
	pool *pgxpool.Pool
}

func NewPGMembers(pool *pgxpool.Pool) MemberAdder {
	return pgMembers{pool: pool}
}

func (p pgMembers) AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	if p.pool == nil {
		return nil
	}
	_, err := p.pool.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (team_id, user_id) DO NOTHING`, teamID, userID, role)
	return err
}

type pgTeams struct {
	pool *pgxpool.Pool
}

func NewPGTeams(pool *pgxpool.Pool) TeamReader {
	return pgTeams{pool: pool}
}

func (p pgTeams) Name(ctx context.Context, teamID uuid.UUID) (string, error) {
	if p.pool == nil {
		return "", nil
	}
	var name string
	err := p.pool.QueryRow(ctx, `SELECT name FROM teams WHERE id = $1`, teamID).Scan(&name)
	return name, err
}

func (p pgTeams) Plan(ctx context.Context, teamID uuid.UUID) (string, error) {
	if p.pool == nil {
		return "", nil
	}
	var plan string
	err := p.pool.QueryRow(ctx, `SELECT subscription_plan FROM teams WHERE id=$1`, teamID).Scan(&plan)
	return plan, err
}

func (p pgTeams) MemberCount(ctx context.Context, teamID uuid.UUID) (int, error) {
	if p.pool == nil {
		return 0, nil
	}
	var n int
	err := p.pool.QueryRow(ctx, `SELECT count(*) FROM team_members WHERE team_id=$1`, teamID).Scan(&n)
	return n, err
}
