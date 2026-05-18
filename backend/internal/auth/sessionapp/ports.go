package sessionapp

import (
	"context"

	"github.com/google/uuid"
)

type Signer interface {
	Issue(ctx context.Context, userID uuid.UUID, email, method string) (token string, err error)
}
