package infragpt

import (
	"context"
	"time"
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
)

type Integration struct {
	ID                      string
	OrganizationID          string
	UserID                  string
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

type IntegrationService interface {
	NewIntegration(ctx context.Context, cmd NewIntegrationCommand) (IntegrationAuthorizationIntent, error)
	AuthorizeIntegration(ctx context.Context, cmd AuthorizeIntegrationCommand) (Integration, error)
	ConfigureIntegration(ctx context.Context, cmd ConfigureIntegrationCommand) (Integration, error)
	RevokeIntegration(ctx context.Context, cmd RevokeIntegrationCommand) error
	Integrations(ctx context.Context, query IntegrationsQuery) ([]Integration, error)
	Integration(ctx context.Context, query IntegrationQuery) (Integration, error)
	Subscribe(ctx context.Context) error
}

type NewIntegrationCommand struct {
	OrganizationID string
	UserID         string
	ConnectorType  ConnectorType
}

type AuthorizeIntegrationCommand struct {
	ConnectorType  ConnectorType
	Code           string
	State          string
	InstallationID string
}

type ConfigureIntegrationCommand struct {
	OrganizationID string
	UserID         string
	ConnectorType  ConnectorType
	InstallationID string
	SetupAction    string
}

type RevokeIntegrationCommand struct {
	IntegrationID  string
	OrganizationID string
}

type IntegrationsQuery struct {
	OrganizationID string
	ConnectorType  ConnectorType
}

type IntegrationQuery struct {
	IntegrationID  string
	OrganizationID string
}
