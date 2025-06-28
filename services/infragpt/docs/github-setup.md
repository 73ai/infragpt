# GitHub App Setup

Quick guide for setting up GitHub App integration with InfraGPT.

## 1. Create GitHub App

**Go to**: GitHub Settings → Developer settings → GitHub Apps → "New GitHub App"

**Basic info**:
- **Name**: `InfraGPT-[YourOrgName]` (must be unique)
- **Description**: `InfraGPT DevOps automation`
- **Homepage URL**: Your organization URL
- **Webhook URL**: `https://your-domain.com/integrations/webhooks/github`

## 2. Set Permissions

**Repository permissions**:
- Contents: Read
- Issues: Read & Write  
- Pull requests: Read & Write
- Metadata: Read

**Organization permissions**:
- Members: Read

**Account permissions**:
- Email addresses: Read

## 3. Subscribe to Events

Enable these webhook events:
- Installation (created, deleted, modified)
- Issues (opened, closed, edited)
- Pull request (opened, closed, merged)
- Push events

## 4. Generate Credentials

1. **Generate private key** - download the .pem file
2. **Note the App ID** from settings page
3. **Save webhook secret** you created

## 5. Environment Variables

```bash
export GITHUB_APP_ID="123456"
export GITHUB_WEBHOOK_SECRET="your-secure-secret"
export GITHUB_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----"
```

## 6. Configuration

Add to your `config.yaml`:

```yaml
integrations:
  github:
    app_id: "${GITHUB_APP_ID}"
    private_key: "${GITHUB_PRIVATE_KEY}"
    webhook_secret: "${GITHUB_WEBHOOK_SECRET}"
    webhook_port: 8081
```

## 7. Install and Test

1. **Install the app** on a test repository
2. **Start InfraGPT**: `go run ./cmd/main.go`
3. **Test webhook**: Create an issue or PR to trigger events

## Local Development with ngrok

For local testing:

```bash
# Install ngrok
brew install ngrok

# Start tunnel
ngrok http 8080

# Update GitHub App webhook URL to:
# https://your-id.ngrok.io/integrations/webhooks/github
```

## Troubleshooting

**"failed to parse private key"**: Check PEM format is complete with headers

**"webhook signature validation failed"**: Verify webhook secret matches

**"installation access token failed"**: Check app permissions and installation

## Test Commands

```bash
# Health check
curl http://localhost:8080/health

# Test JWT generation (if endpoint exists)
curl -X POST http://localhost:8080/integrations/github/test
```

For detailed setup instructions, see the [Configuration Guide](./configuration.md).