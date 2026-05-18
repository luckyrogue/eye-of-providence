package teamsapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgTeams struct {
	pool *pgxpool.Pool
}

func NewPGTeams(pool *pgxpool.Pool) TeamRepository {
	return pgTeams{pool: pool}
}

func (p pgTeams) ListForUser(ctx context.Context, userID uuid.UUID) ([]TeamRow, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT t.id, t.name, tm.role, t.subscription_plan, t.subscription_until, t.subscription_note
		FROM team_members tm JOIN teams t ON t.id = tm.team_id
		WHERE tm.user_id = $1 ORDER BY t.created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TeamRow{}
	for rows.Next() {
		var t TeamRow
		if err := rows.Scan(&t.ID, &t.Name, &t.Role, &t.SubscriptionPlan, &t.SubscriptionUntil, &t.SubscriptionNote); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (p pgTeams) GetName(ctx context.Context, teamID uuid.UUID) (string, error) {
	var name string
	err := p.pool.QueryRow(ctx, `SELECT name FROM teams WHERE id=$1`, teamID).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrTeamNotFound
	}
	return name, err
}

func (p pgTeams) Create(ctx context.Context, in CreateTeamParams) (uuid.UUID, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", in.LockID); err != nil {
		return uuid.Nil, err
	}
	if !in.IsSuper {
		var owned int
		if err := tx.QueryRow(ctx,
			"SELECT count(*) FROM team_members WHERE user_id=$1 AND role='owner'", in.UserID).Scan(&owned); err != nil {
			return uuid.Nil, err
		}
		if owned > 0 {
			return uuid.Nil, ErrOwnerLimit
		}
	}
	if in.BetaLimit > 0 && !in.IsSuper {
		var teamCount int
		if err := tx.QueryRow(ctx, "SELECT count(*) FROM teams").Scan(&teamCount); err != nil {
			return uuid.Nil, err
		}
		if teamCount >= in.BetaLimit {
			return uuid.Nil, ErrBetaFull
		}
	}
	teamID := uuid.New()
	if _, err := tx.Exec(ctx,
		"INSERT INTO teams (id, name, plan, created_by) VALUES ($1, $2, 'free', $3)",
		teamID, in.Name, in.UserID); err != nil {
		return uuid.Nil, err
	}
	if _, err := tx.Exec(ctx,
		"INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, 'owner')",
		teamID, in.UserID); err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return teamID, nil
}

func (p pgTeams) UpdateName(ctx context.Context, teamID uuid.UUID, name string) error {
	_, err := p.pool.Exec(ctx, "UPDATE teams SET name=$1 WHERE id=$2", name, teamID)
	return err
}

func (p pgTeams) Delete(ctx context.Context, teamID uuid.UUID) error {
	_, err := p.pool.Exec(ctx, "DELETE FROM teams WHERE id=$1", teamID)
	return err
}

type pgBeta struct {
	pool *pgxpool.Pool
}

func NewPGBeta(pool *pgxpool.Pool) BetaGate {
	return pgBeta{pool: pool}
}

func (p pgBeta) TeamCount(ctx context.Context) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx, "SELECT count(*) FROM teams").Scan(&n)
	return n, err
}

type pgOwnerLimit struct {
	pool *pgxpool.Pool
}

func NewPGOwnerLimit(pool *pgxpool.Pool) OwnerLimitChecker {
	return pgOwnerLimit{pool: pool}
}

func (p pgOwnerLimit) OwnedTeamCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx,
		"SELECT count(*) FROM team_members WHERE user_id=$1 AND role='owner'", userID).Scan(&n)
	return n, err
}
