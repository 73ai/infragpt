package github

import (
	"crypto/rsa"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/domain"
)

type Config struct {
	AppID         string `mapstructure:"app_id"`
	AppName       string `mapstructure:"app_name"`
	PrivateKey    string `mapstructure:"private_key"`
	WebhookSecret string `mapstructure:"webhook_secret"`
	RedirectURL   string `mapstructure:"redirect_url"`
	WebhookPort   int    `mapstructure:"webhook_port"`

	// Repository dependencies
	UnclaimedInstallationRepo UnclaimedInstallationRepository
	GitHubRepositoryRepo      GitHubRepositoryRepository
	IntegrationRepository     domain.IntegrationRepository
	CredentialRepository      domain.CredentialRepository
}

func (c Config) NewConnector() domain.Connector {
	var privateKey *rsa.PrivateKey

	if c.PrivateKey != "" {
		var err error
		privateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(c.PrivateKey))
		if err != nil {
			// Log the specific error for debugging
			slog.Error("Failed to parse GitHub private key", "error", err)
			privateKey = nil
		}
	}

	connector := &githubConnector{
		config:     c,
		client:     &http.Client{Timeout: 30 * time.Second},
		privateKey: privateKey,
	}

	return connector
}
