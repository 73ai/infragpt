package infragptsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc/domain"
	"log/slog"
	"time"
)

type Service struct {
	slackGateway           domain.SlackGateway
	integrationRepository  domain.IntegrationRepository
	conversationRepository domain.ConversationRepository
	channelRepository      domain.ChannelRepository
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

func (s *Service) CompleteSlackIntegration(ctx context.Context, command infragpt.CompleteSlackIntegrationCommand) error {
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
	if err := s.slackGateway.SubscribeAllMessages(ctx, s.handleUserCommand); err != nil {
		return fmt.Errorf("failed to subscribe to all messages: %w", err)
	}

	return nil
}

func (s *Service) handleUserCommand(ctx context.Context, command domain.UserCommand) error {
	slog.Info("Received user command", "type", command.MessageType, "channel", command.Thread.Channel, "user", command.Thread.Sender.Username)

	conversation, err := s.conversationRepository.GetConversationByThread(ctx, command.Thread.TeamID, command.Thread.Channel, command.Thread.ThreadTS)
	if err != nil {
		slog.Error("Failed to get conversation", "error", err)
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	if conversation == nil {
		conversation, err = s.conversationRepository.CreateConversation(ctx, command.Thread.TeamID, command.Thread.Channel, command.Thread.ThreadTS)
		if err != nil {
			slog.Error("Failed to create conversation", "error", err)
			return fmt.Errorf("failed to create conversation: %w", err)
		}
	}

	message := domain.Message{
		ConversationID: conversation.ID,
		SlackMessageTS: fmt.Sprintf("%d", time.Now().UnixNano()),
		Sender:         command.Thread.Sender,
		MessageText:    command.Thread.Message,
		IsBotMessage:   false,
	}

	// check if the message is already stored
	_, err = s.conversationRepository.MessageBySlackTS(ctx, conversation.ID, command.Thread.Sender.ID, command.MessageTS)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("No existing message found, proceeding to store new message")
		} else {
			slog.Error("Failed to get messages by Slack timestamp", "error", err)
			return fmt.Errorf("failed to get messages by Slack timestamp: %w", err)
		}
	}

	_, err = s.conversationRepository.StoreMessage(ctx, conversation.ID, message)
	if err != nil {
		slog.Error("Failed to store message", "error", err)
		return fmt.Errorf("failed to store message: %w", err)
	}

	conversationHistory, err := s.conversationRepository.GetConversationHistory(ctx, conversation.ID, 50)
	if err != nil {
		slog.Error("Failed to get conversation history", "error", err)
	} else {
		slog.Info("Retrieved conversation history", "messageCount", len(conversationHistory))
	}

	err = s.slackGateway.ReplyMessage(ctx, command.Thread, "I received your message and stored it in the conversation history!")
	if err != nil {
		slog.Error("Failed to reply to message", "error", err)
		return fmt.Errorf("failed to reply to message: %w", err)
	}

	botMessage := domain.Message{
		ConversationID: conversation.ID,
		SlackMessageTS: fmt.Sprintf("%d", time.Now().UnixNano()),
		Sender: domain.SlackUser{
			ID:       "bot",
			Username: "bot",
			Name:     "InfraGPT Bot",
		},
		MessageText:  "I received your message and stored it in the conversation history!",
		IsBotMessage: true,
	}

	_, err = s.conversationRepository.StoreMessage(ctx, conversation.ID, botMessage)
	if err != nil {
		slog.Error("Failed to store bot response", "error", err)
	}

	return nil
}
