package integrationsvc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/connectors/github"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/domain"
)

type service struct {
	integrationRepository domain.IntegrationRepository
	credentialRepository  domain.CredentialRepository
	connectors            map[infragpt.ConnectorType]domain.Connector
}

type ServiceConfig struct {
	IntegrationRepository domain.IntegrationRepository
	CredentialRepository  domain.CredentialRepository
	Connectors            map[infragpt.ConnectorType]domain.Connector
}

func NewService(config ServiceConfig) infragpt.IntegrationService {
	return &service{
		integrationRepository: config.IntegrationRepository,
		credentialRepository:  config.CredentialRepository,
		connectors:            config.Connectors,
	}
}

func (s *service) NewIntegration(ctx context.Context, cmd infragpt.NewIntegrationCommand) (infragpt.IntegrationAuthorizationIntent, error) {
	existingIntegrations, err := s.integrationRepository.FindByOrganizationAndType(ctx, cmd.OrganizationID, cmd.ConnectorType)
	if err != nil {
		return infragpt.IntegrationAuthorizationIntent{}, fmt.Errorf("failed to check existing integrations: %w", err)
	}

	if len(existingIntegrations) > 0 {
		return infragpt.IntegrationAuthorizationIntent{}, fmt.Errorf("integration already exists for connector type %s", cmd.ConnectorType)
	}

	connector, exists := s.connectors[cmd.ConnectorType]
	if !exists {
		return infragpt.IntegrationAuthorizationIntent{}, fmt.Errorf("unsupported connector type: %s", cmd.ConnectorType)
	}

	return connector.InitiateAuthorization(cmd.OrganizationID, cmd.UserID)
}

func (s *service) AuthorizeIntegration(ctx context.Context, cmd infragpt.AuthorizeIntegrationCommand) (infragpt.Integration, error) {
	connector, exists := s.connectors[cmd.ConnectorType]
	if !exists {
		return infragpt.Integration{}, fmt.Errorf("unsupported connector type: %s", cmd.ConnectorType)
	}

	authData := infragpt.AuthorizationData{
		Code:           cmd.Code,
		State:          cmd.State,
		InstallationID: cmd.InstallationID,
	}

	credentials, err := connector.CompleteAuthorization(authData)
	if err != nil {
		return infragpt.Integration{}, fmt.Errorf("failed to complete authorization: %w", err)
	}

	// Check if installation was already claimed by the connector
	if claimed, exists := credentials.Data["claimed"]; exists && claimed == "true" {
		// Installation was already claimed, find and return the existing integration
		organizationID, _, err := connector.ParseState(cmd.State)
		if err != nil {
			return infragpt.Integration{}, fmt.Errorf("failed to parse state: %w", err)
		}

		existingIntegrations, err := s.integrationRepository.FindByOrganizationAndType(ctx, organizationID, cmd.ConnectorType)
		if err != nil {
			return infragpt.Integration{}, fmt.Errorf("failed to find claimed integration: %w", err)
		}

		for _, integration := range existingIntegrations {
			if integration.BotID == cmd.InstallationID {
				return integration, nil
			}
		}

		return infragpt.Integration{}, fmt.Errorf("claimed integration not found")
	}

	// Parse organization ID and user ID from OAuth state using connector
	organizationID, userID, err := connector.ParseState(cmd.State)
	if err != nil {
		return infragpt.Integration{}, fmt.Errorf("failed to parse state: %w", err)
	}

	// Validate that organization ID and user ID are valid UUIDs
	if _, err := uuid.Parse(organizationID); err != nil {
		return infragpt.Integration{}, fmt.Errorf("invalid organization ID in state: %w", err)
	}
	if _, err := uuid.Parse(userID); err != nil {
		return infragpt.Integration{}, fmt.Errorf("invalid user ID in state: %w", err)
	}

	// Check if integration already exists for this org and connector type
	existingIntegrations, err := s.integrationRepository.FindByOrganizationAndType(ctx, organizationID, cmd.ConnectorType)
	if err != nil {
		return infragpt.Integration{}, fmt.Errorf("failed to check existing integrations: %w", err)
	}

	if len(existingIntegrations) > 0 {
		return infragpt.Integration{}, fmt.Errorf("integration already exists for connector type %s in organization %s", cmd.ConnectorType, organizationID)
	}

	now := time.Now()
	integration := infragpt.Integration{
		ID:             uuid.New().String(),
		OrganizationID: organizationID,
		UserID:         userID,
		ConnectorType:  cmd.ConnectorType,
		Status:         infragpt.IntegrationStatusActive,
		Metadata:       make(map[string]string),
		CreatedAt:      now,
		UpdatedAt:      now,
		LastUsedAt:     &now,
	}

	if cmd.InstallationID != "" {
		integration.BotID = cmd.InstallationID
	}

	// Store connector organization info in integration metadata
	if credentials.OrganizationInfo != nil {
		integration.ConnectorOrganizationID = credentials.OrganizationInfo.ExternalID
		integration.Metadata["connector_org_name"] = credentials.OrganizationInfo.Name
		for k, v := range credentials.OrganizationInfo.Metadata {
			integration.Metadata[k] = v
		}
	}

	if err := s.integrationRepository.Store(ctx, integration); err != nil {
		return infragpt.Integration{}, fmt.Errorf("failed to store integration: %w", err)
	}

	credentialRecord := domain.IntegrationCredential{
		ID:              uuid.New().String(),
		IntegrationID:   integration.ID,
		CredentialType:  credentials.Type,
		Data:            credentials.Data,
		ExpiresAt:       credentials.ExpiresAt,
		EncryptionKeyID: "v1",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.credentialRepository.Store(ctx, credentialRecord); err != nil {
		return infragpt.Integration{}, fmt.Errorf("failed to store credentials: %w", err)
	}

	return integration, nil
}

func (s *service) RevokeIntegration(ctx context.Context, cmd infragpt.RevokeIntegrationCommand) error {
	integration, err := s.integrationRepository.FindByID(ctx, cmd.IntegrationID)
	if err != nil {
		return fmt.Errorf("failed to find integration: %w", err)
	}

	if integration.OrganizationID != cmd.OrganizationID {
		return fmt.Errorf("integration not found for organization")
	}

	credential, err := s.credentialRepository.FindByIntegration(ctx, cmd.IntegrationID)
	if err != nil {
		return fmt.Errorf("failed to find credentials: %w", err)
	}

	connector, exists := s.connectors[integration.ConnectorType]
	if exists {
		creds := infragpt.Credentials{
			Type:      credential.CredentialType,
			Data:      credential.Data,
			ExpiresAt: credential.ExpiresAt,
		}

		if err := connector.RevokeCredentials(creds); err != nil {
			return fmt.Errorf("failed to revoke credentials with connector: %w", err)
		}
	}

	if err := s.credentialRepository.Delete(ctx, cmd.IntegrationID); err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	if err := s.integrationRepository.Delete(ctx, cmd.IntegrationID); err != nil {
		return fmt.Errorf("failed to delete integration: %w", err)
	}

	return nil
}

func (s *service) Integrations(ctx context.Context, query infragpt.IntegrationsQuery) ([]infragpt.Integration, error) {
	if query.ConnectorType != "" {
		return s.integrationRepository.FindByOrganizationAndType(ctx, query.OrganizationID, query.ConnectorType)
	}

	return s.integrationRepository.FindByOrganization(ctx, query.OrganizationID)
}

func (s *service) Integration(ctx context.Context, query infragpt.IntegrationQuery) (infragpt.Integration, error) {
	integration, err := s.integrationRepository.FindByID(ctx, query.IntegrationID)
	if err != nil {
		return infragpt.Integration{}, fmt.Errorf("failed to find integration: %w", err)
	}

	if integration.OrganizationID != query.OrganizationID {
		return infragpt.Integration{}, fmt.Errorf("integration not found for organization")
	}

	return integration, nil
}

func (s *service) RefreshIntegration(ctx context.Context, cmd infragpt.RefreshIntegrationCommand) error {
	integration, err := s.integrationRepository.FindByID(ctx, cmd.IntegrationID)
	if err != nil {
		return fmt.Errorf("failed to find integration: %w", err)
	}

	if integration.OrganizationID != cmd.OrganizationID {
		return fmt.Errorf("integration not found for organization")
	}

	credential, err := s.credentialRepository.FindByIntegration(ctx, cmd.IntegrationID)
	if err != nil {
		return fmt.Errorf("failed to find credentials: %w", err)
	}

	connector, exists := s.connectors[integration.ConnectorType]
	if !exists {
		return fmt.Errorf("unsupported connector type: %s", integration.ConnectorType)
	}

	currentCreds := infragpt.Credentials{
		Type:      credential.CredentialType,
		Data:      credential.Data,
		ExpiresAt: credential.ExpiresAt,
	}

	refreshedCreds, err := connector.RefreshCredentials(currentCreds)
	if err != nil {
		return fmt.Errorf("failed to refresh credentials: %w", err)
	}

	now := time.Now()
	updatedCredential := domain.IntegrationCredential{
		ID:              credential.ID,
		IntegrationID:   credential.IntegrationID,
		CredentialType:  refreshedCreds.Type,
		Data:            refreshedCreds.Data,
		ExpiresAt:       refreshedCreds.ExpiresAt,
		EncryptionKeyID: credential.EncryptionKeyID,
		CreatedAt:       credential.CreatedAt,
		UpdatedAt:       now,
	}

	if err := s.credentialRepository.Update(ctx, updatedCredential); err != nil {
		return fmt.Errorf("failed to update credentials: %w", err)
	}

	return nil
}

func (s *service) SyncIntegration(ctx context.Context, cmd infragpt.SyncIntegrationCommand) error {
	integration, err := s.integrationRepository.FindByID(ctx, cmd.IntegrationID)
	if err != nil {
		return fmt.Errorf("failed to find integration: %w", err)
	}

	if integration.OrganizationID != cmd.OrganizationID {
		return fmt.Errorf("integration not found for organization")
	}

	connector, exists := s.connectors[integration.ConnectorType]
	if !exists {
		return fmt.Errorf("unsupported connector type: %s", integration.ConnectorType)
	}

	if err := connector.Sync(ctx, integration, cmd.Parameters); err != nil {
		return fmt.Errorf("failed to sync integration: %w", err)
	}

	// Update last used timestamp
	now := time.Now()
	integration.LastUsedAt = &now
	integration.UpdatedAt = now

	if err := s.integrationRepository.Update(ctx, integration); err != nil {
		return fmt.Errorf("failed to update integration: %w", err)
	}

	return nil
}

// Subscribe starts webhook subscriptions for all connectors
func (s *service) Subscribe(ctx context.Context) error {
	for connectorType, connector := range s.connectors {
		go func(connectorType infragpt.ConnectorType, connector domain.Connector) {
			if err := connector.Subscribe(ctx, s.handleConnectorEvent); err != nil {
				slog.Error("connector subscription failed", "connector_type", connectorType, "error", err)
			}
		}(connectorType, connector)
	}

	return nil
}

// handleConnectorEvent processes events from connectors - delegate to connectors for processing
func (s *service) handleConnectorEvent(ctx context.Context, event any) error {
	// Simple delegation - let each connector handle its own events
	switch e := event.(type) {
	case github.WebhookEvent:
		// Get GitHub connector and delegate event processing
		if connector, exists := s.connectors[infragpt.ConnectorTypeGithub]; exists {
			return connector.ProcessEvent(ctx, e)
		}
		return fmt.Errorf("GitHub connector not found")
	default:
		slog.Debug("received unknown event type", "event_type", fmt.Sprintf("%T", event))
		return nil
	}
}
