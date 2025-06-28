# Quick Start Guide

Get InfraGPT running locally in 5 minutes.

## Prerequisites

- Go 1.24+
- PostgreSQL running locally
- Git

## Setup

1. **Clone and navigate to the project**:
   ```bash
   git clone <repo-url>
   cd infragpt/services/infragpt
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Set up your database**:
   ```bash
   # Create PostgreSQL database
   createdb infragpt
   ```

4. **Create config file**:
   ```bash
   cp config-template.yaml config.yaml
   ```
   
   Edit `config.yaml` with your settings:
   ```yaml
   port: 8080
   grpc_port: 9090
   log_level: "info"
   
   database:
     host: "localhost"
     port: 5432
     db_name: "infragpt"
     user: "your_username"
     password: "your_password"
   ```

5. **Run the application**:
   ```bash
   go run ./cmd/main.go
   ```

You should see:
```
infragpt: http server starting port=8080
infragpt: grpc server starting port=9090
```

## Verify It's Working

Test the health endpoint:
```bash
curl http://localhost:8080/health
```

## Next Steps

- [Architecture Overview](./architecture.md) - Understand the codebase structure
- [Development Guide](./development.md) - Learn common development workflows
- [Configuration](./configuration.md) - Set up Slack, Clerk, and other integrations

## Common Issues

**Database connection fails**: Check PostgreSQL is running and credentials in config.yaml are correct

**Port already in use**: Change the `port` in config.yaml to an available port

**Missing config**: Make sure config.yaml exists and has required fields