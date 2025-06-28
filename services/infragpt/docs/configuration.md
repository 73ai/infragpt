# Configuration Guide

Complete reference for configuring InfraGPT.

## Configuration File

InfraGPT uses `config.yaml` for all configuration. Copy `config-template.yaml` to get started.

## Basic Configuration

### Minimal config.yaml

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

# Optional: Enable HTTP request logging
http_log: true
```

## Service Configuration

### Slack Configuration

For Slack bot functionality:

```yaml
slack:
  # Socket Mode (for bot events)
  app_token: "${SLACK_APP_TOKEN}"    # xapp-... token
  bot_token: "${SLACK_BOT_TOKEN}"    # xoxb-... token
  
  # OAuth (for workspace installations)
  client_id: "${SLACK_CLIENT_ID}"
  client_secret: "${SLACK_CLIENT_SECRET}"
```

**How to get tokens**:
1. Create Slack app at https://api.slack.com/apps
2. Enable Socket Mode and generate App Token
3. Install app to workspace to get Bot Token

### Identity Service (Clerk)

For user authentication:

```yaml
identity:
  clerk:
    # Webhook for user sync
    webhook_secret: "${CLERK_WEBHOOK_SECRET}"
    port: 8082  # Webhook server port
    
    # For JWT validation
    publishable_key: "${CLERK_PUBLISHABLE_KEY}"
```

**Setup**:
1. Create Clerk app at https://clerk.com
2. Configure webhook endpoint: `http://your-domain:8082/identity/webhooks/clerk`
3. Get webhook secret from Clerk dashboard

### Integration Service

For external service connections:

```yaml
integrations:
  # Slack (for OAuth installations)
  slack:
    client_id: "${SLACK_CLIENT_ID}"
    client_secret: "${SLACK_CLIENT_SECRET}"
    redirect_url: "http://localhost:3000/callback/slack"
    scopes: ["chat:write", "channels:read"]
  
  # GitHub App
  github:
    app_id: "${GITHUB_APP_ID}"
    private_key: "${GITHUB_PRIVATE_KEY}"  # Base64 encoded
    webhook_secret: "${GITHUB_WEBHOOK_SECRET}"
    webhook_port: 8081  # Dedicated webhook server
  
  # Other services
  pagerduty:
    api_key: "${PAGERDUTY_API_KEY}"
  
  datadog:
    api_key: "${DATADOG_API_KEY}"
    app_key: "${DATADOG_APP_KEY}"
```

### AI Agent Service

For gRPC communication with AI agents:

```yaml
agent:
  endpoint: "localhost:50051"  # gRPC endpoint
  timeout: "5m"               # Request timeout
  connect_timeout: "10s"      # Connection timeout
```

## Environment Variables

Use environment variables for sensitive data:

```bash
# Slack
export SLACK_APP_TOKEN="xapp-1-A0..."
export SLACK_BOT_TOKEN="xoxb-123..."
export SLACK_CLIENT_SECRET="..."

# Clerk
export CLERK_WEBHOOK_SECRET="whsec_..."
export CLERK_PUBLISHABLE_KEY="pk_..."

# GitHub
export GITHUB_APP_ID="123456"
export GITHUB_PRIVATE_KEY="LS0tLS1CRUdJTi..."  # Base64 encoded
export GITHUB_WEBHOOK_SECRET="..."

# Database
export DB_PASSWORD="your_password"

# External services
export PAGERDUTY_API_KEY="..."
export DATADOG_API_KEY="..."
export DATADOG_APP_KEY="..."
```

## Complete Example

```yaml
# config.yaml
port: 8080
grpc_port: 9090
log_level: "info"
http_log: true

database:
  host: "localhost"
  port: 5432
  db_name: "infragpt"
  user: "infragpt"
  password: "${DB_PASSWORD}"
  ssl_mode: "disable"

slack:
  app_token: "${SLACK_APP_TOKEN}"
  bot_token: "${SLACK_BOT_TOKEN}"
  client_id: "${SLACK_CLIENT_ID}"
  client_secret: "${SLACK_CLIENT_SECRET}"

agent:
  endpoint: "localhost:50051"
  timeout: "5m"
  connect_timeout: "10s"

identity:
  clerk:
    webhook_secret: "${CLERK_WEBHOOK_SECRET}"
    publishable_key: "${CLERK_PUBLISHABLE_KEY}"
    port: 8082

integrations:
  slack:
    client_id: "${SLACK_CLIENT_ID}"
    client_secret: "${SLACK_CLIENT_SECRET}"
    redirect_url: "http://localhost:3000/callback/slack"
    scopes: ["chat:write", "channels:read", "users:read"]
  
  github:
    app_id: "${GITHUB_APP_ID}"
    private_key: "${GITHUB_PRIVATE_KEY}"
    webhook_secret: "${GITHUB_WEBHOOK_SECRET}"
    webhook_port: 8081
  
  pagerduty:
    api_key: "${PAGERDUTY_API_KEY}"
  
  datadog:
    api_key: "${DATADOG_API_KEY}"
    app_key: "${DATADOG_APP_KEY}"
```

## Configuration Patterns

### Service Configuration Struct

Each service defines its own config:

```go
// Example: internal/servicename/config.go
type Config struct {
    APIKey    string        `mapstructure:"api_key"`
    Timeout   time.Duration `mapstructure:"timeout"`
    EnableSSL bool          `mapstructure:"enable_ssl"`
    Database  *sql.DB       // Injected, not from YAML
}

func (c Config) New() (*Service, error) {
    if c.APIKey == "" {
        return nil, fmt.Errorf("api_key is required")
    }
    
    return &Service{
        client: &http.Client{Timeout: c.Timeout},
        db:     c.Database,
    }, nil
}
```

### Conditional Registration

Services only start if properly configured:

```go
// In cmd/main.go
if c.Slack.BotToken != "" {
    slackService, err := c.Slack.New()
    // Register and start service
}

if c.GitHub.AppID != "" {
    githubConnector := c.GitHub.NewConnector()
    connectors[ConnectorTypeGitHub] = githubConnector
}
```

## Development vs Production

### Development

```yaml
log_level: "debug"
http_log: true

database:
  host: "localhost"
  ssl_mode: "disable"
```

### Production

```yaml
log_level: "info"
http_log: false

database:
  host: "prod-db-host"
  ssl_mode: "require"
  max_connections: 25
  
# Use proper domains
integrations:
  slack:
    redirect_url: "https://app.infragpt.com/callback/slack"
```

## Configuration Validation

Services validate configuration at startup:

```go
func (c Config) Validate() error {
    if c.APIKey == "" {
        return fmt.Errorf("api_key is required")
    }
    if c.Timeout <= 0 {
        return fmt.Errorf("timeout must be positive")
    }
    return nil
}
```

## Security Best Practices

1. **Never commit secrets** - use environment variables
2. **Use strong webhook secrets** - generate random values
3. **Enable SSL in production** - set `ssl_mode: "require"`
4. **Rotate keys regularly** - especially webhook secrets
5. **Use minimal scopes** - only request needed permissions
6. **Validate URLs** - ensure redirect URLs are HTTPS in production

## Troubleshooting

### Database Connection Issues

```yaml
database:
  host: "localhost"
  port: 5432
  db_name: "infragpt"
  user: "infragpt"
  password: "password"
  ssl_mode: "disable"    # Try this for local development
  connect_timeout: "10s" # Add if connections are slow
```

### Slack Connection Issues

```bash
# Check tokens are valid
curl -H "Authorization: Bearer ${SLACK_BOT_TOKEN}" \
  https://slack.com/api/auth.test

# Verify Socket Mode is enabled in Slack app settings
# Ensure app is installed to workspace
```

### GitHub App Issues

```bash
# Verify app installation
curl -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  https://api.github.com/app/installations

# Check webhook delivery in GitHub app settings
# Ensure webhook URL is accessible
```

## Common Configurations

### Local Development

```yaml
port: 8080
grpc_port: 9090
log_level: "debug"
http_log: true

database:
  host: "localhost"
  port: 5432
  db_name: "infragpt_dev"
  user: "postgres"
  password: "postgres"
  ssl_mode: "disable"
```

### Docker Compose

```yaml
database:
  host: "postgres"  # Service name in docker-compose
  port: 5432
  db_name: "infragpt"
  user: "infragpt"
  password: "${DB_PASSWORD}"
```

### Production

```yaml
log_level: "info"

database:
  host: "${DB_HOST}"
  port: 5432
  db_name: "${DB_NAME}"
  user: "${DB_USER}"
  password: "${DB_PASSWORD}"
  ssl_mode: "require"
  max_connections: 25
```