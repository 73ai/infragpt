# Contributing to InfraGPT Slack Service

Thank you for your interest in contributing to the InfraGPT Slack service project! This guide will help you understand our development process, coding standards, and how to effectively contribute to the project.

## Table of Contents

- [Project Overview](#project-overview)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Architecture](#architecture)
- [Security Considerations](#security-considerations)

## Project Overview

InfraGPT Slack service is designed to streamline DevOps workflows, beginning with access management for cloud resources. It provides a conversational interface within Slack where users can make requests, approvers can review them, and the system can track and execute approved actions.

The current implementation focuses on the core integration with Slack and database persistence, with plans to expand to a full access request workflow system.

## Getting Started

### Prerequisites

- Go 1.24 or later
- Git
- PostgreSQL
- A Slack workspace for testing

### Setup

1. Clone the repository:
   ```
   git clone https://github.com/priyanshujain/infragpt.git
   cd infragpt/services/infragpt
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Create a `config.yaml` file with your configuration:
   ```yaml
   port: 8080
   slack:
     client_id: "your_slack_client_id"
     client_secret: "your_slack_client_secret"
     signing_secret: "your_slack_signing_secret"
     app_token: "your_slack_app_token"
   database:
     host: "localhost"
     port: 5432
     user: "postgres"
     password: "postgres"
     database: "infragpt"
   ```

4. Build and run the application:
   ```
   go build ./cmd/main.go
   ./main
   ```

5. Run tests:
   ```
   go test ./...
   ```

## Development Workflow

We follow an iterative development approach:

1. **Create a feature branch** from the main branch
2. **Implement your changes** following the code style guidelines
3. **Add tests** to verify your implementation
4. **Run existing tests** to ensure nothing breaks
5. **Submit a pull request** with a clear description of your changes

### Branching Strategy

- `master` - Main development branch
- `feature/xxx` - For new features
- `fix/xxx` - For bug fixes
- `refactor/xxx` - For refactoring work

## Code Style Guidelines

### Core Principles

#### 1. Clean Architecture

We follow clean architecture principles with clear separation between:
- Domain layer (business logic)
- Application layer (orchestration)
- Infrastructure layer (external dependencies)
- API layer (HTTP endpoints)

#### 2. Domain-Driven Design

- Use domain models that reflect the business concepts
- Define clear interfaces between layers
- Focus on the business problem, not the technical implementation

#### 3. Command Pattern

The service uses a command pattern for operations:
- Functions take command objects that encapsulate all needed parameters
- Commands follow naming conventions that clearly express their purpose
- Each command should handle a specific user intent

#### 4. Self-Documenting Code

The code should be self-documenting with:
- Descriptive type and function names that express purpose
- Proper type definitions instead of string constants
- Focused interfaces with minimal surface area
- Explicit parameter objects with clear field names

### Go-Specific Guidelines

- Follow standard Go style from [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` or `goimports` to format your code
- Always check errors and provide meaningful context
- Use CamelCase for exported symbols, camelCase for non-exported
- One package per directory, with package name matching directory name
- Use type definitions over string constants for enumerations
- Unit tests should be in the same package as the code they test, with `_test.go` suffix

## Testing

### Types of Tests

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test interactions with external systems
- **End-to-End Tests**: Test the full workflow

### Running Tests

```
# Run all tests
go test ./...

# Run tests in a specific package
go test ./path/to/package

# Run tests with coverage
go test -cover ./...
```

## Pull Request Process

1. **Create a PR** against the main branch
2. **Fill in the PR template** with details about your changes
3. **Ensure all checks pass** (tests, linting, etc.)
4. **Request reviews** from maintainers
5. **Address feedback** and make necessary changes
6. Once approved, a maintainer will merge your PR

## Architecture

### Service Interface

The service interface defines the contract between the API layer and the core business logic:

```go
type Service interface {
    Integrations(context.Context, IntegrationsQuery) ([]Integration, error)
    CompleteSlackAuthentication(context.Context, CompleteSlackAuthenticationCommand) error
}
```

This separation allows:
- Independent testing of business logic
- Flexibility to potentially support other interfaces
- Clear entry points for all system functionality

### Error Handling

Go's idiomatic error handling approach is used throughout the system:
- Functions return `error` as their last return value
- Callers are responsible for checking and handling errors
- Errors should provide appropriate context for troubleshooting

### Database Access

The system uses PostgreSQL with structured repository interfaces:
- Repository interfaces defined in the domain layer
- Implementations in the infrastructure layer
- Use of prepared statements and proper error handling

## Security Considerations

### Authentication and Authorization

- Always verify Slack message signatures
- Implement proper authorization checks for users and approvers
- Never log sensitive information

### Data Protection

- Never store sensitive credentials in plaintext
- Implement proper encryption for sensitive data
- Be mindful of data retention policies

---

We appreciate your contributions and look forward to your involvement in making the InfraGPT service even better!

If you have any questions, please reach out to the project maintainers.