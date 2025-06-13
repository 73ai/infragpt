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
	return &slackConnector{
		config: c,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}