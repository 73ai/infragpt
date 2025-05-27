# InfraGPT - Infrastructure as Code Management with AI

InfraGPT is a Slack bot that converts natural language infrastructure requests into Terraform code changes. It manages infrastructure through GitHub pull requests and includes an approval workflow for safe deployment.

## Architecture

The application follows a clean architecture pattern with clear separation of concerns:

```
infragpt/
├── cmd/                      # Command-line entry points
│   └── main.go               # Main application entry point
├── internal/                 # Private application code
│   └── infragptsvc/          # Core service implementation
│       ├── domain/           # Domain interfaces
│       │   ├── cloud.go      # Cloud provider interfaces
│       │   ├── github.go     # GitHub integration interfaces
│       │   ├── llm.go        # LLM integration interfaces
│       │   ├── slack.go      # Slack integration interfaces
│       │   └── store.go      # Storage interfaces
│       ├── service.go        # Main service implementation
│       └── supporting/       # Implementation of domain interfaces
│           ├── cloud/        # Cloud provider implementations
│           │   ├── factory.go # Cloud provider factory
│           │   └── gcp/      # GCP-specific implementation
│           ├── github/       # GitHub integration
│           ├── llm/          # LLM service integration
│           ├── slack/        # Slack bot implementation
│           └── store/        # Storage implementations
├── spec.go                   # Core domain types and interfaces
└── go.mod                    # Go module definition
```

## Core Components

1. **Slack Integration**: Processes messages from Slack, handles approval workflows
2. **GitHub Integration**: Manages infrastructure-as-code repository, creates pull requests
3. **LLM Service**: Analyzes natural language requests, generates Terraform code
4. **Cloud Provider**: Fetches current state of resources, provides Terraform import commands
5. **Core Service**: Orchestrates the workflow between all components

## Setup

### Prerequisites

- Go 1.20 or later
- Slack Bot Token & App-Level Token
- GitHub Personal Access Token
- GCP credentials (for accessing Google Cloud resources)
- OpenAI API Key (or other LLM provider)

### Environment Variables

Create a `.env` file with the following:

```
# Slack Configuration
SLACK_BOT_TOKEN=xoxb-your-token
SLACK_APP_TOKEN=xapp-your-token
DEFAULT_APPROVER_ID=UXXXXXXXX

# GitHub Configuration
GITHUB_TOKEN=ghp_your_token
GITHUB_REPO_OWNER=your-github-username-or-org
GITHUB_REPO_NAME=infragpt-iac

# LLM Configuration
LLM_API_KEY=your-api-key
LLM_MODEL=gpt-4o
LLM_URL=https://api.openai.com/v1

# Cloud Provider Configuration
GCP_CREDENTIALS_FILE=/path/to/credentials.json
```

### Building and Running

```bash
# Build the application
go build -o infragpt ./cmd

# Run the application
./infragpt
```

## Usage

1. **Request Infrastructure Change**:
   In Slack, mention the bot with your request:
   ```
   @infragpt Give view permission to customer-data pubsub topic for service-account@project.iam.gserviceaccount.com
   ```

2. **Review Changes**:
   InfraGPT analyzes the request, fetches current state, and creates a pull request with Terraform changes.

3. **Approve Request**:
   Designated approvers can click the Approve button in Slack to authorize the change.

4. **Apply Changes**:
   When approved, InfraGPT merges the pull request, triggering a GitHub Action to apply the changes.

## Implementation Notes

- The core logic follows the interface definitions in `spec.go`
- Domain interfaces define the required functionality for each component
- Supporting implementations provide concrete integrations with external services
- The main service orchestrates the workflow between all components

## Development

To extend the application:

1. Add new domain interfaces in `internal/infragptsvc/domain/` if needed
2. Implement the interfaces in `internal/infragptsvc/supporting/`
3. Update the main service in `internal/infragptsvc/service.go`

For adding support for new cloud providers:
1. Add a new provider type to `CloudProvider` in `spec.go`
2. Create a new implementation in `internal/infragptsvc/supporting/cloud/`
3. Update the cloud factory to support the new provider