package slack

import (
	"context"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc/domain"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/sync/errgroup"
	"net/http"
)

type Slack struct {
	clientID        string
	clientSecret    string
	client          *slack.Client
	socketClient    *socketmode.Client
	tokenRepository domain.WorkSpaceTokenRepository
}

// TODO: Advanced token security via token rotation
func (s *Slack) CompleteAuthentication(ctx context.Context, code string) (string, error) {
	// exchange the code for an access token
	oauthV2Response, err := slack.GetOAuthV2Response(
		http.DefaultClient,
		s.clientID,
		s.clientSecret,
		code,
		"",
	)
	if err != nil {
		return "", fmt.Errorf("failed to get oauth v2 response: %w", err)
	}

	// save the access token
	if err := s.tokenRepository.SaveToken(ctx, oauthV2Response.Team.ID, oauthV2Response.AccessToken); err != nil {
		return "", fmt.Errorf("failed to save token: %w", err)
	}

	return oauthV2Response.Team.ID, nil
}

func (s *Slack) SubscribeAppMentioned(ctx context.Context, f func(ctx context.Context, command domain.UserCommand) error) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := s.subscribe(ctx, f); err != nil {
			return fmt.Errorf("failed to subscribe to app mentioned user: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		if err := s.socketClient.Run(); err != nil {
			return fmt.Errorf("failed to run socket client: %w", err)
		}
		return nil
	})

	return g.Wait()
}

func (s *Slack) ReplyMessage(ctx context.Context, t domain.SlackThread, message string) error {
	teamToken, err := s.tokenRepository.GetToken(ctx, t.TeamID)
	if err != nil {
		return fmt.Errorf("failed to get team token: %w", err)
	}
	teamClient := slack.New(teamToken)
	_, _, err = teamClient.PostMessage(
		t.Channel,
		slack.MsgOptionText(message, false),
		slack.MsgOptionTS(t.ThreadTS), // This keeps replies in the thread
	)
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	return nil
}

var _ domain.SlackGateway = (*Slack)(nil)
