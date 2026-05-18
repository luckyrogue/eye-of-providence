package prcommentapp

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	format Formatter
	post   Poster
}

type Deps struct {
	Format Formatter
	Post   Poster
}

func New(d Deps) *Service {
	return &Service{format: d.Format, post: d.Post}
}

type PostRequest struct {
	Provider      string
	Host          string
	Repo          string
	PRNumber      int
	SHAs          []string
	ProviderToken string
}

func (s *Service) PostPRComment(ctx context.Context, userID uuid.UUID, req PostRequest) (Aggregate, string, error) {
	comment, agg, err := s.format.Markdown(ctx, userID, req.SHAs)
	if err != nil {
		return agg, "", err
	}
	if err := s.post.Post(ctx, req.Provider, req.Host, req.Repo, req.PRNumber, req.ProviderToken, comment); err != nil {
		return agg, comment, err
	}
	return agg, comment, nil
}
