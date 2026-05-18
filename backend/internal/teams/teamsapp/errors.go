package teamsapp

import "errors"

var (
	ErrOwnerLimit    = errors.New("owner limit")
	ErrBetaFull      = errors.New("beta full")
	ErrNotMember     = errors.New("not member")
	ErrOwnerRequired = errors.New("owner required")
	ErrInvalidName   = errors.New("invalid team name")
	ErrTeamNotFound  = errors.New("team not found")
)
