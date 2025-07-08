# Backend Service

Go-based Slack bot service that provides AI-powered infrastructure management through DevOps workflows.

## Quick Start

```bash
go run ./cmd/main.go          # Run with config.yaml
go build ./cmd/main.go        # Build binary
go test ./...                 # Run all tests
sqlc generate                 # Generate Go code from SQL queries
```

## Configuration

Copy `config-template.yaml` to `config.yaml` and set your values:

```yaml
port: 8080
slack:
  client_id: "..."
  client_secret: "..."
  app_token: "..."
database:
  host: "localhost"
  port: 5432
  db_name: "backend"
  user: "backend"
  password: "..."
agent:
  endpoint: "[::]:50051"
```

## Services

- **Backend Service**: Main Slack bot with Socket Mode integration
- **Identity Service**: Clerk authentication and organization management  
- **Integration Service**: Multi-connector architecture (Slack, GitHub, AWS, GCP, PagerDuty, Datadog)

## Dependencies

- PostgreSQL with SQLC for type-safe queries
- Slack Socket Mode for real-time messaging
- gRPC for AI agent communication
- AES-256-GCM encryption for credential storage