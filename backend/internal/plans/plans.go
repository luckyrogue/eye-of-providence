// Package plans — feature gates per pricing tier.
//
// Тиры синхронизированы с pricing page (dashboard/src/pages/pricing).
// Меняешь cap здесь — синхронизируй copy на сайте (dashboard i18n landing.json).
//
// Service.Enforce — global kill-switch. В closed beta (default false) все
// checks no-op'ятся: возвращаются Enterprise limits (всё unlimited / включено).
// На GA включается через EOP_PLAN_LIMITS_ENFORCED=true, и beta-юзеры начнут
// получать plan_limit_exceeded на превышениях.
//
// Где enforce'ится:
//
//	teams/invites.go: count членов команды vs Limits.MaxUsersPerTeam
//	teams/teams.go (create handler): не enforce'им — лимит "1 owner = 1 team"
//	  есть отдельно, и Free→Pro upgrade не должен заставлять создать новую
//	  команду заранее.
//	webhooks/webhooks.go (Create): user-scope count vs Limits.MaxWebhooks
//	sso/admin.go (PUT): Limits.SSO
//	(будущее) audit/* : Limits.AuditLog
//	(будущее) clickhouse event TTL : Limits.RetentionDays
package plans

import "strings"

// Plan константы — те же значения, что хранятся в teams.subscription_plan.
const (
	PlanFree       = "free"
	PlanPro        = "pro"
	PlanBusiness   = "business"
	PlanEnterprise = "enterprise"
)

// Limits — feature gate values для одного тира. Поля где 0 — "unlimited"
// (gate всегда проходит). Эти числа — public contract; меняй вместе с
// pricing page.
type Limits struct {
	Plan            string
	MaxUsersPerTeam int  // 0 = unlimited
	MaxWebhooks     int  // 0 = unlimited
	RetentionDays   int  // 0 = unlimited (event store TTL, future enforcement)
	SSO             bool // OIDC / SAML
	AuditLog        bool // когда будет audit module
	CustomRoles     bool // когда будет custom RBAC
	WebhookSigning  bool // HMAC + retries (per-webhook config)
}

// Презеты соответствуют публичной странице цен. Изменения должны проходить
// через ревью pricing (это user-facing).
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

// For — возвращает презет по строке plan. Неизвестные значения mapping'уются
// в Free (safe default). Не учитывает Service.Enforce — для этого см. Service.Limits.
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

// Service — DI для guards в teams/webhooks/sso. Один глобальный инстанс
// конструируется в main.go из cfg.PlanLimitsEnforced.
type Service struct {
	// Enforce — false в бете (default), true на GA. False = no-op все checks.
	Enforce bool
}

// Limits — эффективные limits для плана с учётом kill-switch. Когда
// Enforce=false возвращает Enterprise (всё unlimited / включено), сохраняя
// исходное значение Plan (чтобы telemetry/логирование показывало настоящий
// тир, а не "enterprise" для всех).
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
