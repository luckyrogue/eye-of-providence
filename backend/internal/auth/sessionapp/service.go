package sessionapp

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Service struct {
	signer Signer
}

type Deps struct {
	Signer Signer
}

func New(d Deps) *Service {
	return &Service{signer: d.Signer}
}

func (s *Service) IssueHandoff(ctx context.Context, userID uuid.UUID, email, method string) (string, error) {
	if s.signer == nil {
		return "", nil
	}
	return s.signer.Issue(ctx, userID, email, method)
}

// HandoffAge returns true if issued-at is older than allowed handoff window.
func HandoffAge(issuedAt *jwt.NumericDate, maxAge time.Duration) bool {
	if issuedAt == nil {
		return false
	}
	return time.Since(issuedAt.Time) > maxAge+5*time.Second
}
