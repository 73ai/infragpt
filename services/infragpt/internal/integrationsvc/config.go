package integrationsvc

import (
	"database/sql"
	"fmt"

	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/domain"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/supporting/postgres"
)

type ServiceConfig struct {
	Database *sql.DB
}

func (c ServiceConfig) New() (infragpt.IntegrationService, error) {
	integrationRepository := postgres.NewIntegrationRepository(c.Database)
	
	credentialRepository, err := postgres.NewCredentialRepository(c.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential repository: %w", err)
	}

	connectors := make(map[infragpt.ConnectorType]domain.Connector)

	config := Config{
		IntegrationRepository: integrationRepository,
		CredentialRepository:  credentialRepository,
		Connectors:            connectors,
	}

	return NewService(config), nil
}