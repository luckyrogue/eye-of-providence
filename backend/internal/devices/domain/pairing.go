package domain

import "github.com/google/uuid"

type PairBeginResult struct {
	PairID    uuid.UUID `json:"pair_id"`
	Secret    string    `json:"secret"`
	Code      string    `json:"code"`
	ExpiresIn int       `json:"expires_in"`
}

type PollResult struct {
	Status  string  `json:"status"`
	Token   *string `json:"token,omitempty"`
	UserID  *string `json:"user_id,omitempty"`
	DevName *string `json:"device_name,omitempty"`
}
