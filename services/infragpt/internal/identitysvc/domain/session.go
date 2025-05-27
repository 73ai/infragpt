package domain

import (
	"context"

	"github.com/priyanshujain/infragpt/services/infragpt/identity"
)

type SessionRepository interface {
	StartUserSession(ctx context.Context, session identity.UserSession) (
		identity.Credentials, error)
	RefreshToken(ctx context.Context, tokenID string) (identity.Credentials, error)

	UserSession(ctx context.Context, sessionID string) (identity.UserSession, error)
	UserSessions(ctx context.Context, userID string) ([]identity.UserSession, error)
}
