package domain

import (
	"github.com/priyanshujain/infragpt/services/infragpt"
)

type Connector interface {
	InitiateAuthorization(organizationID string, userID string) (infragpt.IntegrationAuthorizationIntent, error)
	CompleteAuthorization(authData infragpt.AuthorizationData) (infragpt.Credentials, error)
	ValidateCredentials(creds infragpt.Credentials) error
	RefreshCredentials(creds infragpt.Credentials) (infragpt.Credentials, error)
	RevokeCredentials(creds infragpt.Credentials) error
	ConfigureWebhooks(integrationID string, creds infragpt.Credentials) error
	ValidateWebhookSignature(payload []byte, signature string, secret string) error
}
