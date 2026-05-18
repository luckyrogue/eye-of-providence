package teams

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/mailer"
	"github.com/eye-of-providence/backend/internal/plans"
	"github.com/eye-of-providence/backend/internal/teams/invitesapp"
)

func (s *Service) invitesApp() *invitesapp.Service {
	return invitesapp.New(invitesapp.Deps{
		Invites: invitesapp.NewPGInvites(s.Pool),
		Members: invitesapp.NewPGMembers(s.Pool),
		Teams:   invitesapp.NewPGTeams(s.Pool),
		Mail:    inviteMailAdapter{s: s},
		Plans:   planLimiterAdapter{svc: s.Plans},
		Roles:   memberRoleAdapter{s: s},
	})
}

type planLimiterAdapter struct {
	svc plans.Service
}

func (a planLimiterAdapter) Limits(plan string) invitesapp.PlanLimits {
	lim := a.svc.Limits(plan)
	return invitesapp.PlanLimits{Plan: lim.Plan, MaxUsersPerTeam: lim.MaxUsersPerTeam}
}

type inviteMailAdapter struct {
	s *Service
}

func (a inviteMailAdapter) Send(ctx context.Context, teamID, inviterID uuid.UUID, to, code string) error {
	if a.s.Mailer == nil || a.s.Pool == nil {
		return nil
	}
	var teamName, inviterName string
	var inviterLocale *string
	_ = a.s.Pool.QueryRow(ctx, `SELECT name FROM teams WHERE id = $1`, teamID).Scan(&teamName)
	_ = a.s.Pool.QueryRow(ctx,
		`SELECT COALESCE(display_name, ''), locale FROM users WHERE id = $1`, inviterID,
	).Scan(&inviterName, &inviterLocale)

	loc := mailer.Locale("")
	if inviterLocale != nil {
		loc = mailer.Locale(*inviterLocale)
	}
	inviteURL := strings.TrimRight(a.s.PublicURL, "/") + "/?invite=" + code
	subject, html, text := mailer.InviteEmail(teamName, inviteURL, inviterName, loc)
	return a.s.Mailer.Send(ctx, to, subject, html, text)
}
