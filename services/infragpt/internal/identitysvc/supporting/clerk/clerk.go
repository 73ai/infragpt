package clerk

import (
	"context"
	clerkapi "github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"net/http"
)

type clerk struct {
	port          int
	secretKey     string
	webhookSecret string
}

func (c clerk) Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error {
	webhookConfig := webhookServerConfig{
		port:                c.port,
		callbackHandlerFunc: handler,
	}
	return webhookConfig.startWebhookServer(ctx)
}

func (c Config) NewAuthMiddleware() func(http.Handler) http.Handler {
	clerkapi.SetKey(c.SecretKey)

	return clerkhttp.WithHeaderAuthorization()
}
