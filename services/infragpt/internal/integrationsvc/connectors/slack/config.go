package slack

import (
	"net/http"
	"time"

	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/domain"
)

type Config struct {
	ClientID      string   `mapstructure:"client_id"`
	ClientSecret  string   `mapstructure:"client_secret"`
	RedirectURL   string   `mapstructure:"redirect_url"`
	Scopes        []string `mapstructure:"scopes"`
	SigningSecret string   `mapstructure:"signing_secret"`
	BotToken      string   `mapstructure:"bot_token"`
	AppToken      string   `mapstructure:"app_token"`
}

func (c Config) NewConnector() domain.Connector {
	// Ensure default scopes are set if none are provided
	if len(c.Scopes) == 0 {
		c.Scopes = []string{
			"app_mentions:read",
			"chat:write",
			"im:history",
			"im:write",
			"reactions:write",
			"users:read",
			"users:read.email",
			"channels:read",
			"channels:history",
			"groups:read",
			"groups:history",
		}
	}
	
	return &slackConnector{
		config: c,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}