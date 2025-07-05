package domain

import "context"

type AuthService interface {
	Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error
}
