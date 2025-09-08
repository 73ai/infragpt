.PHONY: dev build build-console check format clean help tail-log

# Default target
help:
	@echo "Available targets:"
	@echo "  dev       - Start development servers (console + backend + agent)"
	@echo "  build     - Build production binary with embedded console"
	@echo "  check     - Run linting and type checking"
	@echo "  format    - Format all code"
	@echo "  clean     - Clean build artifacts"
	@echo "  tail-log  - Show the last 100 lines of the log"

# Development mode - start both console, agent and backend
dev:
	@ENV=development ./development/shoreman.sh

# Production build
build: build-console
	@echo "Building production binary..."
	@mkdir -p bin
	@go build -o bin/infragpt ./services/backend/cmd/main.go
	@echo "Production binary created at bin/infragpt"

# Build console for production
build-console:
	@echo "Building console..."
	@cd services/console && npm install && npm run build

# Linting and type checking
check:
	@echo "Running Go checks..."
	@cd services/backend
	@go vet ./...
	@go mod tidy
	@cd ../..
	@echo "Running console checks..."
	@cd services/console && npm run lint
	@cd ../..
	@echo "Running agent checks..."
	@cd services/agent && ruff check .
	@cd ../.. && echo "All checks passed."


# Format code
format:
	@echo "Formatting Go code..."
	@cd services/backend && go fmt ./...
	@cd ../..
	@echo "Formatting console code..."
	@cd services/console && npm run format
	@cd ../..
	@echo "Formatting agent code..."
	@cd services/agent && ruff format .
	@cd ../.. && echo "All code formatted."

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf services/console/dist/
	@rm -rf services/console/node_modules/
	@rm -rf services/agent/__pycache__/
	@echo "Clean complete."

# Display the last 100 lines of development log with ANSI codes stripped
tail-log:
	@tail -100 ./dev.log | perl -pe 's/\e\[[0-9;]*m(?:\e\[K)?//g'