package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
)

type IntegrationRepository interface {
	Store(ctx context.Context, integration infragpt.Integration) error
	Update(ctx context.Context, integration infragpt.Integration) error
	FindByID(ctx context.Context, id uuid.UUID) (infragpt.Integration, error)
	FindByOrganization(ctx context.Context, orgID uuid.UUID) ([]infragpt.Integration, error)
	FindByOrganizationAndType(ctx context.Context, orgID uuid.UUID, connectorType infragpt.ConnectorType) ([]infragpt.Integration, error)
	FindByOrganizationAndStatus(ctx context.Context, orgID uuid.UUID, status infragpt.IntegrationStatus) ([]infragpt.Integration, error)
	FindByOrganizationTypeAndStatus(ctx context.Context, orgID uuid.UUID, connectorType infragpt.ConnectorType, status infragpt.IntegrationStatus) ([]infragpt.Integration, error)
	FindByBotIDAndType(ctx context.Context, botID string, connectorType infragpt.ConnectorType) (infragpt.Integration, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status infragpt.IntegrationStatus) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
	UpdateMetadata(ctx context.Context, id uuid.UUID, metadata map[string]string) error
	Delete(ctx context.Context, id uuid.UUID) error
}
