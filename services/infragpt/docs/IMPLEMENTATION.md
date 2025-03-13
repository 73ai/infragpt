# InfraGPT Implementation Plan

## Current Project Structure

```
infragpt/
├── cmd/
│   └── main.go              # Application entry point
├── generic/
│   └── httperrors/          # HTTP error handling utilities
├── infragptapi/
│   └── handler.go           # HTTP API handlers
├── internal/
│   └── infragptsvc/         # Core service implementation
│       ├── config.go        # Service configuration
│       ├── service.go       # Main service implementation
│       ├── domain/          # Domain models and interfaces
│       │   ├── integration.go  # Integration domain model
│       │   └── slack.go        # Slack domain interfaces
│       └── supporting/      # Infrastructure implementations
│           ├── postgres/    # Database implementation
│           │   ├── config.go      # Database configuration
│           │   ├── db.go          # Database connection
│           │   ├── models.go      # Database models
│           │   ├── infragpt_db.go # Repository implementation
│           │   └── queries/       # SQL queries
│           └── slack/       # Slack implementation
│               ├── app_mention.go # App mention handling
│               ├── config.go      # Slack configuration
│               ├── slack.go       # Slack client implementation
│               └── subscription.go # Event subscription
├── spec.go                  # Service interface definitions
└── go.mod                   # Go module definition
```

## Implementation Details

### Core Service Architecture

The InfraGPT service follows a clean architecture pattern with:

1. **Domain Layer** - Core business logic and interfaces
   - Defined in `internal/infragptsvc/domain/`
   - Contains domain models like `Integration`
   - Defines repository interfaces like `IntegrationRepository`

2. **Application Layer** - Service implementation
   - Implemented in `internal/infragptsvc/service.go`
   - Orchestrates domain operations
   - Handles business workflows

3. **Infrastructure Layer** - External integrations
   - Slack integration in `internal/infragptsvc/supporting/slack/`
   - Database implementation in `internal/infragptsvc/supporting/postgres/`

4. **API Layer** - HTTP endpoints
   - Implemented in `infragptapi/handler.go`
   - Handles HTTP requests and responses
   - Uses generic error handling from `generic/httperrors/`

### Service Interface

The main service interface is defined in `spec.go`:

```go
type Service interface {
    Integrations(context.Context, IntegrationsQuery) ([]Integration, error)
    CompleteSlackAuthentication(context.Context, CompleteSlackAuthenticationCommand) error
}
```

This interface will be expanded as new features are implemented according to the roadmap.

### Slack Integration

The Slack integration is implemented using the `slack-go/slack` library with socket mode support:

1. **Authentication Flow**
   - OAuth 2.0 implementation in `slack.go`
   - Token storage in PostgreSQL

2. **Event Handling**
   - Socket mode client for real-time events
   - App mention handling for responding to user commands

3. **Messaging**
   - Ability to reply to messages in threads
   - Support for structured messages (to be expanded)

### Database Implementation

PostgreSQL is used for persistent storage:

1. **Schema**
   - Integration tracking tables
   - Slack token storage

2. **Access Patterns**
   - Repository pattern for data access
   - SQL queries defined in `.sql` files and generated code

## Implementation Plan

### Phase 1: Core Infrastructure (Current)

- Complete the Slack integration with full event handling
- Implement thread-based conversation tracking
- Add support for interactive components
- Create comprehensive test suite

### Phase 2: Access Request Workflow

- Implement command parsing for access requests
- Create data model for access requests
- Build approval workflow
- Implement notification system

### Phase 3: Access Management

- Implement secure credential management
- Create command generation for access grants
- Build validation for access requests
- Implement execution tracking

## Technology Stack

- **Language**: Go 1.24
- **Database**: PostgreSQL
- **Libraries**:
  - github.com/slack-go/slack - Slack API client
  - github.com/jackc/pgx/v5 - PostgreSQL driver
  - golang.org/x/sync - Synchronization utilities
  - github.com/mitchellh/mapstructure - Configuration parsing
  - gopkg.in/yaml.v3 - YAML parsing
  - github.com/google/uuid - UUID generation

## Security Considerations

- Secure token storage in database
- Access control based on Slack identities
- Proper error handling and logging
- Secure configuration management