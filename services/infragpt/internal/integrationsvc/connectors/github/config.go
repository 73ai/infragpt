package github

import (
	"net/http"
	"time"

	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/domain"
)

type Config struct {
	AppID         string `mapstructure:"app_id"`
	PrivateKey    string `mapstructure:"private_key"`
	WebhookSecret string `mapstructure:"webhook_secret"`
	RedirectURL   string `mapstructure:"redirect_url"`
}

func (c Config) NewConnector() domain.Connector {
	return &githubConnector{
		config: c,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}