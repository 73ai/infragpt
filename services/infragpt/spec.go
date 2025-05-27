package infragpt

import (
	"context"
)

type IntegrationType string

var (
	IntegrationTypeSlack       IntegrationType = "slack"
	IntegrationTypeGithub      IntegrationType = "github"
	IntegrationTypeGoogleCloud IntegrationType = "google_cloud"
)

type IntegrationStatus string

var (
	IntegrationStatusActive     IntegrationStatus = "active"
	IntegrationStatusInactive   IntegrationStatus = "inactive"
	IntegrationStatusPending    IntegrationStatus = "pending"
	IntegrationStatusNotStarted IntegrationStatus = "not_started"
)

type Integration struct {
	Type   IntegrationType
	Status IntegrationStatus
}

type Service interface {
	Integrations(context.Context, IntegrationsQuery) ([]Integration, error)

	CompleteSlackIntegration(context.Context, CompleteSlackIntegrationCommand) error
}

type IntegrationsQuery struct {
	BusinessID string
}

type CompleteSlackIntegrationCommand struct {
	BusinessID string
	Code       string
}
