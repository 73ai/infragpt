# Contributing to InfraGPT

Thank you for your interest in contributing to InfraGPT! This guide provides comprehensive information about our development standards, architectural patterns, and best practices learned through building a production-ready multi-service platform.

## Table of Contents

- [Project Overview](#project-overview)
- [Getting Started](#getting-started)
- [Architecture Guidelines](#architecture-guidelines)
- [Development Workflow](#development-workflow)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing Standards](#testing-standards)
- [Security Best Practices](#security-best-practices)
- [Integration System Guidelines](#integration-system-guidelines)
- [Pull Request Process](#pull-request-process)

## Project Overview

InfraGPT is a multi-service platform that provides AI-powered infrastructure management through Slack integration. The system consists of three main services:

### Core Services
1. **InfraGPT Core Service** - Main Slack bot functionality and workflow orchestration
2. **Identity Service** - User authentication and organization management via Clerk
3. **Integration Service** - External service connectivity with multi-connector architecture

### Key Features
- Real-time Slack integration via Socket Mode
- Multi-tenant organization management
- Encrypted credential storage for external services
- AI agent integration via gRPC
- Clean architecture with domain-driven design

## Getting Started

### Prerequisites

- **Go 1.24+** - Modern Go with generics and improved type system
- **PostgreSQL** - Primary database for all services
- **Git** - Version control
- **Docker** (optional) - For integration testing with testcontainers

### Environment Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/priyanshujain/infragpt.git
   cd infragpt/services/infragpt
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Set up configuration:**
   ```yaml
   # config.yaml
   port: 8080
   grpc_port: 9090
   log_level: "info"
   
   slack:
     client_id: "${SLACK_CLIENT_ID}"
     app_token: "${SLACK_APP_TOKEN}"
   
   database:
     host: "localhost"
     port: 5432
     db_name: "infragpt"
     user: "infragpt"
     password: "${DB_PASSWORD}"
   
   agent:
     endpoint: "[::]:50051"
   
   identity:
     clerk:
       webhook_secret: "${CLERK_WEBHOOK_SECRET}"
       publishable_key: "${CLERK_PUBLISHABLE_KEY}"
   
   integrations:
     slack:
       client_id: "${SLACK_CLIENT_ID}"
       bot_token: "${SLACK_BOT_TOKEN}"
     github:
       app_id: "${GITHUB_APP_ID}"
       webhook_port: 8081
   ```

4. **Run the application:**
   ```bash
   go run ./cmd/main.go
   ```

5. **Run tests:**
   ```bash
   go test ./...
   ```

## Architecture Guidelines

### Clean Architecture Principles

InfraGPT follows clean architecture with strict layer separation:

```
┌─────────────────────────────────────────┐
│                API Layer                 │
│  HTTP Handlers │ gRPC Server │ Webhooks │
└─────────────────────────────────────────┘
┌─────────────────────────────────────────┐
│            Application Layer             │
│        Service Implementations          │
└─────────────────────────────────────────┘
┌─────────────────────────────────────────┐
│              Domain Layer                │
│  Business Logic │ Interfaces │ Models   │
└─────────────────────────────────────────┘
┌─────────────────────────────────────────┐
│           Infrastructure Layer           │
│  PostgreSQL │ Slack │ Clerk │ Agents   │
└─────────────────────────────────────────┘
```

### Key Architectural Patterns

#### 1. Repository Pattern
Domain interfaces with infrastructure implementations:

```go
// Domain interface (internal/service/domain/)
type IntegrationRepository interface {
    Store(ctx context.Context, integration Integration) error
    FindByOrganization(ctx context.Context, orgID string) ([]Integration, error)
}

// Infrastructure implementation (internal/service/supporting/postgres/)
type integrationRepository struct {
    queries *Queries
}
```

#### 2. Command/Query Pattern
Service interfaces use command/query structure for user actions:

```go
type IntegrationService interface {
    // Commands (state-changing operations)
    NewIntegration(ctx context.Context, cmd NewIntegrationCommand) (IntegrationAuthorizationIntent, error)
    AuthorizeIntegration(ctx context.Context, cmd AuthorizeIntegrationCommand) (Integration, error)
    
    // Queries (read-only operations)
    Integrations(ctx context.Context, query IntegrationsQuery) ([]Integration, error)
}
```

#### 3. Subscribe Pattern
Consistent event handling following Clerk authentication service:

```go
type Service interface {
    Subscribe(ctx context.Context) error
}

func (s *service) Subscribe(ctx context.Context) error {
    // Start event processing in goroutines
    // Handle graceful shutdown with context cancellation
}
```

### Directory Structure Standards

```
internal/servicename/
├── config.go              # Service configuration
├── service.go              # Service implementation
├── domain/                 # Business logic and interfaces
│   ├── models.go          # Domain models
│   └── interfaces.go      # Repository interfaces
└── supporting/            # Infrastructure implementations
    ├── postgres/          # Database layer
    │   ├── config.go      # Database configuration
    │   ├── repository.go  # Repository implementation
    │   ├── queries/       # SQL query files
    │   └── schema/        # Database schema
    └── external/          # External service integrations
```

## Development Workflow

### Branching Strategy

- **`master`** - Main development branch (production-ready)
- **`feature/description`** - New feature development
- **`fix/description`** - Bug fixes
- **`refactor/description`** - Code refactoring

### Development Process

1. **Create feature branch** from master
2. **Implement changes** following coding standards
3. **Add comprehensive tests** (unit, integration)
4. **Update documentation** if needed
5. **Run full test suite** and ensure all pass
6. **Submit pull request** with detailed description

### Database Changes

1. **Create migration scripts** in `migrations/` directory
2. **Update SQLC queries** in appropriate `queries/` directories
3. **Regenerate code** with `sqlc generate`
4. **Test migrations** with fresh database

## Code Style Guidelines

### Modern Go Practices

#### 1. Type Usage
```go
// ✅ Use 'any' instead of 'interface{}'
func handleEvent(ctx context.Context, event any) error
var payload map[string]any

// ❌ Avoid outdated interface{} syntax
func handleEvent(ctx context.Context, event interface{}) error
```

#### 2. Struct Design
```go
// ✅ Clean internal domain structs (no JSON tags)
type MessageEvent struct {
    EventType EventType
    TeamID    string
    ChannelID string
    RawEvent  map[string]any
    CreatedAt time.Time
}

// ✅ API boundary structs (with JSON tags)
type authorizeRequest struct {
    OrganizationID string `json:"organization_id"`
    ConnectorType  string `json:"connector_type"`
}
```

#### 3. Error Handling
```go
// ✅ Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to process integration: %w", err)
}

// ✅ Use structured logging with context
slog.Error("integration processing failed", 
    "integration_id", id, 
    "connector_type", connectorType, 
    "error", err)
```

### Naming Conventions

- **CamelCase** for exported symbols
- **camelCase** for non-exported symbols
- **Descriptive names** over abbreviations
- **Package names** match directory names
- **Interface names** describe behavior, not implementation

### Comment Guidelines

- **Minimal comments** - prefer self-documenting code
- **Explain 'why'**, not 'what'
- **Document public APIs** and complex business logic
- **Remove obvious comments** that restate the code

```go
// ❌ Unnecessary comment
// Set the user ID to the provided user ID
user.ID = userID

// ✅ Valuable comment
// Encrypt credentials using AES-256-GCM with unique nonce per operation
encryptedData, err := encryptCredentials(credentials)
```

## Testing Standards

### Test Types

#### 1. Unit Tests
- Test individual functions and methods
- Mock external dependencies
- Focus on business logic validation

#### 2. Integration Tests
- Use `testcontainers-go` for PostgreSQL
- Test complete workflows
- Validate service interactions

#### 3. Repository Tests
- Test database operations
- Validate query correctness
- Test transaction handling

### Test Organization

```go
// Test files alongside source code
internal/integrationsvc/
├── service.go
├── service_test.go
├── domain/
│   ├── integration.go
│   └── integration_test.go
└── supporting/
    └── postgres/
        ├── repository.go
        └── repository_test.go
```

### Test Utilities

Create test utilities in `*test` packages:

```go
// identitytest/identitytest.go
package identitytest

func NewMockUserRepository() domain.UserRepository {
    // Mock implementation for testing
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests only
go test -tags=integration ./...

# Run specific service tests
go test ./internal/integrationsvc/...
```

## Security Best Practices

### Credential Management

#### 1. Encryption Standards
- **AES-256-GCM** for all credential storage
- **Unique nonce** per encryption operation
- **Environment-based** key derivation
- **Key rotation** support with versioning

```go
// Example credential encryption
type CredentialRepository interface {
    Store(ctx context.Context, cred IntegrationCredential) error
    // Credentials are automatically encrypted before storage
}
```

#### 2. Webhook Security
- **Signature validation** for all webhook endpoints
- **Timing-safe comparison** to prevent timing attacks
- **Panic recovery** middleware for production safety

```go
// Slack HMAC-SHA256 validation
func (s *slackConnector) ValidateWebhookSignature(payload []byte, signature string, secret string) error {
    expectedSignature := s.computeSignature(payload, secret)
    if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
        return fmt.Errorf("webhook signature validation failed")
    }
    return nil
}
```

### Access Control

- **Organization-scoped** access for all operations
- **Clerk JWT** token validation for protected endpoints
- **Audit logging** for sensitive operations
- **Principle of least privilege** for service permissions

### Environment Variables

```bash
# Sensitive configuration via environment variables
export SLACK_CLIENT_SECRET="..."
export GITHUB_PRIVATE_KEY="..."
export ENCRYPTION_SALT="..."
export CLERK_WEBHOOK_SECRET="..."
```

## Core Design Principles & Learnings

### Fundamental Design Principles

These principles were discovered and refined through building the integration system:

#### 1. Connector Events Stay in Connectors
- **Principle**: No global event definitions - all events remain in their respective connector packages
- **Rationale**: Eliminates complexity and prevents over-engineering
- **Implementation**: Each connector defines its own event types (`MessageEvent`, `WebhookEvent`)
- **Benefits**: Clean separation, no coupling between connectors, easier maintenance

#### 2. Simple Service Routing
- **Principle**: Route by connector type first, then by event type within connector
- **Implementation**: Type-based routing with service-level business logic handlers
- **Pattern**: `switch connectorType -> switch eventType -> business logic`

#### 3. No Over-Engineering
- **Principle**: Start simple, add abstractions only when actually needed
- **Example**: Initially avoided complex event hierarchies and domain events
- **Result**: Clean, maintainable code that's easy to understand and extend

#### 4. Clerk Pattern Consistency
- **Principle**: Follow exact same Subscribe pattern as Clerk authentication service
- **Implementation**: Every service implements `Subscribe(ctx context.Context) error`
- **Benefits**: Consistent event handling, familiar patterns, reliable error handling

### Code Style Conventions (Production Learnings)

#### Naming Conventions
```go
// ✅ Use full descriptive names
integrationRepository  // not integrationRepo
credentialRepository   // not credRepo
webhookServerConfig    // not webhookConfig

// ✅ Consistent config naming within packages
type Config struct {    // Every connector has Config struct
    ClientID string
    // ...
}
```

#### Struct Design Patterns
```go
// ✅ Local anonymous structs for internal API responses
func (h *httpHandler) authorize() func(w http.ResponseWriter, r *http.Request) {
    type request struct {  // Local to this handler
        OrganizationID string `json:"organization_id"`
    }
    type response struct { // Local to this handler
        Type string `json:"type"`
        URL  string `json:"url"`
    }
    // Implementation...
}

// ✅ Clean internal domain structs (no JSON tags)
type MessageEvent struct {
    EventType EventType
    TeamID    string
    RawEvent  map[string]any  // Use 'any' not 'interface{}'
    CreatedAt time.Time
}
```

#### Export Control
```go
// ✅ Minimal exports - only what needs to be public
package slack

type Config struct {           // Exported - needed by service
    ClientID string
}

func (c Config) NewConnector() domain.Connector { // Exported - factory method
    return &slackConnector{    // Not exported - implementation detail
        config: c,
    }
}

type slackConnector struct {   // Not exported - internal implementation
    config Config
}
```

#### Configuration Patterns
```go
// ✅ Mapstructure tags for YAML parsing
type Config struct {
    ClientID      string   `mapstructure:"client_id"`
    ClientSecret  string   `mapstructure:"client_secret"`
    RedirectURL   string   `mapstructure:"redirect_url"`
    Scopes        []string `mapstructure:"scopes"`
}

// ✅ Factory pattern for connector creation
func (c Config) NewConnector() domain.Connector {
    return &slackConnector{
        config: c,
        client: &http.Client{Timeout: 30 * time.Second},
    }
}

// ✅ Conditional registration in service
if c.Slack.ClientID != "" {
    connectors[infragpt.ConnectorTypeSlack] = c.Slack.NewConnector()
}
```

### Event Architecture Best Practices

#### Event Design Principles
```go
// ✅ Connector-specific events with typed enums
type EventType string
const (
    EventTypeMessage     EventType = "message"
    EventTypeSlashCommand EventType = "slash_command"
    EventTypeAppMention  EventType = "app_mention"
)

// ✅ Raw event preservation for extensibility
type MessageEvent struct {
    EventType EventType
    // Structured fields for common use cases
    TeamID    string
    ChannelID string
    Text      string
    // Raw data for advanced processing
    RawEvent  map[string]any
    CreatedAt time.Time
}
```

#### Type-Based Event Routing
```go
// ✅ Clean service-level routing
func (s *service) handleConnectorEvent(ctx context.Context, connectorType infragpt.ConnectorType, event any) error {
    switch connectorType {
    case infragpt.ConnectorTypeSlack:
        return s.handleSlackEvent(ctx, event)
    case infragpt.ConnectorTypeGithub:
        return s.handleGitHubEvent(ctx, event)
    }
}

func (s *service) handleSlackEvent(ctx context.Context, event any) error {
    switch e := event.(type) {
    case slack.MessageEvent:
        return s.handleSlackMessage(ctx, e)
    case slack.CommandEvent:
        return s.handleSlackCommand(ctx, e)
    }
}
```

### Database Design Patterns

#### Schema Design Principles
```sql
-- ✅ Single integration constraint per organization/connector
CREATE TABLE integrations (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    connector_type VARCHAR(50) NOT NULL,
    UNIQUE(organization_id, connector_type)  -- Key constraint
);

-- ✅ Optimized indexing for query patterns
CREATE INDEX idx_integrations_org ON integrations (organization_id);
CREATE INDEX idx_integrations_org_type ON integrations (organization_id, connector_type);

-- ✅ Proper foreign key relationships
CREATE TABLE integration_credentials (
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    UNIQUE(integration_id)  -- One credential set per integration
);
```

#### Repository Patterns
```go
// ✅ Descriptive method names that express business intent
type IntegrationRepository interface {
    Store(ctx context.Context, integration Integration) error
    FindByOrganization(ctx context.Context, orgID string) ([]Integration, error)
    FindByOrganizationAndType(ctx context.Context, orgID string, connectorType ConnectorType) ([]Integration, error)
    UpdateLastUsed(ctx context.Context, id string) error  // Specific business operation
}
```

### Concurrency & Error Handling Patterns

#### Concurrent Processing
```go
// ✅ errgroup for coordinated goroutine management
func (s *service) Subscribe(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)
    
    // Each connector in its own goroutine
    for connectorType, connector := range s.connectors {
        connector := connector  // Capture loop variable
        connectorType := connectorType
        
        g.Go(func() error {
            return connector.Subscribe(ctx, func(ctx context.Context, event any) error {
                return s.handleConnectorEvent(ctx, connectorType, event)
            })
        })
    }
    
    return g.Wait()  // Proper error aggregation
}
```

#### Error Context Preservation
```go
// ✅ Always wrap errors with meaningful context
func (r *integrationRepository) Store(ctx context.Context, integration Integration) error {
    if err := r.queries.StoreIntegration(ctx, params); err != nil {
        return fmt.Errorf("failed to store integration for org %s: %w", 
            integration.OrganizationID, err)
    }
    return nil
}

// ✅ Structured logging with relevant context
slog.Error("integration event processing failed",
    "connector_type", connectorType,
    "event_type", eventType,
    "organization_id", orgID,
    "error", err)
```

### Security Implementation Patterns

#### Encryption Standards
```go
// ✅ Production-ready encryption with versioning
type encryptionService struct {
    keyID string    // For key rotation support
}

func (e *encryptionService) Encrypt(data []byte) ([]byte, error) {
    // AES-256-GCM with unique nonce per operation
    nonce := make([]byte, 12)  // GCM standard nonce size
    if _, err := rand.Read(nonce); err != nil {
        return nil, fmt.Errorf("failed to generate nonce: %w", err)
    }
    // ... encryption implementation
}
```

#### Webhook Security
```go
// ✅ Timing-safe signature validation
func validateGitHubSignature(payload []byte, signature string, secret string) bool {
    expectedHash := computeHMAC(payload, secret)
    return hmac.Equal([]byte(signature), []byte(expectedHash))  // Constant-time comparison
}

// ✅ Panic recovery middleware
func panicMiddleware(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if r := recover(); r != nil {
                slog.Error("webhook panic", "recover", r, "url", r.URL.Path)
                w.WriteHeader(http.StatusInternalServerError)
            }
        }()
        h.ServeHTTP(w, r)
    })
}
```

### Implementation Learnings & Pattern Consistency

#### Pattern Consistency Achievements
1. **Clerk Pattern Adoption**: Following the exact Subscribe pattern from Clerk authentication ensured consistency across services
2. **Type-Based Event Routing**: Using Go's type switch for event routing proved clean and extensible
3. **Local Request/Response Types**: Keeping API types local to handlers maintains encapsulation

#### Event Architecture Decisions
1. **Connector-Specific Events**: Keeping events in connector packages eliminated complexity and over-engineering
2. **No Global Events**: Avoiding global event definitions kept the design simple and focused
3. **Raw Event Preservation**: Including raw event data in structs allows for advanced processing when needed

#### Configuration Management Learnings
1. **Mapstructure Integration**: Proper YAML parsing tags enabled clean configuration handling
2. **Conditional Registration**: Only registering connectors when configured prevents startup errors
3. **Environment-Based Config**: Using environment variables for sensitive data follows security best practices

#### Concurrent Processing Learnings
1. **Errgroup Usage**: Using errgroup for connector subscriptions provides proper error handling and graceful shutdown
2. **Goroutine Per Connector**: Each connector runs in its own goroutine for isolation
3. **Context Propagation**: Proper context handling enables clean cancellation

#### Error Handling Strategy
1. **Structured Logging**: Using slog with structured fields provides excellent debugging capability
2. **Graceful Degradation**: Connector failures don't bring down the entire system
3. **Context-Rich Errors**: All errors include meaningful context for troubleshooting

### Production Readiness Patterns

#### Monitoring & Observability
```go
// ✅ Structured logging with relevant context
slog.Info("integration event processed",
    "connector_type", connectorType,
    "event_type", eventType,
    "organization_id", orgID,
    "processing_time_ms", time.Since(start).Milliseconds())

// ✅ Error logging with troubleshooting context
slog.Error("credential validation failed",
    "connector_type", connectorType,
    "integration_id", integrationID,
    "error", err,
    "retry_attempt", retryCount)
```

#### Health Check Implementation
```go
// ✅ Connector health validation
func (s *service) validateConnectorHealth(ctx context.Context, integration Integration) error {
    connector, exists := s.connectors[integration.ConnectorType]
    if !exists {
        return fmt.Errorf("connector not available: %s", integration.ConnectorType)
    }
    
    credentials, err := s.credentialRepository.FindByIntegration(ctx, integration.ID)
    if err != nil {
        return fmt.Errorf("failed to retrieve credentials: %w", err)
    }
    
    return connector.ValidateCredentials(credentials.ToInfraGPTCredentials())
}
```

#### Performance Optimization
```go
// ✅ Connection reuse for HTTP clients
type slackConnector struct {
    config Config
    client *http.Client  // Reused across requests
}

func NewSlackConnector(config Config) domain.Connector {
    return &slackConnector{
        config: config,
        client: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
            },
        },
    }
}
```

### Code Quality Standards (Refined)

#### Variable and Function Naming
```go
// ✅ Descriptive names that explain purpose
func (s *service) handleSlackMessageEvent(ctx context.Context, event slack.MessageEvent) error
func (r *integrationRepository) FindByOrganizationAndType(ctx context.Context, orgID string, connectorType ConnectorType) ([]Integration, error)

// ❌ Avoid abbreviations and unclear names  
func (s *service) handleMsg(ctx context.Context, event any) error
func (r *repo) FindByOrgAndType(ctx context.Context, id string, t string) ([]Integration, error)
```

#### Interface Design
```go
// ✅ Focused interfaces with clear responsibilities
type IntegrationRepository interface {
    Store(ctx context.Context, integration Integration) error
    FindByID(ctx context.Context, id string) (Integration, error)
    FindByOrganization(ctx context.Context, orgID string) ([]Integration, error)
    Delete(ctx context.Context, id string) error
}

type CredentialRepository interface {
    Store(ctx context.Context, cred IntegrationCredential) error
    FindByIntegration(ctx context.Context, integrationID string) (IntegrationCredential, error)
    Update(ctx context.Context, cred IntegrationCredential) error
    Delete(ctx context.Context, integrationID string) error
}
```

#### Package Organization
```go
// ✅ Clear package structure with single responsibility
internal/integrationsvc/
├── domain/                    # Business logic interfaces
│   ├── integration.go         # Integration domain model
│   ├── credential.go          # Credential domain model  
│   └── connector.go           # Connector interface
├── service.go                 # Service implementation (no domain logic)
├── config.go                  # Service configuration
├── supporting/                # Infrastructure implementations
│   └── postgres/              # Database layer
│       ├── integration_repository.go
│       └── credential_repository.go
└── connectors/                # Connector implementations
    ├── slack/                 # Self-contained connector
    │   ├── config.go
    │   ├── slack.go
    │   └── events.go
    └── github/
        ├── config.go
        ├── github.go
        ├── events.go
        └── webhook.go
```

### Anti-Patterns to Avoid

#### Common Mistakes
```go
// ❌ Global event definitions
package infragpt
type SlackMessageEvent struct { ... }  // Wrong - creates coupling

// ✅ Connector-specific events
package slack
type MessageEvent struct { ... }       // Right - encapsulated

// ❌ String constants for types
const SlackConnector = "slack"         // Wrong - not type-safe

// ✅ Typed enums
type ConnectorType string
const ConnectorTypeSlack ConnectorType = "slack"  // Right - type-safe

// ❌ Interface{} everywhere
func handleEvent(event interface{}) error  // Wrong - not descriptive

// ✅ Use any with context
func handleEvent(ctx context.Context, event any) error  // Right - modern Go
```

#### Architectural Anti-Patterns
```go
// ❌ Centralized routing for all connectors
type WebhookRouter struct {
    slackHandler   SlackHandler
    githubHandler  GithubHandler
    // Creates coupling and complexity
}

// ✅ Connector ownership with Subscribe pattern
type Connector interface {
    Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error
    // Each connector handles its own communication
}
```

## Integration System Guidelines

### Connector Architecture

#### 1. Connector Interface
All connectors implement the standard interface:

```go
type Connector interface {
    // Authorization workflow
    InitiateAuthorization(organizationID string, userID string) (AuthorizationIntent, error)
    CompleteAuthorization(authData AuthorizationData) (Credentials, error)
    
    // Credential management
    ValidateCredentials(creds Credentials) error
    RefreshCredentials(creds Credentials) (Credentials, error)
    RevokeCredentials(creds Credentials) error
    
    // Event subscription
    Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error
}
```

#### 2. Configuration Pattern
Each connector uses the factory pattern:

```go
// connectors/slack/config.go
type Config struct {
    ClientID      string   `mapstructure:"client_id"`
    ClientSecret  string   `mapstructure:"client_secret"`
    BotToken      string   `mapstructure:"bot_token"`
    AppToken      string   `mapstructure:"app_token"`
}

func (c Config) NewConnector() domain.Connector {
    return &slackConnector{
        config: c,
        client: &http.Client{Timeout: 30 * time.Second},
    }
}
```

#### 3. Event Handling
- **Connector-specific events** stay within connector packages
- **Type-based routing** in service layer
- **Raw event preservation** for extensibility

```go
// Slack events
type MessageEvent struct {
    EventType EventType
    TeamID    string
    ChannelID string
    RawEvent  map[string]any
    CreatedAt time.Time
}

// GitHub events
type WebhookEvent struct {
    EventType      EventType
    InstallationID int64
    RepositoryName string
    RawPayload     map[string]any
    CreatedAt      time.Time
}
```

### Adding New Connectors

1. **Create connector package**: `internal/integrationsvc/connectors/newconnector/`
2. **Implement required files**:
   - `config.go` - Configuration with factory method
   - `connector.go` - Main connector implementation
   - `events.go` - Connector-specific event types
   - `webhook.go` - HTTP server (if webhook-based)

3. **Register in service**: Add to configuration and service setup
4. **Add tests**: Unit and integration tests for all methods
5. **Update documentation**: API endpoints and configuration

### Communication Patterns

#### Socket-Based (e.g., Slack)
```go
func (s *slackConnector) Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error {
    // WebSocket connection for real-time events
    client := socketmode.New(slack.New(s.config.BotToken))
    return client.Run()
}
```

#### Webhook-Based (e.g., GitHub)
```go
func (g *githubConnector) Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error {
    // Dedicated HTTP server with signature validation
    return g.startWebhookServer(ctx, handler)
}
```

## Pull Request Process

### PR Requirements

1. **Descriptive title** summarizing the change
2. **Detailed description** explaining the purpose and approach
3. **Test coverage** for new functionality
4. **Documentation updates** if needed
5. **No breaking changes** without discussion

### PR Template

```markdown
## Description
Brief description of the changes made.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests pass locally
```

### Review Process

1. **Automated checks** must pass (tests, linting)
2. **Code review** by at least one maintainer
3. **Security review** for sensitive changes
4. **Documentation review** for user-facing changes
5. **Final approval** and merge by maintainer

### Code Review Focus Areas

- **Architecture alignment** with established patterns
- **Security considerations** and best practices
- **Error handling** and edge cases
- **Test coverage** and quality
- **Performance implications**
- **Documentation completeness**

## Development Tools

### Required Tools

```bash
# Install development dependencies
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Code Generation

```bash
# Generate database code
sqlc generate

# Generate protobuf code
protoc --go_out=. --go-grpc_out=. proto/*.proto
```

### Useful Commands

```bash
# Format code
go fmt ./...
goimports -w .

# Lint code
go vet ./...

# Check for common issues
staticcheck ./...

# Update dependencies
go mod tidy
go mod download
```

## Performance Considerations

### Database Optimization

- **Proper indexing** for all query patterns
- **Connection pooling** for high-throughput scenarios
- **Query optimization** with EXPLAIN ANALYZE
- **Transaction management** for data consistency

### Concurrent Processing

- **errgroup** for coordinated goroutine management
- **Context cancellation** for graceful shutdown
- **Rate limiting** for external API calls
- **Circuit breakers** for resilience

### Memory Management

- **Stream processing** for large datasets
- **Connection reuse** for HTTP clients
- **Proper resource cleanup** with defer statements
- **Garbage collection** awareness in hot paths

---

## Getting Help

- **Questions**: Open a GitHub discussion
- **Bug reports**: Create a detailed GitHub issue
- **Feature requests**: Submit a GitHub issue with use case
- **Security issues**: Follow responsible disclosure process

We appreciate your contributions and look forward to building InfraGPT together!