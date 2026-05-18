package adminapp

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/audit"
)

type TeamRow struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Plan              string     `json:"plan"`
	SubscriptionPlan  string     `json:"subscription_plan"`
	SubscriptionUntil *time.Time `json:"subscription_until"`
	SubscriptionNote  *string    `json:"subscription_note"`
	MemberCount       int        `json:"member_count"`
	OwnerEmail        *string    `json:"owner_email"`
	CreatedAt         time.Time  `json:"created_at"`
}

type UserRow struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	GlobalRole  string    `json:"global_role"`
	CreatedAt   time.Time `json:"created_at"`
	TeamsCount  int       `json:"teams_count"`
}

type Stats struct {
	UsersTotal   int `json:"users_total"`
	TeamsTotal   int `json:"teams_total"`
	MembersTotal int `json:"members_total"`
}

type RecentPayment struct {
	ID          uuid.UUID `json:"id"`
	TeamID      uuid.UUID `json:"team_id"`
	TeamName    string    `json:"team_name"`
	AmountCents *int      `json:"amount_cents,omitempty"`
	Currency    *string   `json:"currency,omitempty"`
	Method      *string   `json:"method,omitempty"`
	CoversUntil time.Time `json:"covers_until"`
	PaidAt      time.Time `json:"paid_at"`
	Note        *string   `json:"note,omitempty"`
}

type RevenueReport struct {
	TotalCents   int64           `json:"total_cents"`
	Last30dCents int64           `json:"last_30d_cents"`
	Currency     string          `json:"currency"`
	PayingTeams  int             `json:"paying_teams"`
	ByPlan       map[string]int  `json:"by_plan"`
	Recent       []RecentPayment `json:"recent"`
}

type SSOConfig struct {
	TeamID         uuid.UUID `json:"team_id"`
	TeamName       string    `json:"team_name"`
	Provider       string    `json:"provider"`
	Enabled        bool      `json:"enabled"`
	OIDCIssuer     string    `json:"oidc_issuer"`
	OIDCClientID   string    `json:"oidc_client_id"`
	AllowedDomains []string  `json:"allowed_domains"`
	JITProvision   bool      `json:"jit_provision"`
	JITRole        string    `json:"jit_role"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PaymentRow struct {
	ID          uuid.UUID `json:"id"`
	AmountCents int       `json:"amount_cents"`
	Currency    string    `json:"currency"`
	Method      string    `json:"method"`
	Note        string    `json:"note"`
	CoversUntil time.Time `json:"covers_until"`
	PaidAt      time.Time `json:"paid_at"`
	RecordedBy  uuid.UUID `json:"recorded_by"`
}

type DeleteTeamResult struct {
	TeamName string
}

type DeleteUserResult struct {
	Email string
	Role  string
}

type UpdateUserInput struct {
	TargetID    uuid.UUID
	ActorID     uuid.UUID
	GlobalRole  *string
	DisplayName *string
}

type UpdateUserResult struct {
	RoleChanged bool
	PrevRole    string
	NewRole     string
	VictimEmail string
}

type AddMemberInput struct {
	TeamID uuid.UUID
	Email  string
	Role   string
}

type AddMemberResult struct {
	UserID uuid.UUID
}

type SubscriptionPayment struct {
	AmountCents int
	Currency    string
	Method      string
	Note        string
	CoversUntil string
}

type SetSubscriptionInput struct {
	TeamID      uuid.UUID
	RecordedBy  uuid.UUID
	Plan        *string
	Until       *string
	Note        *string
	Payment     *SubscriptionPayment
}

type SetSubscriptionResult struct {
	PaymentID *uuid.UUID
	PlanNorm  string
	HasPlan   bool
	HasUntil  bool
	UntilTS   *time.Time
	ClearUntil bool
	PaymentMeta *SubscriptionPayment
}

type Store interface {
	ListTeams(ctx context.Context, limit, offset int) ([]TeamRow, error)
	ListUsers(ctx context.Context, limit, offset int) ([]UserRow, error)
	Stats(ctx context.Context) (Stats, error)
	Revenue(ctx context.Context) (RevenueReport, error)
	ListSSOConfigs(ctx context.Context) ([]SSOConfig, error)
	DisableSSO(ctx context.Context, teamID uuid.UUID) error
	ListTeamPayments(ctx context.Context, teamID uuid.UUID) ([]PaymentRow, error)
	DeleteTeam(ctx context.Context, teamID uuid.UUID) (DeleteTeamResult, error)
	DeleteUser(ctx context.Context, targetID uuid.UUID) (DeleteUserResult, error)
	UpdateUserRole(ctx context.Context, targetID uuid.UUID, role string) (prevRole, email string, err error)
	UpdateUserDisplayName(ctx context.Context, targetID uuid.UUID, displayName string) error
	AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error
	CountOtherOwnedTeams(ctx context.Context, userID, teamID uuid.UUID) (int, error)
	SetSubscription(ctx context.Context, in SetSubscriptionInput) (SetSubscriptionResult, error)
}

type AuditLister interface {
	List(ctx context.Context, f audit.ListFilter) ([]audit.Row, error)
}

type UserDeleter interface {
	DeleteUserData(ctx context.Context, userID string) error
}

type TokenBumper interface {
	BumpTokenVersion(ctx context.Context, userID uuid.UUID) error
}

type UserFinder interface {
	FindByEmail(ctx context.Context, email string) (uuid.UUID, error)
}
