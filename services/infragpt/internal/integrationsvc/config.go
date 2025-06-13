package integrationsvc

import (
	"database/sql"
	"fmt"

	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/connectors/slack"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/domain"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/supporting/postgres"
)

type SlackConfig struct {
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	Scopes        []string
	SigningSecret string
}

type ServiceConfig struct {
	Database *sql.DB
	Slack    SlackConfig
}

func (c ServiceConfig) New() (infragpt.IntegrationService, error) {
	integrationRepository := postgres.NewIntegrationRepository(c.Database)
	
	credentialRepository, err := postgres.NewCredentialRepository(c.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential repository: %w", err)
	}

	connectors := make(map[infragpt.ConnectorType]domain.Connector)

	if c.Slack.ClientID != "" {
		slackConfig := slack.Config{
			ClientID:      c.Slack.ClientID,
			ClientSecret:  c.Slack.ClientSecret,
			RedirectURL:   c.Slack.RedirectURL,
			Scopes:        c.Slack.Scopes,
			SigningSecret: c.Slack.SigningSecret,
		}
		connectors[infragpt.ConnectorTypeSlack] = slack.NewConnector(slackConfig)
	}

	config := Config{
		IntegrationRepository: integrationRepository,
		CredentialRepository:  credentialRepository,
		Connectors:            connectors,
	}

	return NewService(config), nil
}