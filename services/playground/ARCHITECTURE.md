# InfraGPT: Infrastructure as Code Management with AI

## Architecture Overview

InfraGPT is a Slack-integrated service that uses natural language to handle infrastructure change requests through Terraform. The system processes user requests, generates Terraform code, creates GitHub pull requests, and applies changes after approval.

## Components

### 1. Slack Integration Module (`slack/`)
- **Responsibilities**:
  - Process incoming Slack messages with the @infragpt mention
  - Extract user information and request details
  - Present approval workflows to authorized approvers
  - Send notifications about request status
  - Respond to user queries about request status

- **Key Interfaces**:
  - `SlackBot`: Main integration point with Slack API
  - `MessageProcessor`: Extracts structured requests from natural language
  - `ApprovalManager`: Handles approval workflows
  - `NotificationService`: Sends updates to users

### 2. GitHub Integration Module (`github/`)
- **Responsibilities**:
  - Manage the InfraGPT-IAC repository
  - Create/update Terraform configuration files
  - Generate pull requests with proposed changes
  - Monitor PR status and trigger necessary actions
  - Checkout code and manage branches

- **Key Interfaces**:
  - `GitHubClient`: Main integration point with GitHub API
  - `RepositoryManager`: Handles repository operations
  - `PullRequestManager`: Creates and manages pull requests
  - `FileManager`: Handles file operations in the repository

### 3. LLM Integration Module (`llm/`)
- **Responsibilities**:
  - Process natural language requests into structured tasks
  - Generate Terraform code from structured tasks
  - Validate generated Terraform against best practices
  - Explain proposed infrastructure changes in human terms

- **Key Interfaces**:
  - `NLProcessor`: Converts natural language to structured tasks
  - `CodeGenerator`: Generates Terraform code
  - `ChangeExplainer`: Creates human-readable summaries

### 4. Cloud Provider Module (`cloud/`)
- **Responsibilities**:
  - Authenticate with cloud providers (GCP, AWS, etc.)
  - Fetch current resource state
  - Generate Terraform imports for existing resources
  - Validate proposed changes against permissions and quotas

- **Key Interfaces**:
  - `CloudClientFactory`: Creates provider-specific clients
  - `ResourceExplorer`: Discovers existing resources
  - `IAMValidator`: Checks permissions
  - `TerraformImporter`: Generates import commands

### 5. Core Orchestration Service (`core/`)
- **Responsibilities**:
  - Coordinate workflow between components
  - Maintain state of in-progress requests
  - Handle error conditions
  - Manage configuration and environment

- **Key Interfaces**:
  - `RequestManager`: Tracks request lifecycle
  - `WorkflowEngine`: Manages the request processing steps
  - `ConfigurationManager`: Handles service configuration
  - `ErrorHandler`: Processes and responds to errors

## Request Lifecycle

1. **Request Intake**:
   - User mentions @infragpt in Slack with a request
   - Slack module captures the request and user information
   - System acknowledges receipt of request

2. **Request Analysis**:
   - LLM module parses natural language into structured operations
   - System identifies affected cloud resources
   - Core module creates a unique request ID and tracks state

3. **Current State Assessment**:
   - Cloud module queries current state of identified resources
   - If resources exist but aren't in Terraform, system generates imports
   - GitHub module checks for existing configuration files

4. **Change Preparation**:
   - LLM module generates Terraform changes
   - GitHub module creates a new branch
   - Changes are committed to the branch
   - Pull request is created with explanation of changes

5. **Approval Workflow**:
   - Slack module notifies designated approvers
   - Approver reviews and approves/rejects via Slack
   - System records the approval decision

6. **Change Implementation**:
   - On approval, GitHub Action is triggered
   - Terraform changes are applied
   - System monitors the apply process
   - Results are reported back through Slack

7. **Completion**:
   - User is notified of completed changes
   - System updates internal state
   - Change is documented in audit logs

## Data Flow

```
┌─────────────┐        ┌─────────────┐        ┌─────────────┐
│             │        │             │        │             │
│    Slack    │◄──────►│    Core     │◄──────►│    GitHub   │
│             │        │             │        │             │
└─────────────┘        └──────┬──────┘        └─────────────┘
                              │
                              ▼
                     ┌─────────────────┐
                     │                 │
                     │      LLM        │
                     │                 │
                     └────────┬────────┘
                              │
                              ▼
                     ┌─────────────────┐
                     │                 │
                     │     Cloud       │
                     │                 │
                     └─────────────────┘
```

## Error Handling

- **Validation Errors**: When requests can't be parsed or resources aren't accessible
  - Send meaningful error messages to the user
  - Suggest corrections or alternatives

- **Timeout Handling**: For long-running operations
  - Implement polling mechanisms
  - Provide status updates to users

- **Recovery Mechanisms**: For failed operations
  - Maintain idempotent operations
  - Implement rollback procedures where possible

## Security Considerations

- **Authentication**: Secure access to all integrated systems
  - Use secrets management for tokens and credentials
  - Implement minimal required permissions

- **Authorization**: Enforce proper permissions for operations
  - Validate user permissions before executing changes
  - Implement approval workflows for sensitive operations

- **Audit**: Track all operations for compliance
  - Log all requests, approvals, and changes
  - Maintain history of changes for auditing

## Implementation Phases

### Phase 1: Basic Integration
- Implement Slack bot with basic request handling
- Create GitHub repository management functionality
- Implement simple LLM integration for parsing requests
- Set up basic approval workflow

### Phase 2: Enhanced Functionality
- Add multi-cloud provider support
- Implement Terraform import for existing resources
- Enhance LLM capabilities for complex requests
- Add error recovery mechanisms

### Phase 3: Production Readiness
- Implement comprehensive logging and monitoring
- Add user management and fine-grained permissions
- Create dashboard for request tracking
- Optimize performance and reliability