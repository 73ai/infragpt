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
│  ┌───────────────┐ ┌────────────┐ ┌────────────┐ ┌─────────┐ │
│  │   PostgreSQL  │ │   Slack    │ │   Agent    │ │  Clerk  │ │
│  │   Database    │ │    API     │ │   gRPC     │ │ Webhooks│ │
│  └───────────────┘ └────────────┘ └────────────┘ └─────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
├── cmd/main.go                    # Application entry point
├── config.yaml                    # Configuration file
├── identity.go                    # Identity service interface
├── spec.go                        # Main service interface
│
├── infragptapi/                   # HTTP/gRPC API layer
│   ├── handler.go                 # HTTP route handlers
│   ├── grpc_server.go            # gRPC server implementation
│   └── proto/                     # Protocol buffer definitions
│
├── identityapi/                   # Identity API handlers
│   ├── handler.go                 # Identity HTTP endpoints
│   ├── middleware.go              # Authentication middleware
│   └── webhooks.go                # Clerk webhook validation
│
├── internal/
│   ├── generic/                   # Shared utilities
│   │   ├── httperrors/           # HTTP error handling
│   │   ├── httplog/              # HTTP request logging middleware
│   │   └── postgresconfig/       # Database configuration
│   │
│   ├── infragptsvc/              # InfraGPT service implementation
│   │   ├── domain/               # Business domain models
│   │   ├── service.go            # Service implementation
│   │   └── supporting/           # Infrastructure implementations
│   │       ├── agent/            # Agent client integration
│   │       ├── postgres/         # Database repositories
│   │       └── slack/            # Slack API integration
│   │
│   └── identitysvc/              # Identity service implementation
│       ├── domain/               # Identity domain models
│       ├── service.go            # Service implementation
│       └── supporting/postgres/  # Database repositories
│
├── migrations/                    # Database migration scripts
└── docs/                         # Documentation
```

## Key Features

### 1. Slack Integration
- **Socket Mode**: Real-time bi-directional communication with Slack
- **Thread Management**: Conversation threading and context preservation
- **Channel Subscriptions**: Automated responses to mentions and messages
- **Workspace Management**: Multi-workspace support with token management

### 2. Identity & Authentication
- **Clerk Integration**: OAuth authentication with webhook synchronization
- **Organization Management**: Multi-tenant organization structure
- **Onboarding Flow**: Guided setup for new organizations
- **Metadata Collection**: Company profile and use case tracking

### 3. AI Agent Integration
- **gRPC Communication**: High-performance communication with AI agents
- **Conversation Context**: Maintains conversation state across interactions
- **Agent Registry**: Support for multiple specialized AI agents
- **Fallback Handling**: Graceful degradation when agents are unavailable

### 4. HTTP Logging Middleware
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
```

## Development Commands

### Build and Run
```bash
go run ./cmd/main.go          # Run with config.yaml
go build ./cmd/main.go        # Build binary
```

### Testing
```bash
go test ./...                 # Run all tests
go test ./internal/identitysvc/... # Run identity service tests
go vet ./...                  # Static analysis
```

### Database Management
```bash
sqlc generate                 # Generate Go code from SQL queries
./migrations/run_migration.sh # Run database migrations
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
- **integrations**: Slack workspace integrations
- **conversations**: Message thread management
- **channels**: Slack channel subscriptions
- **messages**: Message history and context

## API Endpoints

### Identity API
- `POST /webhooks/clerk` - Clerk webhook receiver
- `POST /api/v1/organizations/get` - Get organization details
- `POST /api/v1/organizations/metadata/set` - Set organization metadata

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
- **Slack API**: `github.com/slack-go/slack`
- **PostgreSQL**: `github.com/lib/pq`
- **gRPC**: `google.golang.org/grpc`
- **Configuration**: `gopkg.in/yaml.v3`, `github.com/mitchellh/mapstructure`

### Development Dependencies
- **SQLC**: SQL to Go code generation
- **Testcontainers**: Integration testing with real PostgreSQL
- **Errgroup**: Concurrent goroutine management

## Contributing

See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for detailed contribution guidelines.

## License

This project is part of the InfraGPT platform and follows the project's licensing terms.