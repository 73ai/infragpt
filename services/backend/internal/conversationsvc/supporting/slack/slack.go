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
	// Compile regexes once outside the loop for performance
	var (
		headerRegex    = regexp.MustCompile(`^#{1,6}\s+(.+)$`)
		bulletWithBold = regexp.MustCompile(`^(\d+)\.\s+\*\*([^*]+)\*\*:?\s*(.*)$`)
		numberedList   = regexp.MustCompile(`^(\d+)\.\s+(.+)$`)
		boldRegex      = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	)

	lines := strings.Split(markdown, "\n")
	var result []string
	inCodeFence := false

	for _, line := range lines {
		// Preserve indentation; trim only trailing whitespace
		trimmedLeft := strings.TrimLeft(line, " \t")
		line = strings.TrimRight(line, " \t")

		// Toggle fence state and preserve fence lines as-is
		if strings.HasPrefix(trimmedLeft, "```") {
			inCodeFence = !inCodeFence
			result = append(result, line)
			continue
		}

		// Skip transformations inside code fences
		if inCodeFence {
			result = append(result, line)
			continue
		}

		// Handle empty lines
		if trimmedLeft == "" {
			result = append(result, "")
			continue
		}

		// Convert headers to bold text (Slack doesn't support headers natively)
		if matches := headerRegex.FindStringSubmatch(trimmedLeft); matches != nil {
			content := matches[1]
			result = append(result, "*"+content+"*")
			continue
		}

		// Convert bullet points with ** formatting (must come before generic numbered lists)
		if matches := bulletWithBold.FindStringSubmatch(trimmedLeft); matches != nil {
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

		// Convert numbered lists (generic case)
		if matches := numberedList.FindStringSubmatch(trimmedLeft); matches != nil {
			formatted := matches[1] + ". " + matches[2]
			result = append(result, formatted)
			continue
		}

		// Convert **bold** to *bold* only outside inline code spans
		line = transformBoldOutsideCode(line, boldRegex)

		result = append(result, line)
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}

// transformBoldOutsideCode applies bold transformation only outside inline code spans
func transformBoldOutsideCode(line string, boldRegex *regexp.Regexp) string {
	parts := strings.Split(line, "`")
	for i := 0; i < len(parts); i++ {
		if i%2 == 0 { // outside code spans (even indices)
			parts[i] = boldRegex.ReplaceAllString(parts[i], "*$1*")
		}
	}
	return strings.Join(parts, "`")
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
