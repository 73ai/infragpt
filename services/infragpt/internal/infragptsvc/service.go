package infragptsvc

import (
	"context"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc/domain"
	"log/slog"
)

type Service struct {
	slackGateway          domain.SlackGateway
	integrationRepository domain.IntegrationRepository
}

func (s *Service) Integrations(ctx context.Context, query infragpt.IntegrationsQuery) ([]infragpt.Integration, error) {
	is, err := s.integrationRepository.Integrations(ctx, query.BusinessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get integrations: %w", err)
	}

	var integrations []infragpt.Integration
	for _, i := range is {
		integrations = append(integrations, infragpt.Integration{
			Type:   i.Type,
			Status: i.Status,
		})
	}

	return integrations, nil
}

func (s *Service) CompleteSlackAuthentication(ctx context.Context, command infragpt.CompleteSlackAuthenticationCommand) error {
	if pid, err := s.slackGateway.CompleteAuthentication(ctx, command.Code); err != nil {
		return fmt.Errorf("failed to complete slack authentication: %w", err)
	} else {
		err := s.integrationRepository.SaveIntegration(ctx, domain.Integration{
			Integration: infragpt.Integration{
				Type:   infragpt.IntegrationTypeSlack,
				Status: infragpt.IntegrationStatusActive,
			},
			BusinessID:        command.BusinessID,
			ProviderProjectID: pid,
		})
		if err != nil {
			return fmt.Errorf("failed to complete slack authentication: %w", err)
		}
	}

	return nil
}

var _ infragpt.Service = (*Service)(nil)

func (s *Service) SubscribeSlackNotifications(ctx context.Context) error {
	if err := s.slackGateway.SubscribeAppMentioned(ctx, s.handleUserCommand); err != nil {
		return fmt.Errorf("failed to subscribe app mentioned events: %w", err)
	}

	return nil
}

func (s *Service) handleUserCommand(ctx context.Context, command domain.UserCommand) error {
	slog.Info("Received user command", "command", command)
	return nil
}
