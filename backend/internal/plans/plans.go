package plans

import "strings"

const (
	PlanFree       = "free"
	PlanPro        = "pro"
	PlanBusiness   = "business"
	PlanEnterprise = "enterprise"
)

type Limits struct {
	Plan            string
	MaxUsersPerTeam int
	MaxWebhooks     int
	RetentionDays   int
	SSO             bool
	AuditLog        bool
	CustomRoles     bool
	WebhookSigning  bool
}

var (
	Free = Limits{
		Plan:            PlanFree,
		MaxUsersPerTeam: 5,
		MaxWebhooks:     1,
		RetentionDays:   30,
	}
	Pro = Limits{
		Plan:            PlanPro,
		MaxUsersPerTeam: 50,
		MaxWebhooks:     0,
		RetentionDays:   365,
	}
	Business = Limits{
		Plan:           PlanBusiness,
		SSO:            true,
		AuditLog:       true,
		CustomRoles:    true,
		WebhookSigning: true,
	}
	Enterprise = Limits{
		Plan:           PlanEnterprise,
		SSO:            true,
		AuditLog:       true,
		CustomRoles:    true,
		WebhookSigning: true,
	}
)

func For(plan string) Limits {
	switch strings.ToLower(strings.TrimSpace(plan)) {
	case PlanPro:
		return Pro
	case PlanBusiness:
		return Business
	case PlanEnterprise:
		return Enterprise
	default:
		return Free
	}
}

type Service struct {
	Enforce bool
}

func (s Service) Limits(plan string) Limits {
	if !s.Enforce {
		l := Enterprise
		l.Plan = strings.ToLower(strings.TrimSpace(plan))
		if l.Plan == "" {
			l.Plan = PlanFree
		}
		return l
	}
	return For(plan)
}
