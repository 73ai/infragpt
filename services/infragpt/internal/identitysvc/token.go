package identitysvc

import (
	"time"
)

type RefreshToken struct {
	HashedToken string
	TokenID     string
	TokenString string
	ExpiryAt    time.Time
}

type TokenManager interface {
	NewRefreshToken(sessionID string) (RefreshToken, error)
	NewAccessToken(sessionID string) (string, error)
	ValidateRefreshToken(tokenString string) (tokenID string, err error)
	ValidateAccessToken(tokenString string) (sessionID string, err error)
}
