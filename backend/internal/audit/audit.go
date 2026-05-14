package audit

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Action string

const (
	ActionTeamDeleted       Action = "team.deleted"
	ActionTeamCreated       Action = "team.created"
	ActionUserDeleted       Action = "user.deleted"
	ActionUserRoleChanged   Action = "user.role_changed"
	ActionMemberAdded       Action = "team.member_added"
	ActionMemberRemoved     Action = "team.member_removed"
	ActionMemberRoleChanged Action = "team.member_role_changed"
	ActionSubscriptionSet   Action = "subscription.set"
	ActionSSOSaved          Action = "sso.saved"
	ActionSSODeleted        Action = "sso.deleted"
	ActionDeviceClaimed     Action = "device.claimed"
	ActionDeviceRevoked     Action = "device.revoked"

	ActionEmailTemplateUpdated        Action = "email_template.updated"
	ActionEmailTemplateReverted       Action = "email_template.reverted"
	ActionEmailTemplateUpdateRejected Action = "email_template.update_rejected"
	ActionEmailTemplateAccessDenied   Action = "email_template.access_denied"
	ActionTeamFlagsUpdated            Action = "team.flags_updated"
	ActionTeamFlagsUpdateRejected     Action = "team.flags_update_rejected"
	ActionTeamFlagsUpdateDenied       Action = "team.flags_update_denied"
	ActionTeamPlanOverridesUpdated    Action = "team.plan_overrides_updated"
	ActionTeamPlanOverridesCleared    Action = "team.plan_overrides_cleared"
	ActionTeamPlanOverridesRejected   Action = "team.plan_overrides_update_rejected"

	ActionContentPublished       Action = "content.published"
	ActionContentDraftSaved      Action = "content.draft_saved"
	ActionContentRevertedDefault Action = "content.reverted_to_default"
	ActionContentSaveRejected    Action = "content.save_rejected"
	ActionContentAccessDenied    Action = "content.access_denied"
	ActionContentPreviewAccessed Action = "content.preview_accessed"
)

type Entry struct {
	ActorID    uuid.UUID
	ActorEmail string
	Action     Action
	TargetType string
	TargetID   string
	Metadata   map[string]any
	IP         string
	UserAgent  string
}

type Service struct {
	Pool   *pgxpool.Pool
	Logger *zap.Logger
}

func (s Service) Log(ctx context.Context, e Entry) {
	if s.Pool == nil {
		return
	}
	var meta []byte
	if len(e.Metadata) > 0 {
		b, err := json.Marshal(e.Metadata)
		if err == nil {
			meta = b
		}
	}
	var actorID *uuid.UUID
	if e.ActorID != uuid.Nil {
		actorID = &e.ActorID
	}
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO audit_log (actor_id, actor_email, action, target_type, target_id, metadata, ip, user_agent)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, NULLIF($7, ''), NULLIF($8, ''))`,
		actorID, e.ActorEmail, string(e.Action), e.TargetType, e.TargetID, meta, e.IP, e.UserAgent,
	)
	if err != nil {
		s.Logger.Warn("audit log insert failed",
			zap.String("action", string(e.Action)),
			zap.String("target", e.TargetType+"/"+e.TargetID),
			zap.Error(err))
	}
}

func (s Service) LogFromCtx(c *fiber.Ctx, actorID uuid.UUID, actorEmail string, action Action, targetType, targetID string, metadata map[string]any) {
	s.Log(c.Context(), Entry{
		ActorID:    actorID,
		ActorEmail: actorEmail,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Metadata:   metadata,
		IP:         ClientIP(c),
		UserAgent:  c.Get("User-Agent"),
	})
}

func ClientIP(c *fiber.Ctx) string {
	if xff := c.Get("X-Forwarded-For"); xff != "" {

		first := strings.SplitN(xff, ",", 2)[0]
		first = strings.TrimSpace(first)
		if first != "" {
			if len(first) > 64 {
				return first[:64]
			}
			return first
		}
	}
	ip := c.IP()
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	if len(ip) > 64 {
		return ip[:64]
	}
	return ip
}

type Row struct {
	ID         uuid.UUID       `json:"id"`
	ActorID    *uuid.UUID      `json:"actor_id,omitempty"`
	ActorEmail string          `json:"actor_email,omitempty"`
	Action     string          `json:"action"`
	TargetType string          `json:"target_type,omitempty"`
	TargetID   string          `json:"target_id,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	IP         string          `json:"ip,omitempty"`
	UserAgent  string          `json:"user_agent,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type ListFilter struct {
	Action     string
	TargetType string
	TargetID   string
	ActorID    uuid.UUID
	Limit      int
	Offset     int
}

func (s Service) List(ctx context.Context, f ListFilter) ([]Row, error) {
	if s.Pool == nil {
		return nil, errors.New("audit unavailable: pool nil")
	}
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 100
	}
	args := []any{}
	where := []string{}
	if f.Action != "" {
		args = append(args, f.Action)
		where = append(where, "action = $"+itoaArg(len(args)))
	}
	if f.TargetType != "" {
		args = append(args, f.TargetType)
		where = append(where, "target_type = $"+itoaArg(len(args)))
	}
	if f.TargetID != "" {
		args = append(args, f.TargetID)
		where = append(where, "target_id = $"+itoaArg(len(args)))
	}
	if f.ActorID != uuid.Nil {
		args = append(args, f.ActorID)
		where = append(where, "actor_id = $"+itoaArg(len(args)))
	}
	q := "SELECT id, actor_id, actor_email, action, target_type, target_id, metadata, ip, user_agent, created_at FROM audit_log"
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY created_at DESC"
	args = append(args, f.Limit)
	q += " LIMIT $" + itoaArg(len(args))
	if f.Offset > 0 {
		args = append(args, f.Offset)
		q += " OFFSET $" + itoaArg(len(args))
	}

	rows, err := s.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Row{}
	for rows.Next() {
		var r Row
		var targetType, targetID, actorEmail, ip, ua *string
		var meta []byte
		var actorID *uuid.UUID
		if err := rows.Scan(&r.ID, &actorID, &actorEmail, &r.Action, &targetType, &targetID, &meta, &ip, &ua, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.ActorID = actorID
		if actorEmail != nil {
			r.ActorEmail = *actorEmail
		}
		if targetType != nil {
			r.TargetType = *targetType
		}
		if targetID != nil {
			r.TargetID = *targetID
		}
		if ip != nil {
			r.IP = *ip
		}
		if ua != nil {
			r.UserAgent = *ua
		}
		if len(meta) > 0 {
			r.Metadata = meta
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func itoaArg(n int) string {
	switch n {
	case 1:
		return "1"
	case 2:
		return "2"
	case 3:
		return "3"
	case 4:
		return "4"
	case 5:
		return "5"
	case 6:
		return "6"
	case 7:
		return "7"
	case 8:
		return "8"
	case 9:
		return "9"
	case 10:
		return "10"
	default:
		return "?"
	}
}

func (s Service) GC(ctx context.Context, retentionDays int) (int64, error) {
	if s.Pool == nil || retentionDays <= 0 {
		return 0, nil
	}
	tag, err := s.Pool.Exec(ctx,
		`DELETE FROM audit_log WHERE created_at < now() - $1 * interval '1 day'`,
		retentionDays)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

var _ = pgx.ErrNoRows
