# Architecture Overview

A simple explanation of how InfraGPT is organized.

## The Big Picture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Slack Users    │    │   Web Client    │    │  External APIs  │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          v                      v                      v
┌─────────────────────────────────────────────────────────────────┐
│                    InfraGPT Backend                             │
├─────────────────┬─────────────────┬─────────────────────────────┤
│ Conversation    │ Identity        │ Integration                 │
│ Service         │ Service         │ Service                     │
│                 │                 │                             │
│ • Slack bot     │ • User mgmt     │ • GitHub connector          │
│ • AI agent      │ • Org mgmt      │ • Slack connector           │
│ • Chat flows    │ • Clerk webhooks│ • OAuth flows               │
└─────────────────┴─────────────────┴─────────────────────────────┘
                          │
                          v
                 ┌─────────────────┐
                 │   PostgreSQL    │
                 │                 │
                 │ • Conversations │
                 │ • Users/Orgs    │
                 │ • Integrations  │
                 └─────────────────┘
```

## Three Main Services

### 1. Conversation Service
**What it does**: Handles Slack messages and coordinates AI responses

**Key components**:
- Slack Socket Mode connection for real-time messages
- gRPC client to communicate with AI agents
- Conversation threading and context management

**Files to know**:
- `internal/conversationsvc/service.go` - Main service logic
- `internal/conversationsvc/supporting/slack/` - Slack integration

### 2. Identity Service
**What it does**: Manages users and organizations

**Key components**:
- Clerk webhook handler for user sync
- Organization membership management
- JWT token validation middleware

**Files to know**:
- `internal/identitysvc/service.go` - Main service logic
- `internal/identitysvc/supporting/clerk/` - Clerk integration

### 3. Integration Service
**What it does**: Connects to external services (GitHub, Slack, etc.)

**Key components**:
- OAuth authorization flows
- Encrypted credential storage
- Event handlers for webhooks/real-time events

**Files to know**:
- `internal/integrationsvc/service.go` - Main service logic
- `internal/integrationsvc/connectors/` - Individual connectors

## Code Organization

Each service follows the same pattern:

```
internal/servicename/
├── service.go              # Main service implementation
├── config.go               # Configuration and setup
├── domain/                 # Business logic interfaces
│   ├── models.go          # Core data types
│   └── interfaces.go      # Repository interfaces
└── supporting/            # Infrastructure implementations
    ├── postgres/          # Database layer
    └── external/          # External API clients
```

## How Services Talk to Each Other

**Database**: All services share a PostgreSQL connection but use separate schemas

**HTTP APIs**: Services expose REST endpoints for external clients

**gRPC**: Conversation service calls external AI agents via gRPC

**Events**: Services handle real-time events from external systems (Slack, GitHub, Clerk)

## Database Structure

**Three main areas**:
- **Conversations**: Messages, channels, thread context
- **Identity**: Users, organizations, memberships
- **Integrations**: External service connections and credentials

**Key pattern**: Uses SQLC for type-safe query generation from SQL files

## Key Design Patterns

**Repository Pattern**: Domain interfaces with PostgreSQL implementations

**Command/Query**: Service methods are either commands (change state) or queries (read data)

**Subscribe Pattern**: All services implement `Subscribe(ctx)` for event handling

**Factory Pattern**: Config structs have `New()` methods to create services

## API Structure

**HTTP Endpoints**:
- `/identity/*` - User and organization management
- `/integrations/*` - External service connections
- Everything else goes to conversation service

**gRPC**: Single service for AI agent communication

## Entry Point

Everything starts in `cmd/main.go`:

1. Load config from `config.yaml`
2. Connect to PostgreSQL
3. Create all three services
4. Start HTTP server, gRPC server, and event subscribers
5. Run everything concurrently with errgroup