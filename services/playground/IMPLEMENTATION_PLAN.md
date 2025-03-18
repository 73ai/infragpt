# InfraGPT Implementation Plan

This document outlines the technical implementation plan for InfraGPT, integrating functionality from the prototype slackbot and githubbot components.

## Code Structure

```
infragpt/
│
├── cmd/
│   └── infragpt/
│       └── main.go           # Application entry point
│
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration loading and management
│   │
│   ├── core/
│   │   ├── orchestrator.go   # Main workflow orchestration
│   │   ├── request.go        # Request data structures
│   │   └── storage.go        # Storage interfaces and implementations
│   │
│   ├── slack/
│   │   ├── bot.go            # Slack bot implementation
│   │   ├── events.go         # Event handling for Slack
│   │   ├── approval.go       # Approval workflow handling
│   │   └── message.go        # Message formatting and parsing
│   │
│   ├── github/
│   │   ├── client.go         # GitHub API client
│   │   ├── repository.go     # Repository management
│   │   ├── content.go        # Content (file) operations
│   │   └── pullrequest.go    # PR creation and management
│   │
│   ├── llm/
│   │   ├── client.go         # LLM API client
│   │   ├── prompt.go         # Prompt generation
│   │   ├── parser.go         # Parse LLM responses
│   │   └── generator.go      # Generate Terraform code
│   │
│   ├── cloud/
│   │   ├── factory.go        # Cloud provider factory
│   │   ├── gcp/              # GCP provider implementation
│   │   │   ├── client.go
│   │   │   └── resources.go
│   │   └── aws/              # AWS provider implementation (future)
│   │
│   └── terraform/
│       ├── generator.go      # Generate Terraform configurations
│       ├── parser.go         # Parse existing Terraform files
│       └── validator.go      # Validate Terraform code
│
├── pkg/
│   ├── models/
│   │   └── models.go         # Shared data models
│   │
│   └── utils/
│       ├── logger.go         # Logging utilities
│       └── errors.go         # Error handling utilities
│
└── test/
    ├── mocks/                # Mock implementations for testing
    └── testdata/             # Test data files
```

## Component Integration

### 1. Core Component

```go
// internal/core/orchestrator.go
package core

import (
    "context"
    
    "github.com/company/infragpt/internal/slack"
    "github.com/company/infragpt/internal/github"
    "github.com/company/infragpt/internal/llm"
    "github.com/company/infragpt/internal/cloud"
    "github.com/company/infragpt/pkg/models"
)

// Orchestrator coordinates the overall workflow
type Orchestrator struct {
    slackClient    *slack.Client
    githubClient   *github.Client
    llmClient      *llm.Client
    cloudFactory   *cloud.Factory
    requestStorage RequestStorage
}

// ProcessRequest handles the entire lifecycle of an infrastructure request
func (o *Orchestrator) ProcessRequest(ctx context.Context, req *models.Request) error {
    // 1. Validate and enrich request with LLM
    enrichedReq, err := o.llmClient.EnrichRequest(ctx, req)
    if err != nil {
        return err
    }
    
    // 2. Get current state from cloud provider
    cloudClient, err := o.cloudFactory.GetClient(enrichedReq.CloudProvider)
    if err != nil {
        return err
    }
    
    currentState, err := cloudClient.GetResourceState(ctx, enrichedReq.Resource)
    if err != nil {
        return err
    }
    
    // 3. Generate Terraform code for changes
    terraformChanges, err := o.llmClient.GenerateTerraformChanges(ctx, enrichedReq, currentState)
    if err != nil {
        return err
    }
    
    // 4. Create/update files in GitHub
    prDetails, err := o.githubClient.CreatePullRequest(ctx, enrichedReq, terraformChanges)
    if err != nil {
        return err
    }
    
    // 5. Create approval request in Slack
    approvalReq := &models.ApprovalRequest{
        RequestID:       enrichedReq.ID,
        RequesterID:     enrichedReq.RequesterID,
        ApproverID:      enrichedReq.ApproverID,
        Resource:        enrichedReq.Resource,
        Action:          enrichedReq.Action,
        PullRequestURL:  prDetails.URL,
        TerraformChanges: terraformChanges,
    }
    
    err = o.slackClient.RequestApproval(ctx, approvalReq)
    if err != nil {
        return err
    }
    
    // 6. Update request status
    enrichedReq.Status = models.StatusPendingApproval
    enrichedReq.PullRequestURL = prDetails.URL
    
    return o.requestStorage.UpdateRequest(ctx, enrichedReq)
}

// HandleApproval processes an approval response
func (o *Orchestrator) HandleApproval(ctx context.Context, approval *models.Approval) error {
    // Implementation details
    // ...
}
```

### 2. Slack Integration

```go
// internal/slack/bot.go
package slack

import (
    "context"
    
    "github.com/slack-go/slack"
    "github.com/slack-go/slack/socketmode"
    "github.com/company/infragpt/pkg/models"
)

// Client handles Slack interactions
type Client struct {
    client       *slack.Client
    socketClient *socketmode.Client
    requestCh    chan *models.Request
    approvalCh   chan *models.Approval
    botUserID    string
}

// New creates a new Slack client
func New(token, appToken string) (*Client, error) {
    // Implementation details
    // ...
}

// Start begins listening for Slack events
func (c *Client) Start(ctx context.Context) error {
    go c.handleEvents(ctx)
    return c.socketClient.Run()
}

// RequestApproval sends an approval request to the specified approver
func (c *Client) RequestApproval(ctx context.Context, req *models.ApprovalRequest) error {
    // Implementation details based on slackbot code
    // ...
}

// GetRequestChannel returns a channel that emits new requests
func (c *Client) GetRequestChannel() <-chan *models.Request {
    return c.requestCh
}

// GetApprovalChannel returns a channel that emits approval decisions
func (c *Client) GetApprovalChannel() <-chan *models.Approval {
    return c.approvalCh
}
```

### 3. GitHub Integration

```go
// internal/github/client.go
package github

import (
    "context"
    
    "github.com/google/go-github/v69/github"
    "github.com/company/infragpt/pkg/models"
)

// Client handles GitHub operations
type Client struct {
    client        *github.Client
    repoOwner     string
    repoName      string
    defaultBranch string
}

// New creates a new GitHub client
func New(token, repoOwner, repoName string) (*Client, error) {
    // Implementation details
    // ...
}

// EnsureRepository ensures the IAC repository exists
func (c *Client) EnsureRepository(ctx context.Context) error {
    // Implementation details
    // ...
}

// CreatePullRequest creates a new pull request with Terraform changes
func (c *Client) CreatePullRequest(ctx context.Context, req *models.Request, changes *models.TerraformChanges) (*models.PullRequestDetails, error) {
    // Implementation details
    // ...
}

// GetFileContent retrieves file content from the repository
func (c *Client) GetFileContent(ctx context.Context, path, branch string) (string, error) {
    // Implementation details based on githubbot code
    // ...
}
```

### 4. LLM Integration

```go
// internal/llm/client.go
package llm

import (
    "context"
    
    "github.com/company/infragpt/pkg/models"
)

// Client handles interactions with the LLM service
type Client struct {
    apiKey     string
    apiURL     string
    modelName  string
}

// New creates a new LLM client
func New(apiKey, apiURL, modelName string) *Client {
    return &Client{
        apiKey:    apiKey,
        apiURL:    apiURL,
        modelName: modelName,
    }
}

// EnrichRequest analyzes a request and extracts structured information
func (c *Client) EnrichRequest(ctx context.Context, req *models.Request) (*models.Request, error) {
    // Implementation details
    // ...
}

// GenerateTerraformChanges creates Terraform code for the requested changes
func (c *Client) GenerateTerraformChanges(ctx context.Context, req *models.Request, currentState *models.ResourceState) (*models.TerraformChanges, error) {
    // Implementation details
    // ...
}
```

### 5. Cloud Integration

```go
// internal/cloud/factory.go
package cloud

import (
    "context"
    
    "github.com/company/infragpt/internal/cloud/gcp"
    "github.com/company/infragpt/pkg/models"
)

// Provider represents a cloud provider type
type Provider string

const (
    ProviderGCP Provider = "gcp"
    ProviderAWS Provider = "aws"
)

// Client interface for cloud provider operations
type Client interface {
    GetResourceState(ctx context.Context, resourceID string) (*models.ResourceState, error)
}

// Factory creates cloud provider clients
type Factory struct {
    gcpCredentials string
    // Add other provider credentials as needed
}

// GetClient returns a cloud provider client
func (f *Factory) GetClient(provider Provider) (Client, error) {
    switch provider {
    case ProviderGCP:
        return gcp.New(f.gcpCredentials)
    default:
        return nil, fmt.Errorf("unsupported cloud provider: %s", provider)
    }
}
```

## Shared Data Models

```go
// pkg/models/models.go
package models

import "time"

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

// Request represents an infrastructure change request
type Request struct {
    ID             string         `json:"id"`
    RequesterID    string         `json:"requester_id"`
    RequesterName  string         `json:"requester_name,omitempty"`
    ApproverID     string         `json:"approver_id"`
    ApproverName   string         `json:"approver_name,omitempty"`
    RawText        string         `json:"raw_text"`
    Resource       string         `json:"resource"`
    Action         string         `json:"action"`
    CloudProvider  string         `json:"cloud_provider"`
    Status         RequestStatus  `json:"status"`
    PullRequestURL string         `json:"pull_request_url,omitempty"`
    CreatedAt      time.Time      `json:"created_at"`
    UpdatedAt      time.Time      `json:"updated_at"`
}

// ApprovalRequest represents a request for approval
type ApprovalRequest struct {
    RequestID        string              `json:"request_id"`
    RequesterID      string              `json:"requester_id"`
    ApproverID       string              `json:"approver_id"`
    Resource         string              `json:"resource"`
    Action           string              `json:"action"`
    PullRequestURL   string              `json:"pull_request_url"`
    TerraformChanges *TerraformChanges   `json:"terraform_changes"`
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
    ResourceID   string                 `json:"resource_id"`
    ResourceType string                 `json:"resource_type"`
    Provider     string                 `json:"provider"`
    Attributes   map[string]interface{} `json:"attributes"`
    ExistsInTerraform bool              `json:"exists_in_terraform"`
}

// TerraformChanges represents changes to Terraform configurations
type TerraformChanges struct {
    Files        map[string]string `json:"files"`  // Map of file path to content
    Description  string            `json:"description"`
    Summary      string            `json:"summary"`
}

// PullRequestDetails contains information about a created PR
type PullRequestDetails struct {
    URL         string    `json:"url"`
    Number      int       `json:"number"`
    BranchName  string    `json:"branch_name"`
    CreatedAt   time.Time `json:"created_at"`
}
```

## Implementation Steps

1. **Core Infrastructure Setup**
   - Set up Go module structure
   - Implement configuration loading
   - Set up logging framework

2. **Slack Integration**
   - Port and enhance existing slackbot code
   - Implement event handling and routing
   - Build approval workflow system

3. **GitHub Integration**
   - Port and enhance existing githubbot code for repository management
   - Implement file operations and PR creation
   - Add webhook handling for PR status updates

4. **LLM Integration**
   - Implement API client for LLM service
   - Build prompt generation for request analysis
   - Create Terraform code generation capability

5. **Cloud Provider Integration**
   - Implement GCP client for resource state queries
   - Add support for Terraform state conversion
   - Build resource discovery capabilities

6. **Core Orchestration**
   - Implement request lifecycle management
   - Build workflow coordination between components
   - Add error handling and retry mechanisms

7. **Testing and Refinement**
   - Implement unit and integration tests
   - Create end-to-end test scenarios
   - Refine based on test results

## Deployment Plan

1. **Development Environment**
   - Local Docker-based deployment
   - Use ngrok for Slack and GitHub webhooks

2. **Staging Environment**
   - Kubernetes deployment
   - Integration with staging Slack workspace
   - Separate GitHub organization for testing

3. **Production Environment**
   - Production Kubernetes cluster
   - Integration with production Slack workspace
   - Production GitHub organization
   - Monitoring and alerting setup