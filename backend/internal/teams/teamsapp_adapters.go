package teams

import (
	"github.com/eye-of-providence/backend/internal/teams/teamsapp"
)

func (s *Service) teamsApp() *teamsapp.Service {
	return teamsapp.New(teamsapp.Deps{
		Teams:  teamsapp.NewPGTeams(s.Pool),
		Beta:   teamsapp.NewPGBeta(s.Pool),
		Owners: teamsapp.NewPGOwnerLimit(s.Pool),
	})
}
