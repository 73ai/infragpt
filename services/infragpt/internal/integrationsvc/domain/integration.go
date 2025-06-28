package domain

import (
	"context"

	"github.com/priyanshujain/infragpt/services/infragpt"
)

type IntegrationRepository interface {
	Store(ctx context.Context, integration infragpt.Integration) error
	Update(ctx context.Context, integration infragpt.Integration) error
	FindByID(ctx context.Context, id string) (infragpt.Integration, error)
	FindByOrganization(ctx context.Context, orgID string) ([]infragpt.Integration, error)
	FindByOrganizationAndType(ctx context.Context, orgID string, connectorType infragpt.ConnectorType) ([]infragpt.Integration, error)
	FindByBotIDAndType(ctx context.Context, botID string, connectorType infragpt.ConnectorType) (infragpt.Integration, error)
	UpdateStatus(ctx context.Context, id string, status infragpt.IntegrationStatus) error
	UpdateLastUsed(ctx context.Context, id string) error
	UpdateMetadata(ctx context.Context, id string, metadata map[string]string) error
	Delete(ctx context.Context, id string) error
}
