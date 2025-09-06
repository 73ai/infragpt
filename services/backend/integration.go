package backend

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ConnectorType string

const (
	ConnectorTypeSlack     ConnectorType = "slack"
	ConnectorTypeGithub    ConnectorType = "github"
	ConnectorTypeGCP       ConnectorType = "gcp"
	ConnectorTypeAWS       ConnectorType = "aws"
	ConnectorTypePagerDuty ConnectorType = "pagerduty"
	ConnectorTypeDatadog   ConnectorType = "datadog"
)

type AuthorizationType string

const (
	AuthorizationTypeOAuth2       AuthorizationType = "oauth2"
	AuthorizationTypeInstallation AuthorizationType = "installation"
	AuthorizationTypeAPIKey       AuthorizationType = "api_key"
)

type CredentialType string

const (
	CredentialTypeOAuth2         CredentialType = "oauth2"
	CredentialTypeToken          CredentialType = "token"
	CredentialTypeServiceAccount CredentialType = "service_account"
)

type IntegrationStatus string

const (
	IntegrationStatusActive     IntegrationStatus = "active"
	IntegrationStatusInactive   IntegrationStatus = "inactive"
	IntegrationStatusPending    IntegrationStatus = "pending"
	IntegrationStatusNotStarted IntegrationStatus = "not_started"
	IntegrationStatusSuspended  IntegrationStatus = "suspended"
	IntegrationStatusDeleted    IntegrationStatus = "deleted"
)

type Integration struct {
	ID                      uuid.UUID
	OrganizationID          uuid.UUID
	UserID                  uuid.UUID
	ConnectorType           ConnectorType
	Status                  IntegrationStatus
	BotID                   string
	ConnectorUserID         string
	ConnectorOrganizationID string
	Metadata                map[string]string
	CreatedAt               time.Time
	UpdatedAt               time.Time
	LastUsedAt              *time.Time
}

type IntegrationAuthorizationIntent struct {
	Type AuthorizationType
	URL  string
}

type AuthorizationData struct {
	Code           string
	State          string
	InstallationID string
}

// Credentials contains sensitive authentication data in data map.
// SECURITY NOTE: The Data field may contain private keys and other sensitive information.
// This data should be encrypted before storage and never logged in plaintext.
type Credentials struct {
	Type             CredentialType
	Data             map[string]string
	ExpiresAt        *time.Time
	OrganizationInfo *OrganizationInfo
}

type OrganizationInfo struct {
	ExternalID string
	Name       string
	Metadata   map[string]string
}

type CredentialValidationResult struct {
	Valid   bool
	Details any
	Errors  []string
}

type IntegrationService interface {
	NewIntegration(ctx context.Context, cmd NewIntegrationCommand) (IntegrationAuthorizationIntent, error)
	AuthorizeIntegration(ctx context.Context, cmd AuthorizeIntegrationCommand) (Integration, error)
	SyncIntegration(ctx context.Context, cmd SyncIntegrationCommand) error
	RevokeIntegration(ctx context.Context, cmd RevokeIntegrationCommand) error
	Integrations(ctx context.Context, query IntegrationsQuery) ([]Integration, error)
	Integration(ctx context.Context, query IntegrationQuery) (Integration, error)
	ValidateCredentials(ctx context.Context, connectorType ConnectorType, credentials map[string]any) (CredentialValidationResult, error)
	Subscribe(ctx context.Context) error
}

type NewIntegrationCommand struct {
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	ConnectorType  ConnectorType
}

type AuthorizeIntegrationCommand struct {
	ConnectorType  ConnectorType
	Code           string
	State          string
	InstallationID string
}

type RevokeIntegrationCommand struct {
	IntegrationID  uuid.UUID
	OrganizationID uuid.UUID
}

type IntegrationsQuery struct {
	OrganizationID uuid.UUID
	ConnectorType  ConnectorType
	Status         IntegrationStatus
}

type IntegrationQuery struct {
	IntegrationID  uuid.UUID
	OrganizationID uuid.UUID
}

type SyncIntegrationCommand struct {
	IntegrationID  uuid.UUID
	OrganizationID uuid.UUID
	Parameters     map[string]string
}
