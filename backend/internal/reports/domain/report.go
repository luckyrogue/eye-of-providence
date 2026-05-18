package domain

import "time"

type Report struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Period        string    `json:"period"`
	Model         string    `json:"model"`
	BodyMD        string    `json:"body_md"`
	GeneratedAt   time.Time `json:"generated_at"`
	PromptVersion string    `json:"prompt_version"`
}
