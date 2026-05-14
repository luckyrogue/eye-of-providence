package teamflags

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/plans"
)

type Service struct {
	store FlagStore
	audit AuditSink
}

type Deps struct {
	Store FlagStore
	Audit AuditSink
}

func New(d Deps) *Service {
	return &Service{store: d.Store, audit: d.Audit}
}

func PruneNullsMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		if v == nil {
			continue
		}
		out[k] = v
	}
	return out
}

func (s *Service) Get(ctx context.Context, teamID uuid.UUID) (map[string]any, error) {
	flags, err := s.store.Load(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return flags, nil
}

func (s *Service) Patch(ctx context.Context, meta RequestMeta, actorID uuid.UUID, actorEmail string, teamID uuid.UUID, flags map[string]any) (map[string]any, error) {
	if flags == nil {
		return nil, ErrMissingFlags
	}
	patch, verr := plans.ValidateFlags(flags)
	if verr != nil {
		s.logRejected(ctx, meta, actorID, actorEmail, teamID, verr)
		return nil, verr
	}
	existing, err := s.store.Load(ctx, teamID)
	if err != nil {
		return nil, err
	}
	merged := PruneNullsMap(plans.MergeFlags(existing, patch))
	mergedBytes, err := json.Marshal(merged)
	if err != nil {
		return nil, err
	}
	n, err := s.store.Save(ctx, teamID, mergedBytes)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, ErrTeamNotFound
	}
	diff := plans.FlagsDiff(existing, merged)
	s.logOK(ctx, meta, actorID, actorEmail, teamID, "team.flags_updated", map[string]any{"diff": diff})
	return merged, nil
}

func (s *Service) logRejected(ctx context.Context, meta RequestMeta, actorID uuid.UUID, actorEmail string, teamID uuid.UUID, err error) {
	if s.audit == nil {
		return
	}
	md := map[string]any{}
	var fe *plans.FlagError
	if errors.As(err, &fe) {
		md["error_code"] = fe.Code
		md["field"] = fe.Field
	} else {
		md["error_code"] = "validation_failed"
		md["error_detail"] = err.Error()
	}
	s.audit.Log(ctx, AuditEvent{
		ActorID:    actorID,
		ActorEmail: actorEmail,
		Action:     "team.flags_update_rejected",
		TargetType: "team",
		TargetID:   teamID.String(),
		Metadata:   md,
		IP:         meta.IP,
		UserAgent:  meta.UserAgent,
	})
}

func (s *Service) logOK(ctx context.Context, meta RequestMeta, actorID uuid.UUID, actorEmail string, teamID uuid.UUID, action string, md map[string]any) {
	if s.audit == nil {
		return
	}
	s.audit.Log(ctx, AuditEvent{
		ActorID:    actorID,
		ActorEmail: actorEmail,
		Action:     action,
		TargetType: "team",
		TargetID:   teamID.String(),
		Metadata:   md,
		IP:         meta.IP,
		UserAgent:  meta.UserAgent,
	})
}
