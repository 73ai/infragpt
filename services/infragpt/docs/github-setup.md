# GitHub App Setup

## 1. Create GitHub App

Go to: GitHub Settings → Developer settings → GitHub Apps → "New GitHub App"

- **Name**: `InfraGPT-[YourOrgName]` (must be unique)
- **Description**: `InfraGPT DevOps automation`
- **Homepage URL**: Your organization URL
- **Webhook URL**: `https://your-domain.com/integrations/webhooks/github`

## 2. Set Permissions

**Repository permissions**: Contents (Read), Issues (Read & Write), Pull requests (Read & Write), Metadata (Read)
**Organization permissions**: Members (Read)
**Account permissions**: Email addresses (Read)

## 3. Subscribe to Events

Enable: Installation, Issues, Pull request, Push events

## 4. Generate Credentials

1. Generate private key (download .pem file)
2. Note the App ID from settings page
3. Save webhook secret

## 5. Environment Variables

```bash
export GITHUB_APP_ID="123456"
export GITHUB_WEBHOOK_SECRET="your-secure-secret"
export GITHUB_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----"
```

## 6. Configuration

Add to `config.yaml`:

```yaml
integrations:
  github:
    app_id: "${GITHUB_APP_ID}"
    private_key: "${GITHUB_PRIVATE_KEY}"
    webhook_secret: "${GITHUB_WEBHOOK_SECRET}"
    webhook_port: 8081
```

## 7. Install and Test

1. Install the app on a test repository
2. Start InfraGPT: `go run ./cmd/main.go`
3. Test by creating an issue or PR to trigger events