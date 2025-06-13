package integrationsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
)

type service struct {
}

func NewService() infragpt.IntegrationService {
	return &service{}
}

func (s *service) NewIntegration(ctx context.Context, cmd infragpt.NewIntegrationCommand) (infragpt.IntegrationAuthorizationIntent, error) {
	switch cmd.ConnectorType {
	case infragpt.ConnectorTypeSlack:
		return infragpt.IntegrationAuthorizationIntent{
			Type: infragpt.AuthorizationTypeOAuth2,
			URL:  "https://slack.com/oauth/v2/authorize?client_id=dummy&scope=chat:write,channels:read&state=dummy_state",
		}, nil
	case infragpt.ConnectorTypeGithub:
		return infragpt.IntegrationAuthorizationIntent{
			Type: infragpt.AuthorizationTypeInstallation,
			URL:  "https://github.com/apps/dummy-app/installations/new",
		}, nil
	default:
		return infragpt.IntegrationAuthorizationIntent{}, fmt.Errorf("unsupported connector type: %s", cmd.ConnectorType)
	}
}

func (s *service) AuthorizeIntegration(ctx context.Context, cmd infragpt.AuthorizeIntegrationCommand) (infragpt.Integration, error) {
	now := time.Now()
	return infragpt.Integration{
		ID:                      uuid.New().String(),
		OrganizationID:          cmd.OrganizationID,
		UserID:                  "dummy-user-id",
		ConnectorType:           cmd.ConnectorType,
		Status:                  infragpt.IntegrationStatusActive,
		BotID:                   "dummy-bot-id",
		ConnectorUserID:         "dummy-connector-user-id",
		ConnectorOrganizationID: "dummy-connector-org-id",
		Metadata:                map[string]string{"dummy": "metadata"},
		CreatedAt:               now,
		UpdatedAt:               now,
		LastUsedAt:              &now,
	}, nil
}

func (s *service) RevokeIntegration(ctx context.Context, cmd infragpt.RevokeIntegrationCommand) error {
	return nil
}

func (s *service) Integrations(ctx context.Context, query infragpt.IntegrationsQuery) ([]infragpt.Integration, error) {
	now := time.Now()
	dummyIntegrations := []infragpt.Integration{
		{
			ID:                      uuid.New().String(),
			OrganizationID:          query.OrganizationID,
			UserID:                  "dummy-user-id-1",
			ConnectorType:           infragpt.ConnectorTypeSlack,
			Status:                  infragpt.IntegrationStatusActive,
			BotID:                   "dummy-slack-bot-id",
			ConnectorUserID:         "dummy-slack-user-id",
			ConnectorOrganizationID: "dummy-slack-org-id",
			Metadata:                map[string]string{"team_name": "Dummy Team"},
			CreatedAt:               now.Add(-24 * time.Hour),
			UpdatedAt:               now.Add(-1 * time.Hour),
			LastUsedAt:              &now,
		},
		{
			ID:                      uuid.New().String(),
			OrganizationID:          query.OrganizationID,
			UserID:                  "dummy-user-id-2",
			ConnectorType:           infragpt.ConnectorTypeGithub,
			Status:                  infragpt.IntegrationStatusActive,
			BotID:                   "dummy-github-installation-id",
			ConnectorUserID:         "dummy-github-user-id",
			ConnectorOrganizationID: "dummy-github-org-id",
			Metadata:                map[string]string{"org_name": "Dummy Org"},
			CreatedAt:               now.Add(-48 * time.Hour),
			UpdatedAt:               now.Add(-2 * time.Hour),
			LastUsedAt:              &now,
		},
	}

	if query.ConnectorType != "" {
		var filtered []infragpt.Integration
		for _, integration := range dummyIntegrations {
			if integration.ConnectorType == query.ConnectorType {
				filtered = append(filtered, integration)
			}
		}
		return filtered, nil
	}

	return dummyIntegrations, nil
}

func (s *service) Integration(ctx context.Context, query infragpt.IntegrationQuery) (infragpt.Integration, error) {
	now := time.Now()
	return infragpt.Integration{
		ID:                      query.IntegrationID,
		OrganizationID:          query.OrganizationID,
		UserID:                  "dummy-user-id",
		ConnectorType:           infragpt.ConnectorTypeSlack,
		Status:                  infragpt.IntegrationStatusActive,
		BotID:                   "dummy-bot-id",
		ConnectorUserID:         "dummy-connector-user-id",
		ConnectorOrganizationID: "dummy-connector-org-id",
		Metadata:                map[string]string{"dummy": "metadata"},
		CreatedAt:               now.Add(-24 * time.Hour),
		UpdatedAt:               now.Add(-1 * time.Hour),
		LastUsedAt:              &now,
	}, nil
}
