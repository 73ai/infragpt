# InfraGPT Backend Documentation

Simple, practical guides for working with the InfraGPT backend codebase.

## 🚀 Start Here

New to the codebase? Start with these:

1. **[Quick Start Guide](./quick-start.md)** - Get running in 5 minutes
2. **[Architecture Overview](./architecture.md)** - Understand the big picture  
3. **[Development Guide](./development.md)** - Common workflows

## 📚 Detailed Guides

- **[Database Guide](./database.md)** - PostgreSQL + SQLC patterns
- **[API Guide](./api.md)** - HTTP and gRPC APIs
- **[Configuration](./configuration.md)** - Complete config reference  
- **[Testing](./testing.md)** - Testing patterns and tools
- **[GitHub Setup](./github-setup.md)** - GitHub App integration guide

## What is InfraGPT?

InfraGPT is a Go-based Slack bot that provides AI-powered infrastructure management. It consists of three main services:

- **Conversation Service** - Handles Slack messages and AI agent communication
- **Identity Service** - Manages users and organizations via Clerk
- **Integration Service** - Connects to external services (GitHub, Slack, etc.)

## Tech Stack

- **Language**: Go 1.24+
- **Database**: PostgreSQL with SQLC
- **Configuration**: YAML with mapstructure
- **API**: HTTP REST + gRPC
- **Testing**: testcontainers for integration tests

## Quick Commands

```bash
# Run locally
go run ./cmd/main.go

# Run tests
go test ./...

# Generate database code
sqlc generate

# Format code
goimports -w .
```

## Need Help?

- Check the specific guides above for detailed help
- Look at existing code examples in the codebase
- Run tests to understand expected behavior