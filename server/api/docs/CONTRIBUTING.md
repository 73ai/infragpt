# Contributing to InfraGPT Slack Bot

Thank you for your interest in contributing to the InfraGPT Slack Bot project! This guide will help you understand our development process, coding standards, and how to effectively contribute to the project.

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

InfraGPT Slack Bot is designed to streamline access management for cloud resources by providing a conversational interface within Slack. Users can request access to resources, approvers can review and respond to these requests, and once approved, the bot will automatically execute the necessary commands to grant access.

## Getting Started

### Prerequisites

- Go 1.24 or later
- Git
- A Slack workspace for testing

### Setup

1. Clone the repository:
   ```
   git clone https://github.com/priyanshujain/infragpt.git
   cd infragpt/server/api
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Build the application:
   ```
   go build ./cmd/main.go
   ```

4. Run tests:
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

#### 1. Focused Functionality

We take a focused approach to building the Slack bot by starting with only the core functionality needed for the access request workflow. Additional features should only be added as user needs are validated.

#### 2. Leveraging Slack's Native Features

Utilize Slack's built-in features rather than duplicating them. For example, we don't need to build features like "view my requests" since Slack already provides message history and search.

#### 3. Command-Based Design

The service API is designed using a clean command pattern:
- Functions should take command objects that encapsulate all needed parameters
- Commands should follow naming conventions that clearly express their purpose
- Each command should handle a specific user intent

#### 4. Self-Documenting Code

The code should be self-documenting with:
- Descriptive type and function names that express purpose
- Proper type definitions instead of string constants or comments
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

The service interface defines the contract between the Slack integration layer and the core business logic:

```go
type Service interface {
    AskForAccess(ctx context.Context, command AskForAccessCommand) (*AccessRequest, error)
    RespondToAccessRequest(ctx context.Context, command RespondToAccessRequestCommand) error
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

### Stateful Processing

For multi-step workflows, we use stateful storage:
- Persisted `AccessRequest` objects with status tracking
- Database to maintain state across service restarts

## Security Considerations

### Authentication and Authorization

- Always verify Slack message signatures
- Implement proper authorization checks for users and approvers
- Never log sensitive information

### Command Execution

- Always validate commands before execution
- Follow least privilege principle
- Implement timeouts and rollback mechanisms

### Data Protection

- Never store sensitive credentials in plaintext
- Implement proper encryption for sensitive data
- Be mindful of data retention policies

---

We appreciate your contributions and look forward to your involvement in making the InfraGPT Slack Bot even better!

If you have any questions, please reach out to the project maintainers.