package github

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/domain"
)

type Config struct {
	AppID         string `mapstructure:"app_id"`
	PrivateKey    string `mapstructure:"private_key"`
	WebhookSecret string `mapstructure:"webhook_secret"`
	RedirectURL   string `mapstructure:"redirect_url"`
	WebhookPort   int    `mapstructure:"webhook_port"`
}

func (c Config) NewConnector() domain.Connector {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(c.PrivateKey))
	if err != nil {
		// Return a connector with nil private key that will fail during JWT generation
		// This allows the error to be handled at runtime rather than during initialization
		return &githubConnector{
			config:     c,
			client:     &http.Client{Timeout: 30 * time.Second},
			privateKey: nil,
		}
	}

	return &githubConnector{
		config:     c,
		client:     &http.Client{Timeout: 30 * time.Second},
		privateKey: privateKey,
	}
}
