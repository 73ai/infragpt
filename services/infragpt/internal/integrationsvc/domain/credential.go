package domain

import (
	"context"
	"time"

	"github.com/priyanshujain/infragpt/services/infragpt"
)

type IntegrationCredential struct {
	ID              string
	IntegrationID   string
	CredentialType  infragpt.CredentialType
	Data            map[string]string
	ExpiresAt       *time.Time
	EncryptionKeyID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CredentialRepository interface {
	Store(ctx context.Context, cred IntegrationCredential) error
	FindByIntegration(ctx context.Context, integrationID string) (IntegrationCredential, error)
	Update(ctx context.Context, cred IntegrationCredential) error
	Delete(ctx context.Context, integrationID string) error
	FindExpiring(ctx context.Context, before time.Time) ([]IntegrationCredential, error)
}