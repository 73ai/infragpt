package clerk

import "context"

type clerk struct {
	port           int
	publishableKey string
	webhookSecret  string
}

func (c clerk) Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error {
	webhookConfig := webhookServerConfig{
		port:                c.port,
		callbackHandlerFunc: handler,
	}
	return webhookConfig.startWebhookServer(ctx)
}
