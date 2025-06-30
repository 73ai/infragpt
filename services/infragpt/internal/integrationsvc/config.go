package integrationsvc

import (
	"database/sql"
	"fmt"

	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/connectors/github"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/connectors/slack"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/domain"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/supporting/postgres"
)

type Config struct {
	Database *sql.DB       `mapstructure:"-"`
	Slack    slack.Config  `mapstructure:"slack"`
	GitHub   github.Config `mapstructure:"github"`
}

func (c Config) New() (infragpt.IntegrationService, error) {
	integrationRepository := postgres.NewIntegrationRepository(c.Database)

	credentialRepository, err := postgres.NewCredentialRepository(c.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential repository: %w", err)
	}

	connectors := make(map[infragpt.ConnectorType]domain.Connector)

	if c.Slack.ClientID != "" {
		connectors[infragpt.ConnectorTypeSlack] = c.Slack.New()
	}

	if c.GitHub.AppID != "" {
		// Inject repository dependencies into GitHub config
		c.GitHub.GitHubRepositoryRepo = postgres.NewGitHubRepositoryRepository(c.Database)
		c.GitHub.IntegrationRepository = integrationRepository
		c.GitHub.CredentialRepository = credentialRepository

		connectors[infragpt.ConnectorTypeGithub] = c.GitHub.New()
	}

	serviceConfig := ServiceConfig{
		IntegrationRepository: integrationRepository,
		CredentialRepository:  credentialRepository,
		Connectors:            connectors,
	}

	return NewService(serviceConfig), nil
}
