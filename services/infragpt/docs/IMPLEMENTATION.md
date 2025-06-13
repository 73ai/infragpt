# InfraGPT Implementation Overview

## Service Architecture

InfraGPT consists of three main services that work together to provide AI-powered infrastructure management through Slack:

### 1. InfraGPT Core Service
- **Main service coordination**: Message processing and workflow orchestration
- **Slack integration**: Socket Mode for real-time messaging
- **Agent communication**: gRPC integration with AI agents
- **Database**: PostgreSQL with conversation and channel management

### 2. Identity Service  
- **User management**: Clerk integration for authentication
- **Organization management**: Multi-tenant organization structure
- **Webhook handling**: Real-time user/organization synchronization

### 3. Integration Service
- **External service connectivity**: Multi-connector architecture
- **Credential management**: Encrypted storage with AES-256-GCM
- **Event handling**: Real-time event processing via Subscribe pattern
- **API management**: REST endpoints for integration lifecycle

## Current Implementation Status

### ✅ Production Ready
- **Core Service**: Slack Socket Mode integration with conversation threading
- **Identity Service**: Complete Clerk authentication with organization management  
- **Integration Service**: Full connector architecture with Slack and GitHub connectors
- **Database Layer**: Type-safe SQLC queries with proper indexing
- **API Layer**: REST endpoints with authentication middleware
- **Security**: Encrypted credential storage and webhook signature validation

### 🔧 Active Development
- **AI Agent Integration**: Enhanced conversation context and response generation
- **Business Logic**: Event-driven workflow automation
- **Monitoring**: Health checks and observability improvements

## Technology Stack

### Core Technologies
- **Language**: Go 1.24+
- **Database**: PostgreSQL with SQLC for type-safe queries
- **Configuration**: YAML with mapstructure parsing
- **Concurrency**: errgroup for coordinated goroutine management

### External Integrations
- **Slack**: Socket Mode for real-time events
- **Clerk**: Authentication and user management
- **Agent Service**: gRPC communication with AI agents
- **GitHub**: App installation and webhook events

### Development Tools
- **Testing**: testcontainers for integration tests
- **Code Generation**: SQLC for database queries, protobuf for gRPC
- **Validation**: Webhook signature verification (HMAC-SHA256, SHA256)

## Key Implementation Patterns

### Clean Architecture
```
┌─────────────────────────────────────────┐
│                API Layer                 │
│  HTTP Handlers │ gRPC Server │ Webhooks │
└─────────────────────────────────────────┘
┌─────────────────────────────────────────┐
│            Application Layer             │
│     Service Implementations             │
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

### Subscribe Pattern
Following Clerk authentication service pattern for consistent event handling:

```go
type Service interface {
    Subscribe(ctx context.Context) error
}

// Each service/connector implements Subscribe for event handling
func (s *service) Subscribe(ctx context.Context) error {
    // Start event processing in goroutines
    // Handle graceful shutdown with context cancellation
}
```

### Repository Pattern
Domain interfaces with infrastructure implementations:

```go
// Domain interface
type IntegrationRepository interface {
    Store(ctx context.Context, integration Integration) error
    FindByOrganization(ctx context.Context, orgID string) ([]Integration, error)
}

// PostgreSQL implementation
type integrationRepository struct {
    queries *Queries
}
```

## Configuration Structure

### Main Configuration (`config.yaml`)
```yaml
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

agent:
  endpoint: "[::]:50051"

identity:
  clerk:
    webhook_secret: "${CLERK_WEBHOOK_SECRET}"
    
integrations:
  slack:
    client_id: "${SLACK_CLIENT_ID}"
    bot_token: "${SLACK_BOT_TOKEN}"
  github:
    app_id: "${GITHUB_APP_ID}"
    webhook_port: 8081
```

## Database Schema

### Service-Specific Tables

**Identity Service**
- `users`: Clerk user synchronization
- `organizations`: Organization entities
- `organization_members`: User-organization relationships
- `organization_metadata`: Extended organization data

**Integration Service**  
- `integrations`: External service connections
- `integration_credentials`: Encrypted credential storage

**Core Service**
- `conversations`: Message thread management
- `channels`: Slack channel subscriptions
- `messages`: Message history and context

## Development Workflow

### Running the Service
```bash
# Development
go run ./cmd/main.go

# Build
go build ./cmd/main.go

# Testing  
go test ./...
go test ./internal/integrationsvc/...
```

### Database Operations
```bash
# Generate code from SQL
sqlc generate

# Run migrations (handled automatically on startup)
```

### Adding New Features

1. **Define domain interface** in appropriate `domain/` package
2. **Implement service logic** in service implementation
3. **Add infrastructure support** in `supporting/` packages
4. **Create API endpoints** in API handler packages
5. **Add configuration** to config structures
6. **Wire in main.go** with proper dependency injection

## Security Implementation

### Credential Encryption
- AES-256-GCM encryption for all sensitive data
- Environment-based key derivation with rotation support
- Unique nonce per encryption operation

### Webhook Security
- Signature validation for all webhook endpoints
- Timing-safe comparison to prevent timing attacks
- Panic recovery middleware for production safety

### Access Control
- Organization-scoped access for all operations
- Clerk JWT token validation for protected endpoints
- Audit logging for sensitive operations

## Monitoring & Observability

### Logging
- Structured JSON logging with contextual information
- Request/response logging with colorful output
- Error tracking with proper context wrapping

### Health Checks
- Service startup validation
- Database connection monitoring
- External service connectivity checks

### Metrics (Planned)
- Request duration and error rates
- Event processing throughput
- Credential refresh success rates

## Future Enhancements

### Near-term
- Enhanced AI agent integration with conversation context
- Advanced Slack interaction components (buttons, modals)
- Automated credential refresh workflows

### Long-term
- Multi-region deployment support
- Advanced analytics and usage tracking
- Plugin architecture for custom connectors