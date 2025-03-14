package domain

import "context"

type UserProfile struct {
	Email   string
	Name    string
	Picture string
}

type GoogleAuthGateway interface {
	AuthURL(ctx context.Context) (string, error)
	CompleteAuth(ctx context.Context, code, state string) (UserProfile, error)
}
