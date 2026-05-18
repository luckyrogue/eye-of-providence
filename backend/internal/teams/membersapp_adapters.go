package teams

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/eye-of-providence/backend/internal/teams/domain"
	"github.com/eye-of-providence/backend/internal/teams/invitesapp"
	"github.com/eye-of-providence/backend/internal/teams/membersapp"
)

type pgMemberRepo struct{ s *Service }

func (r pgMemberRepo) ListByTeam(ctx context.Context, teamID domain.TeamID) ([]domain.Member, error) {
	rows, err := r.s.Pool.Query(ctx, `
		SELECT u.id, u.email, COALESCE(u.display_name, u.email), tm.role, tm.joined_at
		FROM team_members tm JOIN users u ON u.id = tm.user_id
		WHERE tm.team_id = $1 ORDER BY tm.joined_at`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Member{}
	for rows.Next() {
		var m domain.Member
		if err := rows.Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

type pgRoleReader struct{ s *Service }

func (r pgRoleReader) TeamRole(ctx context.Context, userID, teamID uuid.UUID) (string, bool) {
	if r.s.Pool == nil {
		return "", false
	}
	var role string
	err := r.s.Pool.QueryRow(ctx,
		"SELECT role FROM team_members WHERE user_id=$1 AND team_id=$2",
		userID, teamID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return role, true
}

type pgMemberMutator struct{ s *Service }

func (m pgMemberMutator) TargetRole(ctx context.Context, teamID, userID uuid.UUID) (string, error) {
	var role string
	err := m.s.Pool.QueryRow(ctx,
		"SELECT role FROM team_members WHERE team_id=$1 AND user_id=$2",
		teamID, userID).Scan(&role)
	return role, err
}

func (m pgMemberMutator) OwnerCount(ctx context.Context, teamID uuid.UUID) (int, error) {
	var n int
	err := m.s.Pool.QueryRow(ctx,
		"SELECT count(*) FROM team_members WHERE team_id=$1 AND role='owner'", teamID).Scan(&n)
	return n, err
}

func (m pgMemberMutator) OwnedTeamCount(ctx context.Context, userID, excludeTeamID uuid.UUID) (int, error) {
	var n int
	err := m.s.Pool.QueryRow(ctx,
		"SELECT count(*) FROM team_members WHERE user_id=$1 AND role='owner' AND team_id<>$2",
		userID, excludeTeamID).Scan(&n)
	return n, err
}

func (m pgMemberMutator) UpdateRole(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	_, err := m.s.Pool.Exec(ctx,
		"UPDATE team_members SET role=$1 WHERE team_id=$2 AND user_id=$3",
		role, teamID, userID)
	return err
}

func (m pgMemberMutator) Remove(ctx context.Context, teamID, userID uuid.UUID) error {
	_, err := m.s.Pool.Exec(ctx,
		"DELETE FROM team_members WHERE team_id=$1 AND user_id=$2",
		teamID, userID)
	return err
}

type eventStoreActivity struct {
	store EventStoreLike
}

func (e eventStoreActivity) AggregateByCategoryBulk(ctx context.Context, userIDs []string, since time.Time) (map[string]map[string]uint64, error) {
	if e.store == nil {
		return map[string]map[string]uint64{}, nil
	}
	return e.store.AggregateByCategoryBulk(ctx, userIDs, since)
}

func (s *Service) membersApp() *membersapp.Service {
	return membersapp.New(membersapp.Deps{
		Members:  pgMemberRepo{s: s},
		Roles:    pgRoleReader{s: s},
		Mutator:  pgMemberMutator{s: s},
		Activity: eventStoreActivity{store: s.EventStore},
	})
}

type memberRoleAdapter struct{ s *Service }

func (a memberRoleAdapter) TeamRole(ctx context.Context, userID, teamID uuid.UUID) (string, bool) {
	return pgRoleReader(a).TeamRole(ctx, userID, teamID)
}

var _ invitesapp.RoleChecker = memberRoleAdapter{}
