package integrationsvc

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/connectors/github"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/connectors/slack"
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

func (s *service) Subscribe(ctx context.Context) error {
	var wg sync.WaitGroup

	for connectorType, connector := range s.connectors {
		wg.Add(1)
		go func(connectorType infragpt.ConnectorType, connector domain.Connector) {
			defer wg.Done()

			if err := connector.Subscribe(ctx, s.handleConnectorEvent); err != nil {
				slog.Error("connector subscription failed", "connector_type", connectorType, "error", err)
			}
		}(connectorType, connector)
	}

	wg.Wait()
	return nil
}

func (s *service) handleConnectorEvent(ctx context.Context, event any) error {
	switch e := event.(type) {
	case slack.MessageEvent:
		return s.handleSlackEvent(ctx, e)
	case github.WebhookEvent:
		return s.handleGitHubEvent(ctx, e)
	default:
		slog.Warn("received unknown event type", "event_type", fmt.Sprintf("%T", event))
		return nil
	}
}

func (s *service) handleSlackEvent(ctx context.Context, event slack.MessageEvent) error {
	slog.Info("handling Slack event",
		"event_type", event.EventType,
		"team_id", event.TeamID,
		"channel_id", event.ChannelID,
		"user_id", event.UserID)

	switch event.EventType {
	case slack.EventTypeMessage:
		return s.handleSlackMessage(ctx, event)
	case slack.EventTypeSlashCommand:
		return s.handleSlackSlashCommand(ctx, event)
	case slack.EventTypeAppMention:
		return s.handleSlackAppMention(ctx, event)
	default:
		slog.Debug("unhandled Slack event type", "event_type", event.EventType)
		return nil
	}
}

func (s *service) handleGitHubEvent(ctx context.Context, event github.WebhookEvent) error {
	slog.Info("handling GitHub event",
		"event_type", event.EventType,
		"installation_id", event.InstallationID,
		"repository_name", event.RepositoryName,
		"sender_login", event.SenderLogin)

	switch event.EventType {
	case github.EventTypePush:
		return s.handleGitHubPush(ctx, event)
	case github.EventTypePullRequest:
		return s.handleGitHubPullRequest(ctx, event)
	case github.EventTypeInstallation:
		return s.handleGitHubInstallation(ctx, event)
	default:
		slog.Debug("unhandled GitHub event type", "event_type", event.EventType)
		return nil
	}
}

func (s *service) handleSlackMessage(ctx context.Context, event slack.MessageEvent) error {
	// TODO: Implement Slack message processing logic
	// This could involve parsing commands, triggering workflows, etc.
	slog.Info("processing Slack message", "text", event.Text, "channel", event.ChannelID)
	return nil
}

func (s *service) handleSlackSlashCommand(ctx context.Context, event slack.MessageEvent) error {
	// TODO: Implement Slack slash command processing logic
	slog.Info("processing Slack slash command", "command", event.Command, "text", event.Text)
	return nil
}

func (s *service) handleSlackAppMention(ctx context.Context, event slack.MessageEvent) error {
	// TODO: Implement Slack app mention processing logic
	slog.Info("processing Slack app mention", "text", event.Text, "channel", event.ChannelID)
	return nil
}

func (s *service) handleGitHubPush(ctx context.Context, event github.WebhookEvent) error {
	// TODO: Implement GitHub push event processing logic
	slog.Info("processing GitHub push", "repository", event.RepositoryName, "branch", event.Branch, "commit", event.CommitSHA)
	return nil
}

func (s *service) handleGitHubPullRequest(ctx context.Context, event github.WebhookEvent) error {
	// TODO: Implement GitHub pull request processing logic
	slog.Info("processing GitHub pull request",
		"repository", event.RepositoryName,
		"pr_number", event.PullRequestNumber,
		"action", event.Action,
		"title", event.PullRequestTitle)
	return nil
}

func (s *service) handleGitHubInstallation(ctx context.Context, event github.WebhookEvent) error {
	// TODO: Implement GitHub installation event processing logic
	slog.Info("processing GitHub installation",
		"installation_id", event.InstallationID,
		"action", event.InstallationAction,
		"repositories_added", event.RepositoriesAdded,
		"repositories_removed", event.RepositoriesRemoved)
	return nil
}
