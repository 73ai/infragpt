package slack

import (
	"context"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc/domain"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"log/slog"
	"strings"
)

func (s *Slack) handleAppMention(ctx context.Context, teamID string, event *slackevents.AppMentionEvent, handler func(context.Context, domain.UserCommand) error) error {
	teamToken, err := s.tokenRepository.GetToken(ctx, teamID)
	if err != nil {
		return fmt.Errorf("error getting team token for tead_id:%s err:%w", teamID, err)
	}

	teamClient := slack.New(teamToken)
	at, err := teamClient.AuthTest()
	if err != nil {
		return fmt.Errorf("error authenticating team: %w", err)
	}

	botUserID := at.UserID

	// Extract text without the bot mention
	text := strings.TrimSpace(strings.Replace(event.Text, fmt.Sprintf("<@%s>", botUserID), "", -1))

	// Acknowledge the request in the same thread
	_, _, err = teamClient.PostMessage(
		event.Channel,
		slack.MsgOptionText(fmt.Sprintf("I'm analyzing your request: *%s*", text), false),
		slack.MsgOptionTS(event.TimeStamp), // This keeps replies in the thread
	)
	if err != nil {
		slog.Error("handleAppMention: error sending acknowledgment message", "error", err)
		return err
	}

	// Get requester info
	requesterInfo, err := teamClient.GetUserInfo(event.User)
	requesterName := ""
	requesterUsername := ""
	requesterEmail := ""
	if err == nil && requesterInfo != nil {
		requesterName = requesterInfo.RealName
		requesterUsername = requesterInfo.Name // This is the @username
		requesterEmail = requesterInfo.Profile.Email
	} else {
		slog.Error("Error getting requester info:", "err", err)
	}

	var inReply bool
	var threadTimeStamp string
	// check if it is new thread or existing thread
	if event.ThreadTimeStamp == "" {
		inReply = false
		threadTimeStamp = event.TimeStamp
	} else {
		inReply = true
		threadTimeStamp = event.ThreadTimeStamp
	}

	m := domain.SlackThread{
		TeamID:   teamID,
		Channel:  event.Channel,
		ThreadTS: threadTimeStamp,
		Sender: domain.SlackUser{
			ID:       event.User,
			Email:    requesterEmail,
			Name:     requesterName,
			Username: requesterUsername,
		},
		Message: text,
	}

	command := domain.UserCommand{
		Thread:  m,
		InReply: inReply,
	}

	return handler(ctx, command)
}
