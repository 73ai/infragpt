# GitHub App Integration - Quick Reference

A concise checklist and reference for GitHub App setup with InfraGPT.

## Environment Variables Checklist

```bash
# Required Variables
export GITHUB_APP_ID="123456"                    # ✅ Numeric App ID from GitHub
export GITHUB_WEBHOOK_SECRET="secure-secret"     # ✅ Webhook validation secret  
export GITHUB_REDIRECT_URL="https://domain.com"  # ✅ Base URL for callbacks
export GITHUB_PRIVATE_KEY="-----BEGIN RSA..."   # ✅ Complete PEM private key

# Optional Variables  
export GITHUB_WEBHOOK_PORT="8081"               # 🔧 Custom webhook port
```

## Configuration Template

### config.yaml
```yaml
integrations:
  github:
    app_id: "${GITHUB_APP_ID}"
    private_key: "${GITHUB_PRIVATE_KEY}"
    webhook_secret: "${GITHUB_WEBHOOK_SECRET}"
    redirect_url: "${GITHUB_REDIRECT_URL}"
    webhook_port: 8081  # Optional
```

### .env File (Development)
```bash
GITHUB_APP_ID=123456
GITHUB_WEBHOOK_SECRET=your-secure-webhook-secret-here
GITHUB_REDIRECT_URL=http://localhost:8080
GITHUB_WEBHOOK_PORT=8081
GITHUB_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----"
```

## GitHub App Setup Checklist

### 1. Create GitHub App
- [ ] Navigate to GitHub Settings → Developer settings → GitHub Apps
- [ ] Click "New GitHub App"
- [ ] Set unique name: `InfraGPT-[OrgName]`
- [ ] Add description and homepage URL

### 2. Configure Webhooks
- [ ] Webhook URL: `https://your-domain.com/webhooks/github`
- [ ] Generate secure webhook secret
- [ ] Enable SSL verification (production)

### 3. Set Permissions
- [ ] **Repository**: Actions (read), Contents (read), Issues (read/write), Metadata (read), Pull requests (read/write), Webhooks (read/write)
- [ ] **Organization**: Administration (read), Members (read)
- [ ] **Account**: Email addresses (read)

### 4. Subscribe to Events
- [ ] Installation (created, deleted, modified)
- [ ] Installation repositories (added, removed)
- [ ] Issues (opened, closed, edited)
- [ ] Pull request (opened, closed, merged)
- [ ] Push events
- [ ] Repository events

### 5. Generate Credentials
- [ ] Generate private key (download .pem file)
- [ ] Note App ID from settings page
- [ ] Save webhook secret

### 6. Configure Installation
- [ ] Set installation scope (organization/any account)
- [ ] Install app on test repository

## Quick Validation Commands

### Test JWT Generation
```bash
curl -X POST http://localhost:8080/api/v1/integrations/github/test-jwt
```

### Test Webhook Signature
```bash
curl -X POST http://localhost:8080/api/v1/integrations/github/test-webhook \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=test-signature" \
  -d '{"test": "payload"}'
```

### Health Check
```bash
curl http://localhost:8080/api/v1/integrations/github/health
```

## Common Error Solutions

| Error | Solution |
|-------|----------|
| `failed to parse private key` | Check PEM format, ensure complete key with headers |
| `failed to generate JWT` | Verify App ID is numeric, check private key matches app |
| `webhook signature validation failed` | Verify webhook secret matches, check payload format |
| `installation access token failed` | Check installation ID, verify app permissions |

## Development Setup (ngrok)

### 1. Install ngrok
```bash
# macOS
brew install ngrok

# Other platforms - download from ngrok.com
```

### 2. Start Tunnel
```bash
ngrok http 8080
```

### 3. Update GitHub App
- Update webhook URL to: `https://your-id.ngrok.io/webhooks/github`
- Update redirect URL to: `https://your-id.ngrok.io`

### 4. Start Application
```bash
go run ./cmd/main.go
```

## Security Checklist

- [ ] Private key stored securely (not in code)
- [ ] Webhook secret is cryptographically random
- [ ] HTTPS used for all webhook URLs
- [ ] Minimal required permissions granted
- [ ] Regular key rotation scheduled
- [ ] Environment variables not committed to VCS

## Monitoring Points

- [ ] GitHub API rate limits
- [ ] Webhook delivery success rates  
- [ ] Installation access token expiration
- [ ] Authentication error rates
- [ ] Webhook endpoint response times

## Useful GitHub URLs

- **App Settings**: `https://github.com/settings/apps`
- **Webhook Deliveries**: `https://github.com/settings/apps/[app-name]/advanced`
- **Installation Management**: `https://github.com/settings/installations`
- **API Documentation**: `https://docs.github.com/en/rest`

## Support Files

- [Complete Setup Guide](./GITHUB_APP_SETUP.md) - Detailed documentation
- [Integration System](./INTEGRATION_SYSTEM.md) - Architecture overview
- [Contributing](./CONTRIBUTING.md) - Development guidelines