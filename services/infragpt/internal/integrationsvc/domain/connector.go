package domain

import (
	"context"
	"github.com/priyanshujain/infragpt/services/infragpt"
)

type Connector interface {
	// Authorization methods
	InitiateAuthorization(organizationID string, userID string) (infragpt.IntegrationAuthorizationIntent, error)
	CompleteAuthorization(authData infragpt.AuthorizationData) (infragpt.Credentials, error)
	ValidateCredentials(creds infragpt.Credentials) error
	RefreshCredentials(creds infragpt.Credentials) (infragpt.Credentials, error)
	RevokeCredentials(creds infragpt.Credentials) error
	
	// Webhook methods
	ConfigureWebhooks(integrationID string, creds infragpt.Credentials) error
	ValidateWebhookSignature(payload []byte, signature string, secret string) error
	
	// Event subscription method - each connector handles its own communication
	Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error
}
