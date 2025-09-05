package integrationsvc

import (
	"database/sql"
	"fmt"

	"github.com/priyanshujain/infragpt/services/backend"
	"github.com/priyanshujain/infragpt/services/backend/internal/integrationsvc/connectors/gcp"
	"github.com/priyanshujain/infragpt/services/backend/internal/integrationsvc/connectors/github"
	"github.com/priyanshujain/infragpt/services/backend/internal/integrationsvc/connectors/slack"
	"github.com/priyanshujain/infragpt/services/backend/internal/integrationsvc/domain"
	"github.com/priyanshujain/infragpt/services/backend/internal/integrationsvc/supporting/postgres"
)

type Config struct {
	Database *sql.DB       `mapstructure:"-"`
	Slack    slack.Config  `mapstructure:"slack"`
	GitHub   github.Config `mapstructure:"github"`
	GCP      gcp.Config    `mapstructure:"gcp"`
}

func (c Config) New() (backend.IntegrationService, error) {
	integrationRepository := postgres.NewIntegrationRepository(c.Database)

	credentialRepository, err := postgres.NewCredentialRepository(c.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential repository: %w", err)
	}

	connectors := make(map[backend.ConnectorType]domain.Connector)

	if c.Slack.ClientID != "" {
		connectors[backend.ConnectorTypeSlack] = c.Slack.New()
	}

	if c.GitHub.AppID != "" {
		// Inject repository dependencies into GitHub config
		c.GitHub.GitHubRepositoryRepo = postgres.NewGitHubRepositoryRepository(c.Database)
		c.GitHub.IntegrationRepository = integrationRepository
		c.GitHub.CredentialRepository = credentialRepository

		connectors[backend.ConnectorTypeGithub] = c.GitHub.New()
	}

	// GCP connector is always available (no config needed for service account auth)
	c.GCP.IntegrationRepository = integrationRepository
	c.GCP.CredentialRepository = credentialRepository
	connectors[backend.ConnectorTypeGCP] = c.GCP.New()

	serviceConfig := ServiceConfig{
		IntegrationRepository: integrationRepository,
		CredentialRepository:  credentialRepository,
		Connectors:            connectors,
	}

	return NewService(serviceConfig), nil
}
