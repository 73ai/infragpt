package integrationsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/domain"
)

type service struct {
	integrationRepository domain.IntegrationRepository
	credentialRepository  domain.CredentialRepository
	connectors            map[infragpt.ConnectorType]domain.Connector
}

type Config struct {
	IntegrationRepository domain.IntegrationRepository
	CredentialRepository  domain.CredentialRepository
	Connectors            map[infragpt.ConnectorType]domain.Connector
}

func NewService(config Config) infragpt.IntegrationService {
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
		ConnectorType:  cmd.ConnectorType,
		Code:           cmd.Code,
		State:          cmd.State,
		InstallationID: cmd.InstallationID,
	}

	credentials, err := connector.CompleteAuthorization(authData)
	if err != nil {
		return infragpt.Integration{}, fmt.Errorf("failed to complete authorization: %w", err)
	}

	now := time.Now()
	integration := infragpt.Integration{
		ID:             uuid.New().String(),
		OrganizationID: cmd.OrganizationID,
		UserID:         "user-from-auth", // TODO: Get from auth context
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
