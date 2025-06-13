# InfraGPT Core Service

InfraGPT is a Go-based Slack bot service that provides AI-powered infrastructure management through intelligent DevOps workflows. The service integrates with external AI agents and provides identity management with Clerk authentication.

## Architecture Overview

The service follows clean architecture principles with clear separation of concerns:

### Core Services

#### 1. InfraGPT Service (`internal/infragptsvc/`)
- **Purpose**: Main Slack bot functionality and infrastructure management
- **Key Features**: 
  - Slack Socket Mode integration for real-time messaging
  - Conversation management and threading
  - Integration with external AI agents via gRPC
  - Channel and workspace management

#### 2. Identity Service (`internal/identitysvc/`)
- **Purpose**: User authentication and organization management
- **Key Features**:
  - Clerk webhook integration for user/organization sync
  - Organization onboarding workflow
  - Metadata collection (company size, team size, use cases, observability stack)
  - Member management within organizations

#### 3. Integration Service (`internal/integrationsvc/`)
- **Purpose**: External service integration management with connector pattern
- **Key Features**:
  - Multi-connector architecture (Slack, GitHub, AWS, GCP, PagerDuty, Datadog)
  - OAuth2 and installation-based authentication flows
  - AES-256-GCM encrypted credential storage
  - Subscribe pattern for real-time event handling
  - Dedicated webhook servers and Socket Mode support

### Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                        API Layer                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌───────────────┐ │
│  │  HTTP REST API  │  │   gRPC API      │  │  Webhooks     │ │
│  └─────────────────┘  └─────────────────┘  └───────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│  ┌─────────────────┐  ┌─────────────────────────────────────┐ │
│  │ InfraGPT Service│  │      Identity Service               │ │
│  └─────────────────┘  └─────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                       │
│  ┌──────────┐ ┌──────┐ ┌───────┐ ┌───────┐ ┌──────┐ ┌──────┐ │
│  │PostgreSQL│ │Slack │ │GitHub │ │ Agent │ │Clerk │ │Other │ │
│  │ Database │ │Socket│ │Webhook│ │ gRPC  │ │ Auth │ │ APIs │ │
│  └──────────┘ └──────┘ └───────┘ └───────┘ └──────┘ └──────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
├── cmd/main.go                    # Application entry point
├── config.yaml                    # Configuration file
├── integration.go                 # Integration service interface  
├── identity.go                    # Identity service interface
├── spec.go                        # Main service interface
│
├── infragptapi/                   # HTTP/gRPC API layer
│   ├── handler.go                 # HTTP route handlers
│   ├── grpc_server.go            # gRPC server implementation
│   └── proto/                     # Protocol buffer definitions
│
├── identityapi/                   # Identity API handlers
│   └── handler.go                 # Identity HTTP endpoints
│
├── integrationapi/                # Integration API handlers
│   └── handler.go                 # Integration HTTP endpoints
│
├── internal/
│   ├── generic/                   # Shared utilities
│   │   ├── httperrors/           # HTTP error handling
│   │   ├── httplog/              # HTTP request logging middleware
│   │   └── postgresconfig/       # Database configuration
│   │
│   ├── conversationsvc/           # Conversation service implementation
│   │   ├── domain/               # Conversation domain models
│   │   ├── service.go            # Service implementation
│   │   └── supporting/           # Infrastructure implementations
│   │       ├── agent/            # Agent client integration
│   │       ├── postgres/         # Database repositories
│   │       └── slack/            # Slack API integration
│   │
│   ├── identitysvc/              # Identity service implementation
│   │   ├── domain/               # Identity domain models
│   │   ├── service.go            # Service implementation
│   │   └── supporting/           # Infrastructure implementations
│   │       ├── clerk/            # Clerk integration
│   │       └── postgres/         # Database repositories
│   │
│   └── integrationsvc/           # Integration service implementation
│       ├── domain/               # Integration domain models
│       ├── service.go            # Service implementation
│       ├── config.go             # Service configuration
│       ├── connectors/           # Connector implementations
│       │   ├── slack/            # Slack OAuth2 + Socket Mode
│       │   └── github/           # GitHub App + Webhooks
│       └── supporting/postgres/  # Database repositories
│
├── migrations/                    # Database migration scripts
└── docs/                         # Documentation
    ├── CONTRIBUTING.md            # Development guidelines
    ├── IMPLEMENTATION.md          # Implementation overview
    └── INTEGRATION_SYSTEM.md     # Integration system design
```

## Implementation Status

### ✅ Production Ready
- **Identity Service**: Complete Clerk authentication with organization management
- **Integration Service**: Full connector architecture with Slack and GitHub connectors
- **Database Layer**: Type-safe SQLC queries with optimized indexing
- **API Layer**: REST endpoints with authentication middleware
- **Security**: AES-256-GCM credential encryption and webhook signature validation
- **Concurrency**: errgroup-based goroutine management with graceful shutdown

### 🔧 Active Development  
- **Conversation Service**: Enhanced AI agent integration and context management
- **Business Logic**: Event-driven workflow automation based on external service events
- **Additional Connectors**: AWS, GCP, PagerDuty, and Datadog connector implementations

### 📋 Planned
- **Advanced Workflows**: Complex multi-step automation and approval flows
- **Analytics Dashboard**: Usage metrics and integration health monitoring
- **Plugin Architecture**: Custom connector development framework

## Key Features

### 1. Real-time Slack Integration
- **Socket Mode**: Real-time bi-directional communication with Slack
- **Thread Management**: Conversation threading and context preservation
- **Channel Subscriptions**: Automated responses to mentions and messages
- **Multi-Workspace**: Support for multiple Slack workspace connections
- **OAuth2 Flow**: Complete Slack app installation and token management

### 2. Identity & Authentication
- **Clerk Integration**: OAuth authentication with real-time webhook synchronization
- **Organization Management**: Multi-tenant organization structure with member management
- **Onboarding Flow**: Guided setup for new organizations with metadata collection
- **Session Management**: JWT token validation and secure authentication middleware

### 3. AI Agent Integration
- **gRPC Communication**: High-performance communication with AI agents
- **Conversation Context**: Maintains conversation state across interactions
- **Agent Registry**: Support for multiple specialized AI agents
- **Fallback Handling**: Graceful degradation when agents are unavailable
- **Message Processing**: Real-time message analysis and response generation

### 4. External Service Integration
- **Multi-Connector Architecture**: Support for 6+ external services (Slack, GitHub, AWS, GCP, PagerDuty, Datadog)
- **OAuth2 & Installation Flows**: Complete authorization workflows with secure credential storage
- **Real-time Event Processing**: Socket Mode (Slack) and webhook servers (GitHub, etc.)
- **AES-256-GCM Encryption**: Production-ready credential encryption with key rotation support
- **Subscribe Pattern**: Consistent event handling across all connectors

### 5. HTTP Logging Middleware
- **Colorful Output**: ANSI color-coded logs for different HTTP methods and status codes
- **Request/Response Tracking**: Complete request lifecycle logging
- **Configurable**: Can be enabled/disabled via boolean parameter
- **Performance Metrics**: Request duration and response size tracking

## Configuration

The service is configured via `config.yaml`:

```yaml
port: 8080                # HTTP server port
grpc_port: 9090          # gRPC server port

slack:
  client_id: "..."       # Slack app client ID
  client_secret: "..."   # Slack app client secret  
  app_token: "..."       # Slack app token for Socket Mode

database:
  host: "localhost"      # PostgreSQL host
  port: 5432            # PostgreSQL port
  db_name: "infragpt"   # Database name
  user: "infragpt"      # Database user
  password: "..."       # Database password

agent:
  endpoint: "[::]:50051" # Agent service gRPC endpoint
  retry_attempts: 3      # Connection retry attempts

identity:
  clerk:
    webhook_secret: "..." # Clerk webhook signing secret
    publishable_key: "..." # Clerk publishable key
    port: 8082            # Webhook server port

integrations:
  slack:
    client_id: "..."      # Slack app client ID
    client_secret: "..."  # Slack app client secret
    bot_token: "..."      # Bot token for Socket Mode
    app_token: "..."      # App token for Socket Mode
    redirect_url: "..."   # OAuth callback URL
    signing_secret: "..." # Webhook signature validation
    
  github:
    app_id: "..."         # GitHub App ID
    private_key: "..."    # GitHub App private key (PEM)
    webhook_secret: "..." # Webhook signature validation
    webhook_port: 8081    # Dedicated webhook server port
    
  # Additional connectors: GCP, AWS, PagerDuty, Datadog
  # See docs/INTEGRATION_SYSTEM.md for complete configuration
```

## Current Capabilities

### Live Features
- **Multi-Service Architecture**: Three production-ready services working together
- **Slack Bot Integration**: Real-time message processing with Socket Mode
- **External Service Connections**: OAuth2 and installation-based authentication flows
- **Secure Credential Storage**: AES-256-GCM encryption with environment-based key management
- **Event-Driven Processing**: Real-time webhook and socket-based event handling
- **Multi-Tenant Organizations**: Complete organization and user management
- **Type-Safe Database**: SQLC-generated queries with optimized indexing
- **Production Security**: Signature validation, panic recovery, and audit logging

### Integration Capabilities
- **Slack**: Complete OAuth2 + Socket Mode integration (Production Ready)
- **GitHub**: GitHub App installation + webhook events (Production Ready)
- **Identity Management**: Clerk authentication and organization sync (Production Ready)
- **AI Agents**: gRPC communication with conversation context (Active Development)

## Development Commands

### Build and Run
```bash
go run ./cmd/main.go          # Run with config.yaml
go build ./cmd/main.go        # Build binary
```

### Testing
```bash
go test ./...                           # Run all tests
go test ./internal/identitysvc/...      # Run identity service tests
go test ./internal/integrationsvc/...   # Run integration service tests
go test ./internal/conversationsvc/...  # Run conversation service tests
go vet ./...                            # Static analysis
```

### Database Management
```bash
sqlc generate                 # Generate Go code from SQL queries
```

### Dependencies
```bash
go mod tidy                   # Clean up dependencies
go mod download               # Download dependencies
```

## Database Schema

### Identity Tables
- **users**: User accounts synced from Clerk
- **organizations**: Organization entities with Clerk integration
- **organization_metadata**: Extended organization information
- **organization_members**: User-organization relationships

### InfraGPT Tables  
- **conversations**: Message thread management
- **channels**: Slack channel subscriptions
- **messages**: Message history and context

### Integration Tables
- **integrations**: External service integrations (Slack, GitHub, etc.)
- **integration_credentials**: Encrypted credential storage with AES-256-GCM

## API Endpoints

### Identity API
- `POST /webhooks/clerk` - Clerk webhook receiver
- `POST /api/v1/organizations/get` - Get organization details
- `POST /api/v1/organizations/metadata/set` - Set organization metadata

### Integration API
- `POST /integrations/authorize/` - Initiate OAuth/installation flow
- `POST /integrations/callback/` - Handle OAuth callbacks
- `POST /integrations/list/` - List organization integrations
- `POST /integrations/revoke/` - Revoke integration
- `POST /integrations/status/` - Integration health check

### InfraGPT API
- `GET /slack` - Complete Slack OAuth flow
- `POST /reply` - Send message replies

### gRPC API
- `SendMessage` - Process incoming messages
- `GetConversation` - Retrieve conversation context

## Security

- **Webhook Verification**: Svix signature validation for Clerk webhooks
- **Token Validation**: Clerk JWT token authentication for protected endpoints
- **CORS Protection**: Configurable cross-origin request handling
- **Panic Recovery**: Graceful error handling with request recovery

## Monitoring & Observability

- **Structured Logging**: JSON-formatted logs with contextual information
- **HTTP Request Logging**: Colorful request/response logging with timing
- **Error Tracking**: Comprehensive error context and stack traces
- **Health Checks**: Service health monitoring endpoints

## Development Guidelines

### Code Style
- Follow standard Go conventions from Effective Go
- Use gofmt/goimports for code formatting
- CamelCase for exported symbols, camelCase for non-exported
- One package per directory, package name matches directory name

### Error Handling
- Always check errors and provide meaningful context
- Use custom error types for domain-specific errors
- Log errors with sufficient context for debugging

### Testing
- Use testcontainers-go for integration tests requiring PostgreSQL
- Create test utilities in `*test` packages
- Mock external dependencies for unit tests

### Database
- Use SQLC for type-safe SQL query generation
- Place schema files in `schema/` directories
- Place query files in `queries/` directories
- Follow migration naming convention: `001_description.sql`

## Dependencies

### Core Dependencies
- **Slack API**: `github.com/slack-go/slack` - Slack integration and Socket Mode
- **PostgreSQL**: `github.com/lib/pq` - Database driver with connection pooling
- **gRPC**: `google.golang.org/grpc` - AI agent communication
- **Configuration**: `gopkg.in/yaml.v3`, `github.com/mitchellh/mapstructure` - YAML config parsing
- **Concurrency**: `golang.org/x/sync/errgroup` - Coordinated goroutine management
- **HTTP**: Standard library `net/http` - API servers and webhook endpoints
- **Cryptography**: Standard library `crypto/*` - AES-256-GCM encryption and signature validation
- **UUID**: `github.com/google/uuid` - Unique identifier generation

### Development Dependencies
- **SQLC**: Type-safe SQL query generation with PostgreSQL support
- **Testcontainers**: Integration testing with real PostgreSQL containers
- **Testing**: Standard library `testing` with table-driven test patterns
- **Code Generation**: `protoc` and related tools for gRPC interface generation

## Integration System

The Integration Service provides a connector-based architecture for integrating with external services. For detailed design documentation, see [docs/INTEGRATION_SYSTEM.md](docs/INTEGRATION_SYSTEM.md).

### Supported Connectors
- **Slack**: OAuth2 + Socket Mode for real-time events
- **GitHub**: GitHub App installation with webhook events
- **GCP**: Service account credential management
- **AWS**: Access key + secret key authentication
- **PagerDuty**: API key authentication
- **Datadog**: API key + application key authentication

### Key Design Principles
- **Connector Ownership**: Each connector manages its own communication method
- **Subscribe Pattern**: Unified event handling following Clerk authentication pattern
- **Secure Storage**: AES-256-GCM encryption for all credentials
- **Clean Architecture**: Domain-driven design with proper separation of concerns

## Contributing

See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for detailed contribution guidelines.

## License

This project is part of the InfraGPT platform and follows the project's licensing terms.