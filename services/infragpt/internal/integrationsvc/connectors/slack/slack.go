package slack

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/priyanshujain/infragpt/services/infragpt"
)

type slackConnector struct {
	config Config
	client *http.Client
}


func (s *slackConnector) InitiateAuthorization(organizationID string, userID string) (infragpt.IntegrationAuthorizationIntent, error) {
	state := fmt.Sprintf("%s:%s:%d", organizationID, userID, time.Now().Unix())
	
	params := url.Values{}
	params.Set("client_id", s.config.ClientID)
	params.Set("scope", strings.Join(s.config.Scopes, ","))
	params.Set("redirect_uri", s.config.RedirectURL)
	params.Set("state", state)
	params.Set("user_scope", "")

	authURL := fmt.Sprintf("https://slack.com/oauth/v2/authorize?%s", params.Encode())

	return infragpt.IntegrationAuthorizationIntent{
		Type: infragpt.AuthorizationTypeOAuth2,
		URL:  authURL,
	}, nil
}

func (s *slackConnector) CompleteAuthorization(authData infragpt.AuthorizationData) (infragpt.Credentials, error) {
	if authData.Code == "" {
		return infragpt.Credentials{}, fmt.Errorf("authorization code is required")
	}

	params := url.Values{}
	params.Set("client_id", s.config.ClientID)
	params.Set("client_secret", s.config.ClientSecret)
	params.Set("code", authData.Code)
	params.Set("redirect_uri", s.config.RedirectURL)

	resp, err := s.client.PostForm("https://slack.com/api/oauth.v2.access", params)
	if err != nil {
		return infragpt.Credentials{}, fmt.Errorf("failed to exchange code for token: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		OK          bool   `json:"ok"`
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		UserID      string `json:"user_id"`
		TeamID      string `json:"team_id"`
		TeamName    string `json:"team_name"`
		Bot         struct {
			BotUserID      string `json:"bot_user_id"`
			BotAccessToken string `json:"bot_access_token"`
		} `json:"bot"`
		Error string `json:"error,omitempty"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return infragpt.Credentials{}, fmt.Errorf("failed to decode OAuth response: %w", err)
	}

	if !response.OK {
		return infragpt.Credentials{}, fmt.Errorf("OAuth error: %s", response.Error)
	}

	credentialData := map[string]string{
		"access_token":     response.AccessToken,
		"bot_access_token": response.Bot.BotAccessToken,
		"bot_user_id":      response.Bot.BotUserID,
		"team_id":          response.TeamID,
		"team_name":        response.TeamName,
		"user_id":          response.UserID,
		"scope":            response.Scope,
	}

	return infragpt.Credentials{
		Type: infragpt.CredentialTypeOAuth2,
		Data: credentialData,
	}, nil
}

func (s *slackConnector) ValidateCredentials(creds infragpt.Credentials) error {
	botToken, exists := creds.Data["bot_access_token"]
	if !exists {
		return fmt.Errorf("bot access token not found in credentials")
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/auth.test", nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", botToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate credentials: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		OK     bool   `json:"ok"`
		URL    string `json:"url"`
		Team   string `json:"team"`
		User   string `json:"user"`
		TeamID string `json:"team_id"`
		UserID string `json:"user_id"`
		BotID  string `json:"bot_id"`
		Error  string `json:"error,omitempty"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode auth test response: %w", err)
	}

	if !response.OK {
		return fmt.Errorf("credential validation failed: %s", response.Error)
	}

	return nil
}

func (s *slackConnector) RefreshCredentials(creds infragpt.Credentials) (infragpt.Credentials, error) {
	return creds, fmt.Errorf("Slack OAuth2 tokens do not support refresh")
}

func (s *slackConnector) RevokeCredentials(creds infragpt.Credentials) error {
	accessToken, exists := creds.Data["access_token"]
	if !exists {
		return fmt.Errorf("access token not found in credentials")
	}

	params := url.Values{}
	params.Set("token", accessToken)

	resp, err := s.client.PostForm("https://slack.com/api/auth.revoke", params)
	if err != nil {
		return fmt.Errorf("failed to revoke credentials: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		OK       bool   `json:"ok"`
		Revoked  bool   `json:"revoked"`
		Error    string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode revoke response: %w", err)
	}

	if !response.OK {
		return fmt.Errorf("failed to revoke credentials: %s", response.Error)
	}

	return nil
}

func (s *slackConnector) ConfigureWebhooks(integrationID string, creds infragpt.Credentials) error {
	return nil
}

func (s *slackConnector) ValidateWebhookSignature(payload []byte, signature string, secret string) error {
	if secret == "" {
		secret = s.config.SigningSecret
	}

	expectedSignature := s.computeSignature(payload, secret)
	
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("webhook signature validation failed")
	}

	return nil
}

func (s *slackConnector) computeSignature(payload []byte, secret string) string {
	timestamp := time.Now().Unix()
	baseString := fmt.Sprintf("v0:%d:%s", timestamp, string(payload))
	
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(baseString))
	
	return fmt.Sprintf("v0=%s", hex.EncodeToString(h.Sum(nil)))
}

func (s *slackConnector) Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error {
	if s.config.BotToken == "" {
		return fmt.Errorf("slack: bot token is required for Socket Mode")
	}
	if s.config.AppToken == "" {
		return fmt.Errorf("slack: app token is required for Socket Mode")
	}

	// TODO: Implement Socket Mode client when Slack library is available
	// For now, return a placeholder implementation
	// 
	// Example implementation would be:
	// client := socketmode.New(
	//     slack.New(s.config.BotToken),
	//     socketmode.OptionAppToken(s.config.AppToken),
	// )
	//
	// go func() {
	//     for evt := range client.Events {
	//         switch evt.Type {
	//         case socketmode.EventTypeEventsAPI:
	//             eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	//             if !ok {
	//                 continue
	//             }
	//             
	//             messageEvent := s.convertToMessageEvent(eventsAPIEvent)
	//             if err := handler(ctx, messageEvent); err != nil {
	//                 // Log error but continue processing
	//             }
	//         }
	//     }
	// }()
	//
	// return client.Run()
	
	return fmt.Errorf("slack Socket Mode implementation pending - requires slack-go library")
}

func (s *slackConnector) convertToMessageEvent(rawEvent interface{}) MessageEvent {
	// TODO: Convert Slack Socket Mode events to our MessageEvent format
	// This would parse different event types (message, slash command, etc.)
	// and create appropriate MessageEvent structs
	
	return MessageEvent{
		EventType: EventTypeMessage,
		TeamID:    "",
		ChannelID: "",
		UserID:    "",
		Text:      "",
		Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
		CreatedAt: time.Now(),
		RawEvent:  make(map[string]interface{}),
	}
}