# Integration System Design

## Overview

The InfraGPT Integration System enables secure connectivity with external services through a connector-based architecture. It supports multiple authentication methods, encrypted credential storage, and real-time event handling.

## Architecture

### Connector Pattern

Each connector implements a standardized interface supporting:
- **Authorization flows**: OAuth2, GitHub App installations, API keys
- **Credential management**: Secure storage, validation, refresh, and revocation
- **Event subscription**: Real-time event handling via Subscribe pattern

### Communication Patterns

**Socket-Based Connectors** (e.g., Slack)
- Persistent WebSocket connections
- Real-time event streaming
- Automatic reconnection handling

**HTTP Webhook Connectors** (e.g., GitHub, PagerDuty)
- Dedicated HTTP servers per connector
- Signature verification for security
- Event conversion to domain objects

## Supported Connectors

| Connector | Auth Type | Events | Description |
|-----------|-----------|---------|-------------|
| **Slack** | OAuth2 + Bot Token | Socket Mode | Real-time messaging and commands |
| **GitHub** | App Installation | Webhooks | Repository events and PR management |
| **GCP** | Service Account | N/A | Cloud resource management |
| **AWS** | Access Key | N/A | Infrastructure provisioning |
| **PagerDuty** | API Key | Webhooks | Incident management |
| **Datadog** | API Key | Webhooks | Monitoring and observability |

## Core Interfaces

### Connector Interface

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

### Service Interface

```go
type IntegrationService interface {
    // Commands
    NewIntegration(ctx context.Context, cmd NewIntegrationCommand) (IntegrationAuthorizationIntent, error)
    AuthorizeIntegration(ctx context.Context, cmd AuthorizeIntegrationCommand) (Integration, error)
    RevokeIntegration(ctx context.Context, cmd RevokeIntegrationCommand) error
    
    // Queries
    Integrations(ctx context.Context, query IntegrationsQuery) ([]Integration, error)
    Integration(ctx context.Context, query IntegrationQuery) (Integration, error)
    
    // Event subscription
    Subscribe(ctx context.Context) error
}
```

## Security Architecture

### Credential Encryption
- **Algorithm**: AES-256-GCM
- **Key Management**: Environment-based derivation with rotation support
- **Storage**: All sensitive data encrypted before database persistence

### Webhook Security
- **Slack**: HMAC-SHA256 signature validation with timing safety
- **GitHub**: SHA256 signature verification with constant-time comparison
- **Middleware**: Panic recovery and request validation

## Event Handling

### Subscribe Pattern

Following the Clerk authentication service pattern:

```go
// Slack - Socket Mode
func (s *slackConnector) Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error {
    // Socket Mode client with real-time event streaming
}

// GitHub - Dedicated Webhook Server
func (g *githubConnector) Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error {
    // HTTP server on dedicated port with signature validation
}
```

### Event Types

**Slack Events**
- `MessageEvent`: Direct messages, mentions, reactions
- Event types: `message`, `slash_command`, `app_mention`, `reaction`

**GitHub Events**
- `WebhookEvent`: Repository events, pull requests, installations
- Event types: `push`, `pull_request`, `installation`, `issues`

## Database Schema

### Core Tables

```sql
-- Integration metadata
CREATE TABLE integrations (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    connector_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, connector_type)
);

-- Encrypted credential storage
CREATE TABLE integration_credentials (
    id UUID PRIMARY KEY,
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    credential_type VARCHAR(50) NOT NULL,
    credential_data_encrypted TEXT NOT NULL,
    expires_at TIMESTAMP,
    encryption_key_id VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(integration_id)
);
```

## API Endpoints

All endpoints use POST methods for consistency:

```
POST /integrations/authorize/     # Initiate authorization
POST /integrations/callback/      # Handle OAuth callbacks  
POST /integrations/list/          # List integrations
POST /integrations/revoke/        # Revoke integration
POST /integrations/status/        # Health check
```

## Configuration

```yaml
integrations:
  slack:
    client_id: "${SLACK_CLIENT_ID}"
    client_secret: "${SLACK_CLIENT_SECRET}"
    bot_token: "${SLACK_BOT_TOKEN}"        # Socket Mode
    app_token: "${SLACK_APP_TOKEN}"        # Socket Mode
    
  github:
    app_id: "${GITHUB_APP_ID}"
    private_key: "${GITHUB_PRIVATE_KEY}"
    webhook_secret: "${GITHUB_WEBHOOK_SECRET}"
    webhook_port: 8081                     # Dedicated server
```

## Design Principles

### 1. Connector Ownership
- Each connector manages its own communication method
- No centralized routing or shared infrastructure
- Complete isolation between connector types

### 2. Clean Architecture
- Domain-driven design with proper layer separation
- Repository pattern for data persistence
- Command/Query structure for service operations

### 3. Subscribe Pattern Consistency
- Follows exact same pattern as Clerk authentication service
- Unified event handling across all connector types
- Context-based cancellation and error handling

### 4. Security First
- Encrypted credential storage with key rotation support
- Signature validation for all webhook endpoints
- No plaintext secrets in logs or database

## Development Guidelines

### Code Style
- Use `any` instead of `interface{}` for modern Go practices
- JSON tags only on API boundary structs (request/response)
- Self-documenting code with minimal comments
- Clean internal domain structures without serialization concerns

### Error Handling
- Context wrapping with `fmt.Errorf` for traceability
- Structured logging with relevant fields
- Graceful degradation when connectors fail

### Testing
- Integration tests using testcontainers for PostgreSQL
- Mock external dependencies for unit tests
- Webhook signature validation testing

## Connector Implementation

### Adding New Connectors

1. **Create connector package**: `internal/integrationsvc/connectors/newconnector/`
2. **Implement Connector interface**: Authorization, credential management, subscription
3. **Add configuration**: Config struct with mapstructure tags
4. **Register in service**: Conditional registration based on config
5. **Add event types**: Connector-specific event structures

### Example Structure

```
internal/integrationsvc/connectors/newconnector/
├── config.go     # Configuration with NewConnector() factory
├── events.go     # Connector-specific event types
├── connector.go  # Connector interface implementation  
└── webhook.go    # HTTP server (if webhook-based)
```

## Production Considerations

### Monitoring
- Health checks for credential validation
- Event processing metrics and error rates
- Connection monitoring for Socket Mode connectors

### Scaling
- Dedicated webhook servers prevent port conflicts
- Concurrent event processing with proper error isolation
- Database connection pooling for high-throughput scenarios

### Maintenance
- Credential refresh automation
- Event replay capabilities for failed processing
- Audit logging for compliance requirements