package infragpt

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID
	ClerkUserID string
	Email       string
	FirstName   string
	LastName    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Organization struct {
	ID              uuid.UUID
	ClerkOrgID      string
	Name            string
	Slug            string
	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Metadata        OrganizationMetadata
}

type OrganizationMetadata struct {
	OrganizationID     uuid.UUID
	CompanySize        CompanySize
	TeamSize           TeamSize
	UseCases           []UseCase
	ObservabilityStack []ObservabilityStack
	CompletedAt        time.Time
	UpdatedAt          time.Time
}

type OrganizationMember struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	ClerkUserID    string
	ClerkOrgID     string
	Role           string
	JoinedAt       time.Time
}

type Profile struct {
	ID             uuid.UUID            `json:"id"`
	Name           string               `json:"name"`
	Slug           string               `json:"slug"`
	CreatedAt      time.Time            `json:"created_at"`
	Metadata       OrganizationMetadata `json:"metadata"`
	OrganizationID uuid.UUID            `json:"organization_id"`
	UserID         uuid.UUID            `json:"user_id"`
}

type CompanySize string

const (
	CompanySizeStartup    CompanySize = "startup"
	CompanySizeSmall      CompanySize = "small"
	CompanySizeMedium     CompanySize = "medium"
	CompanySizeLarge      CompanySize = "large"
	CompanySizeEnterprise CompanySize = "enterprise"
)

type TeamSize string

const (
	TeamSize1To5    TeamSize = "1-5"
	TeamSize6To20   TeamSize = "6-20"
	TeamSize21To50  TeamSize = "21-50"
	TeamSize51To100 TeamSize = "51-100"
	TeamSize100Plus TeamSize = "100+"
)

type UseCase string

const (
	UseCaseInfrastructureMonitoring         UseCase = "infrastructure_monitoring"
	UseCaseApplicationPerformanceMonitoring UseCase = "application_performance_monitoring"
	UseCaseLogManagement                    UseCase = "log_management"
	UseCaseIncidentResponse                 UseCase = "incident_response"
	UseCaseComplianceAuditing               UseCase = "compliance_auditing"
	UseCaseCostOptimization                 UseCase = "cost_optimization"
	UseCaseSecurityMonitoring               UseCase = "security_monitoring"
	UseCaseDevOpsAutomation                 UseCase = "devops_automation"
)

type ObservabilityStack string

const (
	ObservabilityStackDatadog               ObservabilityStack = "datadog"
	ObservabilityStackNewRelic              ObservabilityStack = "new_relic"
	ObservabilityStackSplunk                ObservabilityStack = "splunk"
	ObservabilityStackElasticStack          ObservabilityStack = "elastic_stack"
	ObservabilityStackPrometheusGrafana     ObservabilityStack = "prometheus_grafana"
	ObservabilityStackAppDynamics           ObservabilityStack = "app_dynamics"
	ObservabilityStackDynatrace             ObservabilityStack = "dynatrace"
	ObservabilityStackCloudWatch            ObservabilityStack = "cloudwatch"
	ObservabilityStackAzureMonitor          ObservabilityStack = "azure_monitor"
	ObservabilityStackGoogleCloudMonitoring ObservabilityStack = "google_cloud_monitoring"
	ObservabilityStackPagerDuty             ObservabilityStack = "pagerduty"
	ObservabilityStackOpsgenie              ObservabilityStack = "opsgenie"
	ObservabilityStackOther                 ObservabilityStack = "other"
)

type IdentityService interface {
	SubscribeUserCreated(context.Context, UserCreatedEvent) error
	SubscribeUserUpdated(context.Context, UserUpdatedEvent) error
	SubscribeUserDeleted(context.Context, UserDeletedEvent) error
	SubscribeOrganizationCreated(context.Context, OrganizationCreatedEvent) error
	SubscribeOrganizationUpdated(context.Context, OrganizationUpdatedEvent) error
	SubscribeOrganizationDeleted(context.Context, OrganizationDeletedEvent) error
	SubscribeOrganizationMemberAdded(context.Context, OrganizationMemberAddedEvent) error
	SubscribeOrganizationMemberUpdated(context.Context, OrganizationMemberUpdatedEvent) error
	SubscribeOrganizationMemberDeleted(context.Context, OrganizationMemberDeletedEvent) error

	SetOrganizationMetadata(context.Context, OrganizationMetadataCommand) error
	Profile(context.Context, ProfileQuery) (Profile, error)
}

type OrganizationMetadataCommand struct {
	OrganizationID     uuid.UUID
	CompanySize        CompanySize
	TeamSize           TeamSize
	UseCases           []UseCase
	ObservabilityStack []ObservabilityStack
}

type ProfileQuery struct {
	ClerkUserID string
	ClerkOrgID  string
}

type UserCreatedEvent struct {
	ClerkUserID string
	Email       string
	FirstName   string
	LastName    string
}

type UserUpdatedEvent struct {
	ClerkUserID string
	Email       string
	FirstName   string
	LastName    string
}

type OrganizationCreatedEvent struct {
	ClerkOrgID      string
	Name            string
	Slug            string
	CreatedByUserID string
}

type OrganizationUpdatedEvent struct {
	ClerkOrgID string
	Name       string
	Slug       string
}

type OrganizationMemberAddedEvent struct {
	ClerkUserID string
	ClerkOrgID  string
	Role        string
}

type OrganizationMemberDeletedEvent struct {
	ClerkUserID string
	ClerkOrgID  string
}

type OrganizationMemberUpdatedEvent struct {
	ClerkUserID string
	ClerkOrgID  string
	Role        string
}

type UserDeletedEvent struct {
	ClerkUserID string
}

type OrganizationDeletedEvent struct {
	ClerkOrgID string
}
