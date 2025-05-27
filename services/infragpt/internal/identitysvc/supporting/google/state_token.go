package google

import (
	"context"
)

type StateTokenRepository interface {
	ValidateStateToken(ctx context.Context, token string) error
	NewStateToken(ctx context.Context) (string, error)
	ExpireStateToken(ctx context.Context, token string) error
}
