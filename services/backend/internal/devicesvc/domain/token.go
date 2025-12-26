package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DeviceToken struct {
	ID             uuid.UUID
	AccessToken    string
	RefreshToken   string
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	DeviceName     string
	ExpiresAt      time.Time
	CreatedAt      time.Time
	RevokedAt      *time.Time
}

type DeviceTokenRepository interface {
	Create(ctx context.Context, token DeviceToken) error
	GetByAccessToken(ctx context.Context, accessToken string) (*DeviceToken, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (*DeviceToken, error)
	Revoke(ctx context.Context, accessToken string) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	UpdateTokens(ctx context.Context, oldRefreshToken string, token DeviceToken) error
}
