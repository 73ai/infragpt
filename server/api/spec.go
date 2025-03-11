package api

import (
	"context"
	"time"
)

// AccessRequestStatus defines the possible states of an access request
type AccessRequestStatus string

const (
	StatusPending  AccessRequestStatus = "pending"
	StatusApproved AccessRequestStatus = "approved"
	StatusDenied   AccessRequestStatus = "denied"
)

// ChangeType defines the type of infrastructure change requested
type ChangeType string

const (
	ChangeCreate   ChangeType = "create"
	ChangeUpdate   ChangeType = "update"
	ChangeDelete   ChangeType = "delete"
	ChangeGrant    ChangeType = "grant"
	ChangeRevoke   ChangeType = "revoke"
)

// CloudProvider defines the supported cloud service providers
type CloudProvider string

const (
	ProviderGCP   CloudProvider = "gcp"
	ProviderAWS   CloudProvider = "aws"
	ProviderAzure CloudProvider = "azure"
)

// CloudResource represents a resource in a cloud provider
type CloudResource struct {
	Type       string
	Name       string
	Project    string
	Attributes map[string]string
}

// InfraRequest represents a request to change infrastructure
type InfraRequest struct {
	ID            string
	RequesterID   string
	RequesterName string
	Provider      CloudProvider
	ChangeType    ChangeType
	Resource      CloudResource
	TargetResource CloudResource    // For relationships like permissions
	RequestedAt   time.Time
	Status        AccessRequestStatus
	ApproverID    string
	TerraformChanges map[string]string // Map of file paths to content
	PullRequestURL   string
	CompletedAt   time.Time
}

// ParseNLPCommand represents a command to parse a natural language request
type ParseNLPCommand struct {
	UserID   string
	UserName string
	Message  string
}

// FetchStateCommand represents a command to fetch current infrastructure state
type FetchStateCommand struct {
	Provider CloudProvider
	Project  string
}

// GenerateTerraformCommand represents a command to generate Terraform files
type GenerateTerraformCommand struct {
	Request InfraRequest
	CurrentState interface{}
}

// CreatePullRequestCommand represents a command to create a GitHub pull request
type CreatePullRequestCommand struct {
	Request        InfraRequest
	RepoOwner      string
	RepoName       string
	Branch         string
	TerraformFiles map[string]string
}

// ApproveChangeCommand represents a command to approve an infrastructure change
type ApproveChangeCommand struct {
	RequestID  string
	Approved   bool
	ApproverID string
	Message    string
	AutoMerge  bool
}

// ExecuteTerraformCommand represents a command to execute Terraform changes
type ExecuteTerraformCommand struct {
	RequestID  string
	RepoOwner  string
	RepoName   string
	Branch     string
}

// Service defines the core operations of the InfraGPT service
type Service interface {
	// Natural language processing
	ParseRequest(ctx context.Context, command ParseNLPCommand) (*InfraRequest, error)
	
	// Infrastructure state management
	FetchCurrentState(ctx context.Context, command FetchStateCommand) (interface{}, error)
	GenerateTerraform(ctx context.Context, command GenerateTerraformCommand) (map[string]string, error)
	
	// GitHub integration
	EnsureRepository(ctx context.Context, owner string, name string) error
	CreatePullRequest(ctx context.Context, command CreatePullRequestCommand) (string, error)
	
	// Workflow management
	ApproveChange(ctx context.Context, command ApproveChangeCommand) error
	ExecuteTerraform(ctx context.Context, command ExecuteTerraformCommand) error
}

// NLPService defines operations for natural language processing
type NLPService interface {
	ExtractIntent(ctx context.Context, message string) (ChangeType, error)
	ExtractEntities(ctx context.Context, message string, provider CloudProvider) (CloudResource, CloudResource, error)
	ValidateRequest(ctx context.Context, request InfraRequest) error
}

// TerraformService defines operations for Terraform management
type TerraformService interface {
	ConvertStateToTerraform(ctx context.Context, state interface{}, provider CloudProvider) (map[string]string, error)
	GenerateChanges(ctx context.Context, request InfraRequest, currentConfig map[string]string) (map[string]string, error)
	ValidateConfiguration(ctx context.Context, config map[string]string) error
	ExecutePlan(ctx context.Context, config map[string]string) error
}

// GitHubService defines operations for GitHub integration
type GitHubService interface {
	CreateRepository(ctx context.Context, owner, name string) error
	CreateBranch(ctx context.Context, owner, repo, base, branch string) error
	CommitFiles(ctx context.Context, owner, repo, branch string, files map[string]string, message string) error
	CreatePR(ctx context.Context, owner, repo, base, head, title, body string) (string, error)
	GetPRStatus(ctx context.Context, owner, repo string, number int) (string, error)
	MergePR(ctx context.Context, owner, repo string, number int) error
}

// CSPService defines operations for Cloud Service Provider interaction
type CSPService interface {
	GetCurrentState(ctx context.Context, provider CloudProvider, project string) (interface{}, error)
	ValidateAccess(ctx context.Context, provider CloudProvider, credentials interface{}) error
	GetAuthorizationURL(ctx context.Context, provider CloudProvider, callbackURL string) (string, error)
	ExchangeAuthorizationCode(ctx context.Context, provider CloudProvider, code string) (interface{}, error)
	StoreCredentials(ctx context.Context, provider CloudProvider, userID string, credentials interface{}) error
	GetCredentials(ctx context.Context, provider CloudProvider, userID string) (interface{}, error)
}

// SlackAuthService defines operations for Slack authentication and authorization
type SlackAuthService interface {
	GetAuthorizationURL(ctx context.Context, callbackURL string) (string, error)
	ExchangeAuthorizationCode(ctx context.Context, code string) (string, error) // Returns access token
	ValidateToken(ctx context.Context, token string) (bool, error)
	StoreToken(ctx context.Context, userID string, token string) error
	GetToken(ctx context.Context, userID string) (string, error)
}

// GitHubAuthService defines operations for GitHub authentication and authorization
type GitHubAuthService interface {
	GetAuthorizationURL(ctx context.Context, callbackURL string) (string, error)
	ExchangeAuthorizationCode(ctx context.Context, code string) (string, error) // Returns access token
	ValidateToken(ctx context.Context, token string) (bool, error)
	StoreToken(ctx context.Context, userID string, token string) error
	GetToken(ctx context.Context, userID string) (string, error)
}