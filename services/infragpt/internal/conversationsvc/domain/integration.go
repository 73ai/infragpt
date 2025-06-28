package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
)

type Integration struct {
	infragpt.Integration
	BusinessID        string
	ProviderProjectID string
}

type IntegrationRepository interface {
	Integrations(ctx context.Context, businessID uuid.UUID) ([]Integration, error)
	SaveIntegration(ctx context.Context, integration Integration) error
}
