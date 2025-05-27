package infragpt

import (
	"context"
	"time"
)

// RequestStatus represents the status of an infrastructure request
type RequestStatus string

const (
	StatusNew             RequestStatus = "new"
	StatusAnalyzing       RequestStatus = "analyzing"
	StatusPendingApproval RequestStatus = "pending_approval"
	StatusApproved        RequestStatus = "approved"
	StatusRejected        RequestStatus = "rejected"
	StatusInProgress      RequestStatus = "in_progress"
	StatusCompleted       RequestStatus = "completed"
	StatusFailed          RequestStatus = "failed"
)

// CloudProvider represents a cloud service provider
type CloudProvider string

const (
	ProviderGCP CloudProvider = "gcp"
	ProviderAWS CloudProvider = "aws"
)

// Request represents an infrastructure change request
type Request struct {
	ID             string         `json:"id"`
	RequesterID    string         `json:"requester_id"`
	RequesterName  string         `json:"requester_name,omitempty"`
	RequesterEmail string         `json:"requester_email,omitempty"`
	ApproverID     string         `json:"approver_id"`
	ApproverName   string         `json:"approver_name,omitempty"`
	RawText        string         `json:"raw_text"`
	Resource       string         `json:"resource"`
	ResourceType   string         `json:"resource_type"`
	Action         string         `json:"action"`
	CloudProvider  CloudProvider  `json:"cloud_provider"`
	Status         RequestStatus  `json:"status"`
	PullRequestURL string         `json:"pull_request_url,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// ApprovalRequest represents a request for approval
type ApprovalRequest struct {
	RequestID        string            `json:"request_id"`
	RequesterID      string            `json:"requester_id"`
	RequesterName    string            `json:"requester_name,omitempty"`
	ApproverID       string            `json:"approver_id"`
	Resource         string            `json:"resource"`
	ResourceType     string            `json:"resource_type"`
	Action           string            `json:"action"`
	PullRequestURL   string            `json:"pull_request_url"`
	TerraformChanges *TerraformChanges `json:"terraform_changes"`
}

// Approval represents an approval decision
type Approval struct {
	RequestID   string    `json:"request_id"`
	ApproverID  string    `json:"approver_id"`
	Approved    bool      `json:"approved"`
	Comment     string    `json:"comment,omitempty"`
	ApprovedAt  time.Time `json:"approved_at"`
}

// ResourceState represents the current state of a cloud resource
type ResourceState struct {
	ResourceID        string                 `json:"resource_id"`
	ResourceType      string                 `json:"resource_type"`
	Provider          CloudProvider          `json:"provider"`
	Attributes        map[string]interface{} `json:"attributes"`
	ExistsInTerraform bool                   `json:"exists_in_terraform"`
	TerraformPath     string                 `json:"terraform_path,omitempty"`
}

// TerraformChanges represents changes to Terraform configurations
type TerraformChanges struct {
	Files       map[string]string `json:"files"` // Map of file path to content
	Description string            `json:"description"`
	Summary     string            `json:"summary"`
}

// PullRequestDetails contains information about a created PR
type PullRequestDetails struct {
	URL        string    `json:"url"`
	Number     int       `json:"number"`
	BranchName string    `json:"branch_name"`
	CreatedAt  time.Time `json:"created_at"`
}

// InfraGPTService is the main service interface for InfraGPT
type InfraGPTService interface {
	// HandleMessage processes an incoming message from a user
	HandleMessage(ctx context.Context, channel, user, text, threadTs string) error

	// ProcessRequest handles the lifecycle of an infrastructure request
	ProcessRequest(ctx context.Context, request *Request) error

	// ProcessApproval handles an approval response for a request
	ProcessApproval(ctx context.Context, approval *Approval) error

	// GetRequestStatus returns the current status of a request
	GetRequestStatus(ctx context.Context, requestID string) (*Request, error)

	// Start initializes and starts the service
	Start(ctx context.Context) error

	// Shutdown gracefully stops the service
	Shutdown(ctx context.Context) error
}