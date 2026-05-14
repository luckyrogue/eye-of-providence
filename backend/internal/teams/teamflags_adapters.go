package teams

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/teams/teamflags"
)

type pgTeamFlagStore struct {
	pool *pgxpool.Pool
}

func (s pgTeamFlagStore) Load(ctx context.Context, teamID uuid.UUID) (map[string]any, error) {
	if s.pool == nil {
		return map[string]any{}, nil
	}
	var raw []byte
	err := s.pool.QueryRow(ctx, "SELECT flags FROM teams WHERE id = $1", teamID).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, teamflags.ErrTeamNotFound
	}
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}, nil
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func (s pgTeamFlagStore) Save(ctx context.Context, teamID uuid.UUID, flagsJSON []byte) (int64, error) {
	if s.pool == nil {
		return 0, errors.New("nil pool")
	}
	tag, err := s.pool.Exec(ctx, "UPDATE teams SET flags = $1 WHERE id = $2", flagsJSON, teamID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

type teamflagsAuditAdapter struct {
	svc audit.Service
}

func (a teamflagsAuditAdapter) Log(ctx context.Context, e teamflags.AuditEvent) {
	if a.svc.Pool == nil {
		return
	}
	a.svc.Log(ctx, audit.Entry{
		ActorID:    e.ActorID,
		ActorEmail: e.ActorEmail,
		Action:     audit.Action(e.Action),
		TargetType: e.TargetType,
		TargetID:   e.TargetID,
		Metadata:   e.Metadata,
		IP:         e.IP,
		UserAgent:  e.UserAgent,
	})
}

func (s Service) newTeamFlagsService() *teamflags.Service {
	return teamflags.New(teamflags.Deps{
		Store: pgTeamFlagStore{pool: s.Pool},
		Audit: teamflagsAuditAdapter{svc: s.Audit},
	})
}
