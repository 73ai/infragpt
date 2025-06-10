package clerk

import "github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/domain"

type Config struct {
	Port           int    `mapstructure:"port"`
	WebhookSecret  string `mapstructure:"webhook_secret"`
	PublishableKey string `mapstructure:"publishable_key"`
}

func (c Config) NewAuthService() domain.AuthService {
	return &clerk{
		port:           c.Port,
		publishableKey: c.PublishableKey,
		webhookSecret:  c.WebhookSecret,
	}
}
