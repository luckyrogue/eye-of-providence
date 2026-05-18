package domain

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	RepoURL   *string   `json:"repo_url"`
	LangPri   *string   `json:"lang_primary"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateProjectInput struct {
	Name    string
	RepoURL string
}
