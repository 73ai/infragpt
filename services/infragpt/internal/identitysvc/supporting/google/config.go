package google

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	ClientID             string               `mapstructure:"client_id"`
	ClientSecret         string               `mapstructure:"client_secret"`
	RedirectURL          string               `mapstructure:"redirect_url"`
	CallbackPort         int                  `mapstructure:"callback_port"`
	StateTokenRepository StateTokenRepository `mapstructure:"-"`
}

func (c Config) New(ctx context.Context) (*Google, error) {
	if c.ClientID == "" {
		return nil, fmt.Errorf("client_id is required")
	}
	if c.ClientSecret == "" {
		return nil, fmt.Errorf("client_secret is required")
	}
	if c.RedirectURL == "" {
		return nil, fmt.Errorf("redirect_url is required")
	}
	if c.StateTokenRepository == nil {
		return nil, fmt.Errorf("state_token_repository is required")
	}
	if c.CallbackPort == 0 {
		return nil, fmt.Errorf("callback_port is required")
	}
	oauthConfig := &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURL,
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	return &Google{
		oauthConfig:          oauthConfig,
		stateTokenRepository: c.StateTokenRepository,
		callbackPort:         c.CallbackPort,
	}, nil
}
