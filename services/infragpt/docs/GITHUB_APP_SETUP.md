# GitHub App Integration Setup Guide

This guide provides comprehensive instructions for setting up GitHub App integration with InfraGPT, including environment variable configuration, GitHub App creation, and local development setup.

## Overview

InfraGPT uses GitHub Apps for secure, scalable GitHub integration. GitHub Apps provide fine-grained permissions and organization-level installations, making them ideal for enterprise DevOps workflows.

## Required Environment Variables

The following environment variables must be configured for GitHub App integration:

### Core Configuration Variables

| Variable | Description | Format | Example |
|----------|-------------|---------|---------|
| `GITHUB_APP_ID` | GitHub App ID (numeric identifier) | String (numeric) | `"123456"` |
| `GITHUB_PRIVATE_KEY` | RSA private key in PEM format | Multiline string | See [Private Key Format](#private-key-format) |
| `GITHUB_WEBHOOK_SECRET` | Webhook signature validation secret | String | `"your-webhook-secret-here"` |
| `GITHUB_REDIRECT_URL` | Base URL for OAuth callbacks and webhooks | URL | `"https://your-domain.com"` |
| `GITHUB_WEBHOOK_PORT` | Port for webhook server (optional) | Integer | `8081` |

### Private Key Format

The `GITHUB_PRIVATE_KEY` should be provided as a complete PEM-formatted RSA private key:

```
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1234567890abcdef...
...your private key content...
-----END RSA PRIVATE KEY-----
```

**Important Notes:**
- Keep all newlines intact in the private key
- The key must be in PEM format (not PKCS#8 or other formats)
- Store the key securely and never commit it to version control

## GitHub App Creation Guide

### Step 1: Create a New GitHub App

1. Navigate to your GitHub organization settings
2. Go to "Developer settings" → "GitHub Apps"
3. Click "New GitHub App"

### Step 2: Basic Information

Fill in the basic app information:

- **GitHub App name**: `InfraGPT-[YourOrgName]` (must be globally unique)
- **Description**: `InfraGPT DevOps automation and monitoring integration`
- **Homepage URL**: Your organization's website or repository URL
- **User authorization callback URL**: `https://your-domain.com/auth/github/callback`

### Step 3: Webhook Configuration

Configure webhook settings:

- **Webhook URL**: `https://your-domain.com/webhooks/github`
- **Webhook secret**: Generate a secure random string (save this for `GITHUB_WEBHOOK_SECRET`)
- **SSL verification**: Enable (recommended for production)

### Step 4: Permissions

Configure the following permissions based on your needs:

#### Repository Permissions
- **Actions**: Read (for CI/CD integration)
- **Administration**: Read (for repository settings)
- **Contents**: Read (for code access)
- **Issues**: Read & Write (for issue management)
- **Metadata**: Read (required for basic functionality)
- **Pull requests**: Read & Write (for PR management)
- **Webhooks**: Read & Write (for webhook management)

#### Organization Permissions
- **Administration**: Read (for organization settings)
- **Members**: Read (for team management)
- **Plan**: Read (for billing information)

#### Account Permissions
- **Email addresses**: Read (for user identification)

### Step 5: Events Subscription

Subscribe to the following webhook events:

- **Installation**: Installation created, deleted, modified
- **Installation repositories**: Repositories added/removed
- **Issues**: Issue opened, closed, edited
- **Pull request**: PR opened, closed, merged
- **Push**: Code pushed to repositories
- **Repository**: Repository created, deleted, modified

### Step 6: Installation Settings

- **Where can this GitHub App be installed?**: 
  - Choose "Only on this account" for organization-specific apps
  - Choose "Any account" for multi-tenant applications

### Step 7: Generate Private Key

1. After creating the app, scroll down to "Private keys"
2. Click "Generate a private key"
3. Download the `.pem` file
4. Store the contents securely (this becomes your `GITHUB_PRIVATE_KEY`)

### Step 8: Note Your App ID

- Copy the App ID from the GitHub App settings page
- This becomes your `GITHUB_APP_ID` environment variable

## Environment Variable Configuration

### Development Environment (.env file)

Create a `.env` file in your project root:

```bash
# GitHub App Configuration
GITHUB_APP_ID="123456"
GITHUB_WEBHOOK_SECRET="your-webhook-secret-here"
GITHUB_REDIRECT_URL="http://localhost:8080"
GITHUB_WEBHOOK_PORT="8081"

# Private key (multiline)
GITHUB_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1234567890abcdef...
...your private key content...
-----END RSA PRIVATE KEY-----"
```

### Production Environment

For production environments, use your platform's secure environment variable management:

#### Docker Compose
```yaml
version: '3.8'
services:
  infragpt:
    image: infragpt:latest
    environment:
      GITHUB_APP_ID: "123456"
      GITHUB_WEBHOOK_SECRET: "your-webhook-secret-here"
      GITHUB_REDIRECT_URL: "https://your-domain.com"
      GITHUB_WEBHOOK_PORT: "8081"
      GITHUB_PRIVATE_KEY: |
        -----BEGIN RSA PRIVATE KEY-----
        MIIEpAIBAAKCAQEA1234567890abcdef...
        ...your private key content...
        -----END RSA PRIVATE KEY-----
```

#### Kubernetes Secrets
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-app-secrets
type: Opaque
stringData:
  GITHUB_APP_ID: "123456"
  GITHUB_WEBHOOK_SECRET: "your-webhook-secret-here"
  GITHUB_REDIRECT_URL: "https://your-domain.com"
  GITHUB_PRIVATE_KEY: |
    -----BEGIN RSA PRIVATE KEY-----
    MIIEpAIBAAKCAQEA1234567890abcdef...
    ...your private key content...
    -----END RSA PRIVATE KEY-----
```

#### Cloud Platforms

**AWS ECS/Fargate**: Use AWS Systems Manager Parameter Store or AWS Secrets Manager
**Google Cloud Run**: Use Google Secret Manager
**Azure Container Instances**: Use Azure Key Vault

## Local Development Setup

### Prerequisites

- Go 1.21 or later
- Docker (optional, for containerized development)
- ngrok or similar tunneling tool (for webhook testing)

### Step 1: Clone and Setup

```bash
git clone <repository-url>
cd infragpt/services/infragpt
go mod download
```

### Step 2: Environment Configuration

1. Copy the example configuration:
```bash
cp config-template.yaml config.yaml
```

2. Update the GitHub section in `config.yaml`:
```yaml
integrations:
  github:
    app_id: "123456"
    private_key: "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----"
    webhook_secret: "your-webhook-secret-here"
    redirect_url: "https://your-ngrok-url.ngrok.io"
```

### Step 3: Webhook Testing with ngrok

1. Install ngrok:
```bash
# macOS
brew install ngrok

# Linux
wget https://bin.equinox.io/c/4VmDzA7iaHb/ngrok-stable-linux-amd64.zip
unzip ngrok-stable-linux-amd64.zip
```

2. Start ngrok tunnel:
```bash
ngrok http 8080
```

3. Update your GitHub App webhook URL with the ngrok URL:
```
https://your-ngrok-id.ngrok.io/webhooks/github
```

### Step 4: Run the Application

```bash
go run ./cmd/main.go
```

The application will start with:
- Main HTTP server on port 8080
- GitHub webhook server on port 8081 (if configured)
- gRPC server on port 9090

## Configuration Validation

### Testing GitHub App Configuration

You can test your GitHub App configuration using the built-in validation:

```bash
# Test JWT generation
curl -X POST http://localhost:8080/api/v1/integrations/github/test-jwt

# Test webhook signature validation
curl -X POST http://localhost:8080/api/v1/integrations/github/test-webhook \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=..." \
  -d '{"test": "payload"}'
```

### Verifying Installation

1. Install your GitHub App on a test repository
2. Check the application logs for installation events
3. Verify webhook delivery in GitHub App settings

## Troubleshooting

### Common Issues

#### 1. Private Key Format Errors

**Error**: `failed to parse private key`

**Solutions**:
- Ensure the private key is in PEM format
- Check for missing newlines or formatting issues
- Verify the key starts with `-----BEGIN RSA PRIVATE KEY-----`
- Make sure there are no extra spaces or characters

#### 2. JWT Generation Failures

**Error**: `failed to generate JWT`

**Solutions**:
- Verify the App ID is correct and numeric
- Check that the private key matches the GitHub App
- Ensure the private key is properly formatted

#### 3. Webhook Signature Validation Failures

**Error**: `webhook signature validation failed`

**Solutions**:
- Verify the webhook secret matches between GitHub and your app
- Check the webhook URL is correctly configured
- Ensure the payload is being read correctly

#### 4. Installation Access Token Errors

**Error**: `failed to get installation access token`

**Solutions**:
- Verify the installation ID is correct
- Check that the GitHub App has the required permissions
- Ensure the installation is active and not suspended

### Debugging Tips

1. **Enable Debug Logging**:
```yaml
log_level: "debug"
```

2. **Check GitHub App Webhook Deliveries**:
   - Go to GitHub App settings → Advanced → Recent Deliveries
   - Check response codes and error messages

3. **Validate JWT Tokens**:
   - Use [jwt.io](https://jwt.io) to decode and verify JWT tokens
   - Check the `iss` (issuer) claim matches your App ID

4. **Test Webhook Signatures**:
```bash
# Generate test signature
echo -n "test payload" | openssl dgst -sha256 -hmac "your-webhook-secret"
```

## Security Best Practices

### Environment Variable Security

1. **Never commit secrets to version control**
2. **Use secure secret management systems in production**
3. **Rotate webhook secrets regularly**
4. **Limit private key access to necessary personnel**

### Network Security

1. **Use HTTPS for all webhook URLs**
2. **Implement proper firewall rules**
3. **Consider IP whitelisting for webhook endpoints**

### Application Security

1. **Validate all webhook payloads**
2. **Implement rate limiting on webhook endpoints**
3. **Log security events for monitoring**
4. **Use least-privilege permissions for GitHub App**

## Monitoring and Maintenance

### Health Checks

The application provides health check endpoints:

```bash
# General health check
curl http://localhost:8080/health

# GitHub integration health check
curl http://localhost:8080/api/v1/integrations/github/health
```

### Metrics and Logging

Key metrics to monitor:

- GitHub API rate limit usage
- Webhook delivery success rates
- Installation access token refresh frequency
- Authentication failures

### Maintenance Tasks

1. **Regular Key Rotation**: Rotate private keys annually
2. **Permission Auditing**: Review GitHub App permissions quarterly
3. **Webhook Monitoring**: Monitor webhook delivery success rates
4. **Access Token Management**: Monitor token expiration and refresh

## Advanced Configuration

### Multi-Environment Setup

For organizations with multiple environments:

```yaml
# Development
integrations:
  github:
    app_id: "123456"  # Development GitHub App
    
# Production  
integrations:
  github:
    app_id: "789012"  # Production GitHub App
```

### Custom Webhook Handling

To implement custom webhook event handling:

1. Extend the `WebhookEvent` processing in `github.go`
2. Add new event types to the `EventType` enum
3. Implement custom handlers in the `ProcessEvent` method

### Enterprise Features

For GitHub Enterprise Server:

```yaml
integrations:
  github:
    api_url: "https://github.enterprise.com/api/v3"
    app_id: "123456"
    # ... other configuration
```

## Support and Resources

### GitHub Documentation

- [GitHub Apps Documentation](https://docs.github.com/en/developers/apps)
- [GitHub Webhooks Guide](https://docs.github.com/en/developers/webhooks-and-events/webhooks)
- [GitHub API Reference](https://docs.github.com/en/rest)

### InfraGPT Resources

- [Integration System Documentation](./INTEGRATION_SYSTEM.md)
- [Contributing Guidelines](./CONTRIBUTING.md)
- [Implementation Details](./IMPLEMENTATION.md)

### Getting Help

1. Check application logs for detailed error messages
2. Review GitHub App webhook delivery logs
3. Consult the troubleshooting section above
4. Create an issue in the project repository with:
   - Environment details
   - Configuration (sanitized)
   - Error messages
   - Steps to reproduce