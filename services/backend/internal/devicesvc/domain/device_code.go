package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DeviceCodeStatus string

const (
	DeviceCodeStatusPending    DeviceCodeStatus = "pending"
	DeviceCodeStatusAuthorized DeviceCodeStatus = "authorized"
	DeviceCodeStatusExpired    DeviceCodeStatus = "expired"
	DeviceCodeStatusUsed       DeviceCodeStatus = "used"
)

type DeviceCode struct {
	ID             uuid.UUID
	DeviceCode     string
	UserCode       string
	Status         DeviceCodeStatus
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	ExpiresAt      time.Time
	CreatedAt      time.Time
}

type DeviceCodeRepository interface {
	Create(ctx context.Context, code DeviceCode) error
	GetByUserCode(ctx context.Context, userCode string) (*DeviceCode, error)
	GetByDeviceCode(ctx context.Context, deviceCode string) (*DeviceCode, error)
	Authorize(ctx context.Context, userCode string, organizationID, userID uuid.UUID) error
	MarkAsUsed(ctx context.Context, deviceCode string) error
	DeleteExpired(ctx context.Context) error
}
