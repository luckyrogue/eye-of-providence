package domain

import "github.com/google/uuid"

type SSOConfig struct {
	TeamID       uuid.UUID `json:"team_id"`
	Provider     string    `json:"provider"`
	ClientID     string    `json:"client_id"`
	IssuerURL    string    `json:"issuer_url"`
	Enabled      bool      `json:"enabled"`
	EnforceSSO   bool      `json:"enforce_sso"`
}
