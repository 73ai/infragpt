package slack

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/priyanshujain/infragpt/services/backend/internal/conversationsvc/domain"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/sync/errgroup"
)

// transformMarkdownToSlack converts standard markdown to Slack's mrkdwn format
func transformMarkdownToSlack(markdown string) string {
	lines := strings.Split(markdown, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			result = append(result, "")
			continue
		}

		// Convert headers to bold text (Slack doesn't support headers natively)
		if matched, _ := regexp.MatchString(`^#{1,6}\s+`, line); matched {
			headerRegex := regexp.MustCompile(`^#{1,6}\s+(.+)$`)
			content := headerRegex.ReplaceAllString(line, "$1")
			result = append(result, "*"+content+"*")
			continue
		}

		// Convert numbered lists
		if matched, _ := regexp.MatchString(`^(\d+)\.\s+`, line); matched {
			numberedListRegex := regexp.MustCompile(`^(\d+)\.\s+(.+)$`)
			formatted := numberedListRegex.ReplaceAllString(line, "$1. $2")
			result = append(result, formatted)
			continue
		}

		// Convert bullet points with ** formatting
		if matched, _ := regexp.MatchString(`^\d+\.\s+\*\*`, line); matched {
			bulletRegex := regexp.MustCompile(`^(\d+)\.\s+\*\*([^*]+)\*\*:?\s*(.*)$`)
			matches := bulletRegex.FindStringSubmatch(line)
			if len(matches) >= 4 {
				num := matches[1]
				title := matches[2]
				content := matches[3]
				if content != "" {
					result = append(result, num+". *"+title+"*: "+content)
				} else {
					result = append(result, num+". *"+title+"*")
				}
				continue
			}
		}

		// Convert remaining ** bold ** to * bold *
		line = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(line, "*$1*")

		// Handle inline code (preserve as is)
		line = regexp.MustCompile("`([^`]*)`").ReplaceAllString(line, "`$1`")

		result = append(result, line)
	}

	// Handle code blocks (preserve them as is) - process after line-by-line
	finalText := strings.Join(result, "\n")
	finalText = regexp.MustCompile("```([^`]*)```").ReplaceAllString(finalText, "```$1```")

	return strings.TrimSpace(finalText)
}

type Slack struct {
	clientID          string
	clientSecret      string
	client            *slack.Client
	socketClient      *socketmode.Client
	tokenRepository   domain.WorkSpaceTokenRepository
	channelRepository domain.ChannelRepository
}

// TODO: Advanced token security via token rotation
func (s *Slack) CompleteAuthentication(ctx context.Context, code string) (string, error) {
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

	// TODO: store refresh token and handle token refresh
	if err := s.tokenRepository.SaveToken(ctx, oauthV2Response.Team.ID, oauthV2Response.AccessToken); err != nil {
		return "", fmt.Errorf("failed to save token: %w", err)
	}

	return oauthV2Response.Team.ID, nil
}

func (s *Slack) SubscribeAllMessages(ctx context.Context, f func(ctx context.Context, command domain.UserCommand) error) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := s.subscribe(ctx, f); err != nil {
			return fmt.Errorf("failed to subscribe to all messages: %w", err)
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

	// Transform markdown to Slack format
	slackFormattedMessage := transformMarkdownToSlack(message)

	_, _, err = teamClient.PostMessage(
		t.Channel,
		slack.MsgOptionText(slackFormattedMessage, false),
		slack.MsgOptionTS(t.ThreadTS),
	)
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	return nil
}

var _ domain.SlackGateway = (*Slack)(nil)
