package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/conversationsvc/domain"
)

type InfraGPTDB struct {
	db *sql.DB
	Querier
}

func (i *InfraGPTDB) DB() *sql.DB {
	return i.db
}

var _ domain.WorkSpaceTokenRepository = (*InfraGPTDB)(nil)
var _ domain.IntegrationRepository = (*InfraGPTDB)(nil)
var _ domain.ConversationRepository = (*InfraGPTDB)(nil)
var _ domain.ChannelRepository = (*InfraGPTDB)(nil)

func (i InfraGPTDB) SaveToken(ctx context.Context, teamID, token string) error {
	err := i.saveSlackToken(ctx, saveSlackTokenParams{
		TeamID:  teamID,
		Token:   token,
		TokenID: uuid.New(),
	})
	if err != nil {
		return fmt.Errorf("failed to save slack token: %w", err)
	}
	return nil
}

func (i InfraGPTDB) GetToken(ctx context.Context, teamID string) (string, error) {
	token, err := i.slackToken(ctx, teamID)
	if err != nil {
		return "", fmt.Errorf("failed to get slack token: %w", err)
	}
	return token, nil
}

func (i InfraGPTDB) Integrations(ctx context.Context, businessID string) ([]domain.Integration, error) {
	bid := uuid.MustParse(businessID)
	is, err := i.integrations(ctx, bid)
	if err != nil {
		return nil, err
	}

	var integrations []domain.Integration
	for _, i := range is {
		integrations = append(integrations, domain.Integration{
			Integration: infragpt.Integration{
				ConnectorType: infragpt.ConnectorType(i.Provider),
				Status:        infragpt.IntegrationStatus(i.Status),
			},
			BusinessID:        businessID,
			ProviderProjectID: i.ProviderProjectID,
		})
	}

	return integrations, nil
}

func (i InfraGPTDB) SaveIntegration(ctx context.Context, integration domain.Integration) error {
	bid := uuid.MustParse(integration.BusinessID)
	err := i.saveIntegration(ctx, saveIntegrationParams{
		ID:                uuid.New(),
		BusinessID:        bid,
		Provider:          string(integration.ConnectorType),
		ProviderProjectID: integration.ProviderProjectID,
		Status:            string(integration.Status),
	})
	if err != nil {
		return fmt.Errorf("failed to save integration: %w", err)
	}
	return nil
}
