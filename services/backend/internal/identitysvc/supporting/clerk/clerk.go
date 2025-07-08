package clerk

import (
	"context"
	"net/http"

	clerkapi "github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

type clerk struct {
	port          int
	secretKey     string
	webhookSecret string
}

func (c clerk) Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error {
	if c.port == 0 {
		panic("clerk: invalid port")
	}
	if c.webhookSecret == "" {
		panic("clerk: webhook secret is empty")
	}
	if c.secretKey == "" {
		panic("clerk: secret key is empty")
	}

	webhookConfig := webhookServerConfig{
		port:                c.port,
		webhookSecret:       c.webhookSecret,
		callbackHandlerFunc: handler,
	}
	return webhookConfig.startWebhookServer(ctx)
}

func (c Config) NewAuthMiddleware() func(http.Handler) http.Handler {
	clerkapi.SetKey(c.SecretKey)

	return clerkhttp.WithHeaderAuthorization()
}
