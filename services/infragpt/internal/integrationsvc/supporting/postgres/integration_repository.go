package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/domain"
	"github.com/sqlc-dev/pqtype"
)

type integrationRepository struct {
	queries *Queries
}

func NewIntegrationRepository(sqlDB *sql.DB) domain.IntegrationRepository {
	return &integrationRepository{
		queries: New(sqlDB),
	}
}

func (r *integrationRepository) Store(ctx context.Context, integration infragpt.Integration) error {
	metadata := make(map[string]any)
	for k, v := range integration.Metadata {
		metadata[k] = v
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	integrationID, err := uuid.Parse(integration.ID)
	if err != nil {
		return fmt.Errorf("invalid integration ID: %w", err)
	}

	organizationID, err := uuid.Parse(integration.OrganizationID)
	if err != nil {
		return fmt.Errorf("invalid organization ID: %w", err)
	}

	userID, err := uuid.Parse(integration.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	var botID sql.NullString
	if integration.BotID != "" {
		botID = sql.NullString{String: integration.BotID, Valid: true}
	}

	var connectorUserID sql.NullString
	if integration.ConnectorUserID != "" {
		connectorUserID = sql.NullString{String: integration.ConnectorUserID, Valid: true}
	}

	var connectorOrganizationID sql.NullString
	if integration.ConnectorOrganizationID != "" {
		connectorOrganizationID = sql.NullString{String: integration.ConnectorOrganizationID, Valid: true}
	}

	var lastUsedAt sql.NullTime
	if integration.LastUsedAt != nil {
		lastUsedAt = sql.NullTime{Time: *integration.LastUsedAt, Valid: true}
	}

	return r.queries.StoreIntegration(ctx, StoreIntegrationParams{
		ID:                      integrationID,
		OrganizationID:          organizationID,
		UserID:                  userID,
		ConnectorType:           string(integration.ConnectorType),
		Status:                  string(integration.Status),
		BotID:                   botID,
		ConnectorUserID:         connectorUserID,
		ConnectorOrganizationID: connectorOrganizationID,
		Metadata:                pqtype.NullRawMessage{RawMessage: metadataJSON, Valid: true},
		CreatedAt:               integration.CreatedAt,
		UpdatedAt:               integration.UpdatedAt,
		LastUsedAt:              lastUsedAt,
	})
}

func (r *integrationRepository) Update(ctx context.Context, integration infragpt.Integration) error {
	metadata := make(map[string]any)
	for k, v := range integration.Metadata {
		metadata[k] = v
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	integrationID, err := uuid.Parse(integration.ID)
	if err != nil {
		return fmt.Errorf("invalid integration ID: %w", err)
	}

	var botID sql.NullString
	if integration.BotID != "" {
		botID = sql.NullString{String: integration.BotID, Valid: true}
	}

	var connectorUserID sql.NullString
	if integration.ConnectorUserID != "" {
		connectorUserID = sql.NullString{String: integration.ConnectorUserID, Valid: true}
	}

	var connectorOrganizationID sql.NullString
	if integration.ConnectorOrganizationID != "" {
		connectorOrganizationID = sql.NullString{String: integration.ConnectorOrganizationID, Valid: true}
	}

	var lastUsedAt sql.NullTime
	if integration.LastUsedAt != nil {
		lastUsedAt = sql.NullTime{Time: *integration.LastUsedAt, Valid: true}
	}

	return r.queries.UpdateIntegration(ctx, UpdateIntegrationParams{
		ID:                      integrationID,
		ConnectorType:           string(integration.ConnectorType),
		Status:                  string(integration.Status),
		BotID:                   botID,
		ConnectorUserID:         connectorUserID,
		ConnectorOrganizationID: connectorOrganizationID,
		Metadata:                pqtype.NullRawMessage{RawMessage: metadataJSON, Valid: true},
		UpdatedAt:               integration.UpdatedAt,
		LastUsedAt:              lastUsedAt,
	})
}

func (r *integrationRepository) FindByID(ctx context.Context, id string) (infragpt.Integration, error) {
	integrationID, err := uuid.Parse(id)
	if err != nil {
		return infragpt.Integration{}, fmt.Errorf("invalid integration ID: %w", err)
	}

	dbIntegration, err := r.queries.FindIntegrationByID(ctx, integrationID)
	if err != nil {
		return infragpt.Integration{}, fmt.Errorf("failed to find integration: %w", err)
	}

	return r.toSpecIntegration(dbIntegration)
}

func (r *integrationRepository) FindByOrganization(ctx context.Context, orgID string) ([]infragpt.Integration, error) {
	organizationID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	dbIntegrations, err := r.queries.FindIntegrationsByOrganization(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find integrations: %w", err)
	}

	integrations := make([]infragpt.Integration, len(dbIntegrations))
	for i, dbIntegration := range dbIntegrations {
		integration, err := r.toSpecIntegration(dbIntegration)
		if err != nil {
			return nil, fmt.Errorf("failed to map integration: %w", err)
		}
		integrations[i] = integration
	}

	return integrations, nil
}

func (r *integrationRepository) FindByOrganizationAndType(ctx context.Context, orgID string, connectorType infragpt.ConnectorType) ([]infragpt.Integration, error) {
	organizationID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	dbIntegrations, err := r.queries.FindIntegrationsByOrganizationAndType(ctx, FindIntegrationsByOrganizationAndTypeParams{
		OrganizationID: organizationID,
		ConnectorType:  string(connectorType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find integrations: %w", err)
	}

	integrations := make([]infragpt.Integration, len(dbIntegrations))
	for i, dbIntegration := range dbIntegrations {
		integration, err := r.toSpecIntegration(dbIntegration)
		if err != nil {
			return nil, fmt.Errorf("failed to map integration: %w", err)
		}
		integrations[i] = integration
	}

	return integrations, nil
}

func (r *integrationRepository) UpdateStatus(ctx context.Context, id string, status infragpt.IntegrationStatus) error {
	integrationID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid integration ID: %w", err)
	}

	return r.queries.UpdateIntegrationStatus(ctx, UpdateIntegrationStatusParams{
		ID:     integrationID,
		Status: string(status),
	})
}

func (r *integrationRepository) UpdateLastUsed(ctx context.Context, id string) error {
	integrationID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid integration ID: %w", err)
	}

	return r.queries.UpdateIntegrationLastUsed(ctx, integrationID)
}

func (r *integrationRepository) UpdateMetadata(ctx context.Context, id string, metadata map[string]string) error {
	integrationID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid integration ID: %w", err)
	}

	// Convert metadata to map[string]any for JSON marshaling
	metadataMap := make(map[string]any)
	for k, v := range metadata {
		metadataMap[k] = v
	}

	metadataJSON, err := json.Marshal(metadataMap)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return r.queries.UpdateIntegrationMetadata(ctx, UpdateIntegrationMetadataParams{
		ID:       integrationID,
		Metadata: pqtype.NullRawMessage{RawMessage: metadataJSON, Valid: true},
	})
}

func (r *integrationRepository) Delete(ctx context.Context, id string) error {
	integrationID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid integration ID: %w", err)
	}

	return r.queries.DeleteIntegration(ctx, integrationID)
}

func (r *integrationRepository) FindByBotIDAndType(ctx context.Context, botID string, connectorType infragpt.ConnectorType) (infragpt.Integration, error) {
	dbIntegration, err := r.queries.FindIntegrationByBotIDAndType(ctx, FindIntegrationByBotIDAndTypeParams{
		BotID:         sql.NullString{String: botID, Valid: true},
		ConnectorType: string(connectorType),
	})
	if err != nil {
		return infragpt.Integration{}, fmt.Errorf("failed to find integration by bot ID: %w", err)
	}

	return r.toSpecIntegration(dbIntegration)
}

func (r *integrationRepository) toSpecIntegration(dbIntegration Integration) (infragpt.Integration, error) {
	metadata := make(map[string]string)
	if dbIntegration.Metadata.Valid {
		var metadataMap map[string]any
		if err := json.Unmarshal(dbIntegration.Metadata.RawMessage, &metadataMap); err != nil {
			return infragpt.Integration{}, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		for k, v := range metadataMap {
			if str, ok := v.(string); ok {
				metadata[k] = str
			}
		}
	}

	var lastUsedAt *time.Time
	if dbIntegration.LastUsedAt.Valid {
		lastUsedAt = &dbIntegration.LastUsedAt.Time
	}

	return infragpt.Integration{
		ID:                      dbIntegration.ID.String(),
		OrganizationID:          dbIntegration.OrganizationID.String(),
		UserID:                  dbIntegration.UserID.String(),
		ConnectorType:           infragpt.ConnectorType(dbIntegration.ConnectorType),
		Status:                  infragpt.IntegrationStatus(dbIntegration.Status),
		BotID:                   dbIntegration.BotID.String,
		ConnectorUserID:         dbIntegration.ConnectorUserID.String,
		ConnectorOrganizationID: dbIntegration.ConnectorOrganizationID.String,
		Metadata:                metadata,
		CreatedAt:               dbIntegration.CreatedAt,
		UpdatedAt:               dbIntegration.UpdatedAt,
		LastUsedAt:              lastUsedAt,
	}, nil
}
