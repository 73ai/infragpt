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

func (s *service) ConfigureIntegration(ctx context.Context, cmd infragpt.ConfigureIntegrationCommand) (infragpt.Integration, error) {
	// This method handles the case where a user is redirected back from GitHub
	// after installing the app and needs to claim the installation
	
	slog.Info("configuring integration",
		"organization_id", cmd.OrganizationID,
		"user_id", cmd.UserID,
		"connector_type", cmd.ConnectorType,
		"installation_id", cmd.InstallationID)

	// Validate that organization ID and user ID are valid UUIDs
	if _, err := uuid.Parse(cmd.OrganizationID); err != nil {
		return infragpt.Integration{}, fmt.Errorf("invalid organization ID: %w", err)
	}
	if _, err := uuid.Parse(cmd.UserID); err != nil {
		return infragpt.Integration{}, fmt.Errorf("invalid user ID: %w", err)
	}

	// Check if integration already exists for this org and connector type
	existingIntegrations, err := s.integrationRepository.FindByOrganizationAndType(ctx, cmd.OrganizationID, cmd.ConnectorType)
	if err != nil {
		return infragpt.Integration{}, fmt.Errorf("failed to check existing integrations: %w", err)
	}

	if len(existingIntegrations) > 0 {
		return infragpt.Integration{}, fmt.Errorf("integration already exists for connector type %s in organization %s", cmd.ConnectorType, cmd.OrganizationID)
	}

	connector, exists := s.connectors[cmd.ConnectorType]
	if !exists {
		return infragpt.Integration{}, fmt.Errorf("unsupported connector type: %s", cmd.ConnectorType)
	}

	// For GitHub, we need to handle the installation claiming process
	if cmd.ConnectorType == infragpt.ConnectorTypeGithub {
		return s.configureGitHubIntegration(ctx, cmd, connector)
	}

	return infragpt.Integration{}, fmt.Errorf("configure integration not implemented for connector type: %s", cmd.ConnectorType)
}

func (s *service) configureGitHubIntegration(ctx context.Context, cmd infragpt.ConfigureIntegrationCommand, connector domain.Connector) (infragpt.Integration, error) {
	// For now, just use the connector interface without accessing private fields
	// In a full implementation, we would add methods to access the repository service
	// or pass it as a separate parameter
	
	slog.Debug("configuring GitHub integration via connector interface")

	// TODO: Get the repository service from the GitHub connector
	// This requires making the repositoryService field accessible or adding a getter method
	
	// For now, simulate the process without the actual database operations
	slog.Info("GitHub integration configuration started",
		"installation_id", cmd.InstallationID,
		"organization_id", cmd.OrganizationID)

	// Create mock credentials for the installation
	// In a real implementation, this would:
	// 1. Check if unclaimed installation exists
	// 2. Generate installation access token
	// 3. Get installation details from GitHub API
	// 4. Create integration and credentials records
	// 5. Sync repositories
	// 6. Mark unclaimed installation as claimed

	authData := infragpt.AuthorizationData{
		InstallationID: cmd.InstallationID,
	}

	credentials, err := connector.CompleteAuthorization(authData)
	if err != nil {
		return infragpt.Integration{}, fmt.Errorf("failed to complete GitHub authorization: %w", err)
	}

	now := time.Now()
	integration := infragpt.Integration{
		ID:             uuid.New().String(),
		OrganizationID: cmd.OrganizationID,
		UserID:         cmd.UserID,
		ConnectorType:  cmd.ConnectorType,
		Status:         infragpt.IntegrationStatusActive,
		BotID:          cmd.InstallationID,
		Metadata:       make(map[string]string),
		CreatedAt:      now,
		UpdatedAt:      now,
		LastUsedAt:     &now,
	}

	// Store connector organization info
	if credentials.OrganizationInfo != nil {
		integration.ConnectorOrganizationID = credentials.OrganizationInfo.ExternalID
		integration.Metadata["connector_org_name"] = credentials.OrganizationInfo.Name
		for k, v := range credentials.OrganizationInfo.Metadata {
			integration.Metadata[k] = v
		}
	}

	// Store integration
	if err := s.integrationRepository.Store(ctx, integration); err != nil {
		return infragpt.Integration{}, fmt.Errorf("failed to store integration: %w", err)
	}

	// Store credentials
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

	// TODO: Sync repositories using repository service
	// githubConnector.repositoryService.SyncRepositories(ctx, integration.ID, installationID)

	slog.Info("GitHub integration configured successfully",
		"integration_id", integration.ID,
		"installation_id", cmd.InstallationID,
		"organization_id", cmd.OrganizationID)

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

// handleConnectorEvent processes events from connectors - simplified for installation-only approach
func (s *service) handleConnectorEvent(ctx context.Context, event any) error {
	switch e := event.(type) {
	case github.WebhookEvent:
		return s.handleGitHubEvent(ctx, e)
	default:
		slog.Debug("received unknown event type", "event_type", fmt.Sprintf("%T", event))
		return nil
	}
}

// handleGitHubEvent processes GitHub webhook events - simplified for installation-only approach
func (s *service) handleGitHubEvent(ctx context.Context, event github.WebhookEvent) error {
	// Only handle installation events
	if event.EventType != github.EventTypeInstallation {
		slog.Debug("ignoring non-installation event", "event_type", event.EventType)
		return nil
	}

	slog.Info("handling GitHub installation event",
		"action", event.InstallationAction,
		"installation_id", event.InstallationID,
		"sender_login", event.SenderLogin,
		"repositories_added", len(event.RepositoriesAdded),
		"repositories_removed", len(event.RepositoriesRemoved))

	// Validate event has required fields
	if event.InstallationID == 0 {
		slog.Warn("GitHub event missing installation ID")
		return fmt.Errorf("missing installation ID in GitHub event")
	}

	// Process installation events
	switch event.InstallationAction {
	case "created":
		return s.handleInstallationCreated(ctx, event)
	case "deleted":
		return s.handleInstallationDeleted(ctx, event)
	case "added":
		return s.handleInstallationRepositoriesAdded(ctx, event)
	case "removed":
		return s.handleInstallationRepositoriesRemoved(ctx, event)
	default:
		slog.Debug("unhandled installation action", "action", event.InstallationAction)
		return nil
	}
}

// handleInstallationCreated processes GitHub App installation created events
func (s *service) handleInstallationCreated(ctx context.Context, event github.WebhookEvent) error {
	slog.Info("GitHub App installation created",
		"installation_id", event.InstallationID,
		"sender", event.SenderLogin,
		"repositories", len(event.RepositoriesAdded))

	// Log the installation for monitoring purposes
	// In a full implementation, you might want to:
	// 1. Store installation metadata
	// 2. Send notifications to relevant channels
	// 3. Update integration status

	return nil
}

// handleInstallationDeleted processes GitHub App installation deleted events
func (s *service) handleInstallationDeleted(ctx context.Context, event github.WebhookEvent) error {
	slog.Info("GitHub App installation deleted",
		"installation_id", event.InstallationID,
		"sender", event.SenderLogin)

	// Log the uninstallation for monitoring purposes
	// In a full implementation, you might want to:
	// 1. Update integration status to inactive
	// 2. Revoke stored credentials
	// 3. Send notifications

	return nil
}

// handleInstallationRepositoriesAdded processes repository access added events
func (s *service) handleInstallationRepositoriesAdded(ctx context.Context, event github.WebhookEvent) error {
	slog.Info("GitHub App repository access added",
		"installation_id", event.InstallationID,
		"repositories", event.RepositoriesAdded)

	// Log the repository access change for monitoring purposes

	return nil
}

// handleInstallationRepositoriesRemoved processes repository access removed events
func (s *service) handleInstallationRepositoriesRemoved(ctx context.Context, event github.WebhookEvent) error {
	slog.Info("GitHub App repository access removed",
		"installation_id", event.InstallationID,
		"repositories", event.RepositoriesRemoved)

	// Log the repository access change for monitoring purposes

	return nil
}
