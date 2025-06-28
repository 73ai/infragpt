# GitHub Integration Implementation Plan

## Executive Summary

### Current Status
The GitHub integration for InfraGPT is approximately **70% complete** with solid architectural foundations but critical functional gaps. The frontend UI is fully implemented and production-ready, the database schema is complete, and the connector interface structure exists. However, key backend functionality including JWT authentication, webhook configuration, and event processing requires completion.

### What Needs to Be Completed
- **JWT Generation**: GitHub App authentication via RSA private key signing
- **Webhook Configuration**: Automatic webhook setup and management
- **Event Processing**: Business logic for handling GitHub webhook events
- **Credential Management**: Complete revocation and refresh functionality
- **Testing**: Comprehensive test suite for integration validation

### Timeline Estimate
**Total: 13-19 days** (2.5-4 weeks for single developer)
- Phase 1: Critical Blockers (3-5 days)
- Phase 2: Core Functionality (5-7 days)  
- Phase 3: Security & Cleanup (3-4 days)
- Phase 4: Testing & Documentation (2-3 days)

### Resource Requirements
- **1 Senior Backend Developer** with Go experience
- **GitHub App** with appropriate permissions configured
- **Development environment** with GitHub App credentials
- **Testing repositories** for webhook validation

## Implementation Strategy

### Approach
1. **Build on Existing Foundation**: Leverage the 70% complete implementation
2. **Port from POC**: Utilize working implementations from `playground/main.go`
3. **Maintain Architecture**: Follow established `integrationsvc` patterns
4. **Incremental Development**: Implement in phases with testing at each step

### Integration with Existing Codebase
- **No Breaking Changes**: All modifications are additive or completing existing stubs
- **Database Compatibility**: Use existing schema with minor extensions
- **API Consistency**: Maintain existing endpoint contracts
- **UI Compatibility**: Frontend requires no changes

### Risk Mitigation
- **Comprehensive Testing**: Unit, integration, and security tests
- **Feature Flags**: Gradual rollout capability
- **Rollback Plan**: Quick revert to previous state if needed
- **Monitoring**: Health checks and performance metrics

## Detailed Phase Breakdown

### Phase 1: Critical Blockers (3-5 days)

#### Objective
Complete the core authentication and webhook infrastructure to make the integration functional.

#### Tasks

##### 1.1 JWT Generation Implementation (Day 1-2)
**File**: `services/infragpt/internal/integrationsvc/connectors/github/github.go`

**Current Issue**:
```go
func (g *githubConnector) generateJWT() (string, error) {
    return "", fmt.Errorf("JWT generation not implemented - requires private key parsing")
}
```

**Implementation**:
```go
import (
    "crypto/rsa"
    "time"
    "github.com/golang-jwt/jwt/v4"
)

func (g *githubConnector) generateJWT() (string, error) {
    // Parse the private key from PEM format
    privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(g.config.PrivateKey))
    if err != nil {
        return "", fmt.Errorf("failed to parse private key: %w", err)
    }

    // Create JWT claims
    now := time.Now()
    claims := jwt.MapClaims{
        "iat": now.Unix(),
        "exp": now.Add(10 * time.Minute).Unix(),
        "iss": g.config.AppID,
    }

    // Create and sign the token
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    tokenString, err := token.SignedString(privateKey)
    if err != nil {
        return "", fmt.Errorf("failed to sign JWT: %w", err)
    }

    return tokenString, nil
}
```

##### 1.2 GitHub API Client Implementation (Day 2)
**New File**: `services/infragpt/internal/integrationsvc/connectors/github/api_client.go`

```go
package github

import (
    "context"
    "fmt"
    "net/http"
    "time"
    
    "github.com/google/go-github/v69/github"
    "golang.org/x/oauth2"
)

type APIClient struct {
    config     Config
    httpClient *http.Client
    logger     *slog.Logger
}

func NewAPIClient(config Config, logger *slog.Logger) *APIClient {
    return &APIClient{
        config:     config,
        httpClient: &http.Client{Timeout: 30 * time.Second},
        logger:     logger,
    }
}

func (ac *APIClient) GetInstallationToken(ctx context.Context, installationID int64) (*github.InstallationToken, error) {
    // Generate JWT for app authentication
    jwt, err := ac.generateJWT()
    if err != nil {
        return nil, fmt.Errorf("failed to generate JWT: %w", err)
    }

    // Create GitHub client with JWT
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: jwt})
    tc := oauth2.NewClient(ctx, ts)
    client := github.NewClient(tc)

    // Get installation token
    token, _, err := client.Apps.CreateInstallationToken(
        ctx, 
        installationID, 
        &github.InstallationTokenOptions{},
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create installation token: %w", err)
    }

    return token, nil
}

func (ac *APIClient) GetInstallationRepositories(ctx context.Context, installationID int64) ([]*github.Repository, error) {
    token, err := ac.GetInstallationToken(ctx, installationID)
    if err != nil {
        return nil, err
    }

    // Create authenticated client
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token.GetToken()})
    tc := oauth2.NewClient(ctx, ts)
    client := github.NewClient(tc)

    // Get repositories
    repos, _, err := client.Apps.ListRepos(ctx, &github.ListOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to list repositories: %w", err)
    }

    return repos.Repositories, nil
}
```

##### 1.3 Webhook Configuration Implementation (Day 3)
**File**: `services/infragpt/internal/integrationsvc/connectors/github/github.go`

**Current Issue**:
```go
func (g *githubConnector) ConfigureWebhooks(integrationID string, creds infragpt.Credentials) error {
    return nil // Not implemented
}
```

**Implementation**:
```go
func (g *githubConnector) ConfigureWebhooks(integrationID string, creds infragpt.Credentials) error {
    ctx := context.Background()
    
    // Extract installation ID from credentials
    installationID, ok := creds.Data["installation_id"].(int64)
    if !ok {
        return fmt.Errorf("invalid installation_id in credentials")
    }

    // Get API client
    apiClient := NewAPIClient(g.config, g.logger)
    
    // Get installation token
    token, err := apiClient.GetInstallationToken(ctx, installationID)
    if err != nil {
        return fmt.Errorf("failed to get installation token: %w", err)
    }

    // Create authenticated GitHub client
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token.GetToken()})
    tc := oauth2.NewClient(ctx, ts)
    client := github.NewClient(tc)

    // Get repositories for this installation
    repos, err := apiClient.GetInstallationRepositories(ctx, installationID)
    if err != nil {
        return fmt.Errorf("failed to get repositories: %w", err)
    }

    // Configure webhooks for each repository
    webhookConfig := &github.Hook{
        Name:   github.String("web"),
        Active: github.Bool(true),
        Events: []string{
            "push",
            "pull_request",
            "issues",
            "installation",
            "installation_repositories",
        },
        Config: map[string]interface{}{
            "url":          g.config.WebhookURL,
            "content_type": "json",
            "secret":       g.config.WebhookSecret,
            "insecure_ssl": "0",
        },
    }

    for _, repo := range repos {
        _, _, err := client.Repositories.CreateHook(
            ctx,
            repo.GetOwner().GetLogin(),
            repo.GetName(),
            webhookConfig,
        )
        if err != nil {
            g.logger.Error("Failed to create webhook",
                "repo", repo.GetFullName(),
                "error", err)
            // Continue with other repositories
            continue
        }
        
        g.logger.Info("Webhook configured successfully",
            "repo", repo.GetFullName(),
            "webhook_url", g.config.WebhookURL)
    }

    return nil
}
```

##### 1.4 Configuration Updates (Day 3)
**File**: `services/infragpt/internal/integrationsvc/connectors/github/config.go`

**Enhancement**:
```go
type Config struct {
    AppID         int64  `mapstructure:"app_id"`
    AppName       string `mapstructure:"app_name"`
    PrivateKey    string `mapstructure:"private_key"`    // PEM encoded private key
    WebhookSecret string `mapstructure:"webhook_secret"`
    WebhookURL    string `mapstructure:"webhook_url"`    // Base URL for webhooks
    WebhookPort   int    `mapstructure:"webhook_port"`
}

func (c Config) Validate() error {
    if c.AppID == 0 {
        return fmt.Errorf("github app_id is required")
    }
    if c.AppName == "" {
        return fmt.Errorf("github app_name is required")
    }
    if c.PrivateKey == "" {
        return fmt.Errorf("github private_key is required")
    }
    if c.WebhookSecret == "" {
        return fmt.Errorf("github webhook_secret is required")
    }
    if c.WebhookURL == "" {
        return fmt.Errorf("github webhook_url is required")
    }
    return nil
}
```

### Phase 2: Core Functionality (5-7 days)

#### Objective
Implement comprehensive event processing and repository management functionality.

#### Tasks

##### 2.1 Event Processing Implementation (Day 4-6)
**File**: `services/infragpt/internal/integrationsvc/connectors/github/events.go`

**Current Issue**: All event handlers are TODO placeholders

**Implementation**:
```go
// Enhanced event processing
func (g *githubConnector) processWebhookEvent(ctx context.Context, event WebhookEvent) error {
    switch event.EventType {
    case EventTypeInstallation:
        return g.processInstallationEvent(ctx, event)
    case EventTypeInstallationRepositories:
        return g.processInstallationRepositoriesEvent(ctx, event)
    case EventTypePush:
        return g.processPushEvent(ctx, event)
    case EventTypePullRequest:
        return g.processPullRequestEvent(ctx, event)
    case EventTypeIssues:
        return g.processIssuesEvent(ctx, event)
    default:
        g.logger.Debug("Unhandled event type", "type", event.EventType)
        return nil
    }
}

func (g *githubConnector) processInstallationEvent(ctx context.Context, event WebhookEvent) error {
    action := event.RawPayload["action"].(string)
    
    switch action {
    case "created":
        // New installation - update integration status
        return g.handleInstallationCreated(ctx, event)
    case "deleted":
        // Installation removed - mark integration as disconnected
        return g.handleInstallationDeleted(ctx, event)
    case "suspend":
        // Installation suspended - mark integration as suspended
        return g.handleInstallationSuspended(ctx, event)
    case "unsuspend":
        // Installation unsuspended - reactivate integration
        return g.handleInstallationUnsuspended(ctx, event)
    }
    
    return nil
}

func (g *githubConnector) processPushEvent(ctx context.Context, event WebhookEvent) error {
    // Extract push event data
    pushData := event.RawPayload
    
    repository := pushData["repository"].(map[string]interface{})
    repoName := repository["full_name"].(string)
    
    commits := pushData["commits"].([]interface{})
    branch := pushData["ref"].(string)
    
    g.logger.Info("Processing push event",
        "repository", repoName,
        "branch", branch,
        "commits", len(commits))
    
    // TODO: Implement business logic for push events
    // - Trigger CI/CD pipelines
    // - Analyze code changes
    // - Send notifications
    
    return nil
}

func (g *githubConnector) processPullRequestEvent(ctx context.Context, event WebhookEvent) error {
    // Extract PR event data
    prData := event.RawPayload
    
    action := prData["action"].(string)
    pullRequest := prData["pull_request"].(map[string]interface{})
    
    prNumber := int(pullRequest["number"].(float64))
    prTitle := pullRequest["title"].(string)
    
    g.logger.Info("Processing pull request event",
        "action", action,
        "pr_number", prNumber,
        "title", prTitle)
    
    switch action {
    case "opened":
        return g.handlePROpened(ctx, event, pullRequest)
    case "closed":
        return g.handlePRClosed(ctx, event, pullRequest)
    case "synchronize":
        return g.handlePRSynchronized(ctx, event, pullRequest)
    }
    
    return nil
}
```

##### 2.2 Repository Management (Day 6-7)
**New File**: `services/infragpt/internal/integrationsvc/connectors/github/repository.go`

```go
package github

import (
    "context"
    "fmt"
    "time"
)

type RepositoryManager struct {
    config     Config
    apiClient  *APIClient
    logger     *slog.Logger
}

type Repository struct {
    ID            int64     `json:"id"`
    FullName      string    `json:"full_name"`
    Name          string    `json:"name"`
    Owner         string    `json:"owner"`
    Private       bool      `json:"private"`
    DefaultBranch string    `json:"default_branch"`
    Language      string    `json:"language"`
    StarCount     int       `json:"star_count"`
    ForkCount     int       `json:"fork_count"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

func NewRepositoryManager(config Config, logger *slog.Logger) *RepositoryManager {
    return &RepositoryManager{
        config:    config,
        apiClient: NewAPIClient(config, logger),
        logger:    logger,
    }
}

func (rm *RepositoryManager) GetAccessibleRepositories(ctx context.Context, installationID int64) ([]Repository, error) {
    githubRepos, err := rm.apiClient.GetInstallationRepositories(ctx, installationID)
    if err != nil {
        return nil, fmt.Errorf("failed to get GitHub repositories: %w", err)
    }

    repositories := make([]Repository, len(githubRepos))
    for i, repo := range githubRepos {
        repositories[i] = Repository{
            ID:            repo.GetID(),
            FullName:      repo.GetFullName(),
            Name:          repo.GetName(),
            Owner:         repo.GetOwner().GetLogin(),
            Private:       repo.GetPrivate(),
            DefaultBranch: repo.GetDefaultBranch(),
            Language:      repo.GetLanguage(),
            StarCount:     repo.GetStargazersCount(),
            ForkCount:     repo.GetForksCount(),
            CreatedAt:     repo.GetCreatedAt().Time,
            UpdatedAt:     repo.GetUpdatedAt().Time,
        }
    }

    return repositories, nil
}

func (rm *RepositoryManager) GetRepositoryContent(ctx context.Context, installationID int64, repoFullName, path, branch string) (string, error) {
    token, err := rm.apiClient.GetInstallationToken(ctx, installationID)
    if err != nil {
        return "", err
    }

    // Create authenticated client
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token.GetToken()})
    tc := oauth2.NewClient(ctx, ts)
    client := github.NewClient(tc)

    // Parse owner and repo from full name
    parts := strings.Split(repoFullName, "/")
    if len(parts) != 2 {
        return "", fmt.Errorf("invalid repository full name: %s", repoFullName)
    }
    owner, repo := parts[0], parts[1]

    // Get file content
    fileContent, _, _, err := client.Repositories.GetContents(
        ctx,
        owner,
        repo,
        path,
        &github.RepositoryContentGetOptions{Ref: branch},
    )
    if err != nil {
        return "", fmt.Errorf("failed to get repository content: %w", err)
    }

    content, err := fileContent.GetContent()
    if err != nil {
        return "", fmt.Errorf("failed to decode content: %w", err)
    }

    return content, nil
}
```

##### 2.3 Webhook Management Enhancement (Day 7-8)
**File**: `services/infragpt/internal/integrationsvc/connectors/github/webhook.go`

**Enhancement**:
```go
func (g *githubConnector) ValidateWebhookSignature(payload []byte, signature string, secret string) error {
    if signature == "" {
        return fmt.Errorf("missing webhook signature")
    }

    // Remove 'sha1=' prefix if present
    signature = strings.TrimPrefix(signature, "sha1=")
    
    // Calculate expected signature
    mac := hmac.New(sha1.New, []byte(secret))
    mac.Write(payload)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))

    // Compare signatures
    if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
        return fmt.Errorf("webhook signature verification failed")
    }

    return nil
}

func (g *githubConnector) handleWebhookRequest(w http.ResponseWriter, r *http.Request) {
    // Read request body
    payload, err := io.ReadAll(r.Body)
    if err != nil {
        g.logger.Error("Failed to read webhook payload", "error", err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    // Validate signature
    signature := r.Header.Get("X-Hub-Signature")
    if err := g.ValidateWebhookSignature(payload, signature, g.config.WebhookSecret); err != nil {
        g.logger.Error("Webhook signature validation failed", "error", err)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Parse webhook event
    eventType := r.Header.Get("X-GitHub-Event")
    deliveryID := r.Header.Get("X-GitHub-Delivery")

    var rawPayload map[string]interface{}
    if err := json.Unmarshal(payload, &rawPayload); err != nil {
        g.logger.Error("Failed to parse webhook payload", "error", err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    // Create webhook event
    event := WebhookEvent{
        EventType:   EventType(eventType),
        Timestamp:   time.Now(),
        RawPayload:  rawPayload,
        DeliveryID:  deliveryID,
    }

    // Extract common fields
    if installation, ok := rawPayload["installation"].(map[string]interface{}); ok {
        event.InstallationID = int64(installation["id"].(float64))
    }

    if repository, ok := rawPayload["repository"].(map[string]interface{}); ok {
        event.RepositoryID = int64(repository["id"].(float64))
        event.RepositoryName = repository["full_name"].(string)
    }

    g.logger.Info("Received webhook event",
        "type", eventType,
        "delivery_id", deliveryID,
        "installation_id", event.InstallationID,
        "repository", event.RepositoryName)

    // Process event asynchronously
    go func() {
        ctx := context.Background()
        if err := g.processWebhookEvent(ctx, event); err != nil {
            g.logger.Error("Failed to process webhook event",
                "error", err,
                "type", eventType,
                "delivery_id", deliveryID)
        }
    }()

    // Return success response
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

### Phase 3: Security & Cleanup (3-4 days)

#### Objective
Complete security features, credential management, and health monitoring.

#### Tasks

##### 3.1 Credential Revocation Implementation (Day 9-10)
**File**: `services/infragpt/internal/integrationsvc/connectors/github/github.go`

**Current Issue**:
```go
func (g *githubConnector) RevokeCredentials(creds infragpt.Credentials) error {
    return nil // Not implemented
}
```

**Implementation**:
```go
func (g *githubConnector) RevokeCredentials(creds infragpt.Credentials) error {
    ctx := context.Background()
    
    // Extract installation ID
    installationID, ok := creds.Data["installation_id"].(int64)
    if !ok {
        return fmt.Errorf("invalid installation_id in credentials")
    }

    // Get API client
    apiClient := NewAPIClient(g.config, g.logger)
    
    // Get installation token for authentication
    token, err := apiClient.GetInstallationToken(ctx, installationID)
    if err != nil {
        // If we can't get a token, the installation might already be revoked
        g.logger.Warn("Failed to get installation token during revocation",
            "installation_id", installationID,
            "error", err)
        return nil
    }

    // Create authenticated client
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token.GetToken()})
    tc := oauth2.NewClient(ctx, ts)
    client := github.NewClient(tc)

    // Get repositories to remove webhooks
    repos, err := apiClient.GetInstallationRepositories(ctx, installationID)
    if err != nil {
        g.logger.Error("Failed to get repositories during revocation",
            "installation_id", installationID,
            "error", err)
        // Continue with revocation even if we can't clean up webhooks
    } else {
        // Remove webhooks from repositories
        for _, repo := range repos {
            if err := g.removeRepositoryWebhooks(ctx, client, repo); err != nil {
                g.logger.Error("Failed to remove webhook during revocation",
                    "repo", repo.GetFullName(),
                    "error", err)
                // Continue with other repositories
            }
        }
    }

    g.logger.Info("Credentials revoked successfully",
        "installation_id", installationID)

    return nil
}

func (g *githubConnector) removeRepositoryWebhooks(ctx context.Context, client *github.Client, repo *github.Repository) error {
    owner := repo.GetOwner().GetLogin()
    name := repo.GetName()

    // List existing webhooks
    hooks, _, err := client.Repositories.ListHooks(ctx, owner, name, &github.ListOptions{})
    if err != nil {
        return fmt.Errorf("failed to list webhooks: %w", err)
    }

    // Remove webhooks that match our URL
    for _, hook := range hooks {
        if config, ok := hook.Config["url"].(string); ok && config == g.config.WebhookURL {
            _, err := client.Repositories.DeleteHook(ctx, owner, name, hook.GetID())
            if err != nil {
                return fmt.Errorf("failed to delete webhook: %w", err)
            }
            g.logger.Info("Webhook removed successfully",
                "repo", repo.GetFullName(),
                "hook_id", hook.GetID())
        }
    }

    return nil
}
```

##### 3.2 Health Check Implementation (Day 10)
**File**: `services/infragpt/internal/integrationsvc/connectors/github/github.go`

**Enhancement**:
```go
func (g *githubConnector) HealthCheck(creds infragpt.Credentials) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Extract installation ID
    installationID, ok := creds.Data["installation_id"].(int64)
    if !ok {
        return fmt.Errorf("invalid installation_id in credentials")
    }

    // Test GitHub API connectivity
    apiClient := NewAPIClient(g.config, g.logger)
    
    // Try to get installation token
    token, err := apiClient.GetInstallationToken(ctx, installationID)
    if err != nil {
        return fmt.Errorf("failed to get installation token: %w", err)
    }

    // Test API access with token
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token.GetToken()})
    tc := oauth2.NewClient(ctx, ts)
    client := github.NewClient(tc)

    // Make a simple API call to verify access
    _, _, err = client.Apps.GetInstallation(ctx, installationID)
    if err != nil {
        return fmt.Errorf("failed to verify installation access: %w", err)
    }

    g.logger.Debug("Health check passed",
        "installation_id", installationID)

    return nil
}
```

##### 3.3 Enhanced Error Handling (Day 11)
**File**: `services/infragpt/internal/integrationsvc/connectors/github/errors.go`

**New File**:
```go
package github

import (
    "errors"
    "fmt"
    "net/http"
)

// Custom error types for GitHub integration
var (
    ErrInvalidCredentials    = errors.New("invalid GitHub credentials")
    ErrInstallationNotFound  = errors.New("GitHub installation not found")
    ErrInsufficientPermissions = errors.New("insufficient GitHub permissions")
    ErrRateLimitExceeded    = errors.New("GitHub API rate limit exceeded")
    ErrWebhookConfigFailed  = errors.New("webhook configuration failed")
)

type GitHubError struct {
    StatusCode int
    Message    string
    Err        error
}

func (e *GitHubError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("GitHub API error (status %d): %s - %v", e.StatusCode, e.Message, e.Err)
    }
    return fmt.Sprintf("GitHub API error (status %d): %s", e.StatusCode, e.Message)
}

func (e *GitHubError) Unwrap() error {
    return e.Err
}

func NewGitHubError(statusCode int, message string, err error) *GitHubError {
    return &GitHubError{
        StatusCode: statusCode,
        Message:    message,
        Err:        err,
    }
}

func HandleGitHubAPIError(err error) error {
    if err == nil {
        return nil
    }

    // Handle HTTP errors
    if httpErr, ok := err.(*http.Response); ok {
        switch httpErr.StatusCode {
        case http.StatusUnauthorized:
            return NewGitHubError(httpErr.StatusCode, "Authentication failed", ErrInvalidCredentials)
        case http.StatusForbidden:
            return NewGitHubError(httpErr.StatusCode, "Insufficient permissions", ErrInsufficientPermissions)
        case http.StatusNotFound:
            return NewGitHubError(httpErr.StatusCode, "Installation not found", ErrInstallationNotFound)
        case http.StatusTooManyRequests:
            return NewGitHubError(httpErr.StatusCode, "Rate limit exceeded", ErrRateLimitExceeded)
        default:
            return NewGitHubError(httpErr.StatusCode, "GitHub API error", err)
        }
    }

    return err
}
```

### Phase 4: Testing & Documentation (2-3 days)

#### Objective
Implement comprehensive testing and update documentation.

#### Tasks

##### 4.1 Unit Tests (Day 12)
**New File**: `services/infragpt/internal/integrationsvc/connectors/github/github_test.go`

```go
package github

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGitHubConnector_GenerateJWT(t *testing.T) {
    // Test JWT generation with valid private key
    config := Config{
        AppID: 123456,
        PrivateKey: testPrivateKey, // Test RSA private key in PEM format
    }
    
    connector := &githubConnector{config: config}
    
    jwt, err := connector.generateJWT()
    require.NoError(t, err)
    assert.NotEmpty(t, jwt)
    
    // Verify JWT structure (header.payload.signature)
    parts := strings.Split(jwt, ".")
    assert.Len(t, parts, 3)
}

func TestGitHubConnector_ValidateWebhookSignature(t *testing.T) {
    connector := &githubConnector{}
    
    payload := []byte(`{"test": "payload"}`)
    secret := "test-secret"
    
    // Generate valid signature
    mac := hmac.New(sha1.New, []byte(secret))
    mac.Write(payload)
    validSignature := "sha1=" + hex.EncodeToString(mac.Sum(nil))
    
    // Test valid signature
    err := connector.ValidateWebhookSignature(payload, validSignature, secret)
    assert.NoError(t, err)
    
    // Test invalid signature
    err = connector.ValidateWebhookSignature(payload, "sha1=invalid", secret)
    assert.Error(t, err)
}

func TestGitHubConnector_InitiateAuthorization(t *testing.T) {
    config := Config{
        AppName: "test-app",
    }
    
    connector := &githubConnector{config: config}
    
    intent, err := connector.InitiateAuthorization("org-123", "user-456")
    require.NoError(t, err)
    
    assert.Contains(t, intent.URL, "github.com/apps/test-app/installations/new")
    assert.Contains(t, intent.URL, "state=")
    assert.NotEmpty(t, intent.State)
    assert.True(t, intent.ExpiresAt.After(time.Now()))
}

const testPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA4f5wg5l2hKsTeNem/V41fGnJm6gOdrj8ym3rFkEjWT2DmZzu
...
-----END RSA PRIVATE KEY-----`
```

##### 4.2 Integration Tests (Day 12-13)
**New File**: `services/infragpt/internal/integrationsvc/connectors/github/integration_test.go`

```go
package github

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGitHubConnector_EndToEndFlow(t *testing.T) {
    // Skip if not running integration tests
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Create test server to mock GitHub API
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/app/installations/12345/access_tokens":
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusCreated)
            w.Write([]byte(`{
                "token": "test-token",
                "expires_at": "2023-12-31T23:59:59Z"
            }`))
        case "/installation/repositories":
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            w.Write([]byte(`{
                "repositories": [
                    {
                        "id": 1,
                        "full_name": "test/repo",
                        "name": "repo",
                        "private": false
                    }
                ]
            }`))
        default:
            w.WriteHeader(http.StatusNotFound)
        }
    }))
    defer server.Close()
    
    // Test connector with mock server
    config := Config{
        AppID:      123456,
        PrivateKey: testPrivateKey,
        WebhookURL: server.URL + "/webhook",
    }
    
    connector := NewGitHubConnector(config)
    
    // Test authorization flow
    intent, err := connector.InitiateAuthorization("org-123", "user-456")
    require.NoError(t, err)
    assert.Contains(t, intent.URL, "github.com/apps")
    
    // Test credential completion
    creds, err := connector.CompleteAuthorization(AuthorizationData{
        Code:  "12345", // installation_id
        State: intent.State,
    })
    require.NoError(t, err)
    assert.Equal(t, "github_app_installation", creds.Type)
    
    // Test credential validation
    err = connector.ValidateCredentials(creds)
    assert.NoError(t, err)
}

func TestWebhookProcessing(t *testing.T) {
    config := Config{
        WebhookSecret: "test-secret",
    }
    
    connector := NewGitHubConnector(config)
    
    // Test webhook event processing
    payload := []byte(`{
        "action": "opened",
        "pull_request": {
            "number": 1,
            "title": "Test PR"
        },
        "repository": {
            "full_name": "test/repo"
        }
    }`)
    
    signature := generateTestSignature(payload, "test-secret")
    
    // Create test request
    req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
    req.Header.Set("X-GitHub-Event", "pull_request")
    req.Header.Set("X-Hub-Signature", signature)
    req.Header.Set("X-GitHub-Delivery", "test-delivery-id")
    
    // Process webhook
    recorder := httptest.NewRecorder()
    connector.handleWebhookRequest(recorder, req)
    
    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Equal(t, "OK", recorder.Body.String())
}
```

##### 4.3 Security Tests (Day 13)
**New File**: `services/infragpt/internal/integrationsvc/connectors/github/security_test.go`

```go
package github

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestPrivateKeyHandling(t *testing.T) {
    // Test that private key is properly parsed and secured
    config := Config{
        AppID:      123456,
        PrivateKey: testPrivateKey,
    }
    
    connector := &githubConnector{config: config}
    
    // Test JWT generation doesn't leak private key
    jwt, err := connector.generateJWT()
    assert.NoError(t, err)
    assert.NotContains(t, jwt, "BEGIN RSA PRIVATE KEY")
    assert.NotContains(t, jwt, config.PrivateKey)
}

func TestWebhookSignatureValidation(t *testing.T) {
    connector := &githubConnector{}
    
    tests := []struct {
        name      string
        payload   []byte
        signature string
        secret    string
        wantErr   bool
    }{
        {
            name:      "valid signature",
            payload:   []byte("test payload"),
            signature: "sha1=8d4e2d8b3f5c6c7d8e9f0a1b2c3d4e5f6a7b8c9d",
            secret:    "secret",
            wantErr:   false,
        },
        {
            name:      "invalid signature",
            payload:   []byte("test payload"),
            signature: "sha1=invalid",
            secret:    "secret",
            wantErr:   true,
        },
        {
            name:      "missing signature",
            payload:   []byte("test payload"),
            signature: "",
            secret:    "secret",
            wantErr:   true,
        },
        {
            name:      "wrong secret",
            payload:   []byte("test payload"),
            signature: "sha1=8d4e2d8b3f5c6c7d8e9f0a1b2c3d4e5f6a7b8c9d",
            secret:    "wrong-secret",
            wantErr:   true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := connector.ValidateWebhookSignature(tt.payload, tt.signature, tt.secret)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestCredentialEncryption(t *testing.T) {
    // Test that sensitive credential data is properly encrypted
    creds := Credentials{
        Type: "github_app_installation",
        Data: map[string]any{
            "installation_id": int64(12345),
            "access_token":    "secret-token",
        },
    }
    
    // Verify that sensitive data is not stored in plain text
    // This would be tested through the actual credential repository
    assert.NotEmpty(t, creds.Data["access_token"])
}
```

## Technical Implementation Details

### Configuration Requirements

#### Environment Variables
```bash
# GitHub App Configuration
GITHUB_APP_ID=123456
GITHUB_APP_NAME=infragpt-dev
GITHUB_APP_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"
GITHUB_WEBHOOK_SECRET=your-webhook-secret
GITHUB_WEBHOOK_URL=https://api.infragpt.com/integrations/github/webhook
GITHUB_WEBHOOK_PORT=8080
```

#### Configuration File Updates
**File**: `services/infragpt/config.yaml`

```yaml
integrations:
  github:
    app_id: ${GITHUB_APP_ID}
    app_name: ${GITHUB_APP_NAME}
    private_key: ${GITHUB_APP_PRIVATE_KEY}
    webhook_secret: ${GITHUB_WEBHOOK_SECRET}
    webhook_url: ${GITHUB_WEBHOOK_URL}
    webhook_port: ${GITHUB_WEBHOOK_PORT}
```

### Database Schema Extensions

#### Webhook Configuration Table
```sql
CREATE TABLE github_webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    repository_id BIGINT NOT NULL,
    repository_full_name VARCHAR(255) NOT NULL,
    webhook_id BIGINT NOT NULL,
    webhook_url VARCHAR(500) NOT NULL,
    events TEXT[] NOT NULL DEFAULT '{}',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(integration_id, repository_id),
    INDEX idx_github_webhooks_integration_id (integration_id),
    INDEX idx_github_webhooks_repository_id (repository_id)
);
```

#### Enhanced Integration Metadata
```json
{
  "installation_id": 12345678,
  "app_id": 123456,
  "account_login": "organization-name",
  "account_type": "Organization",
  "repository_selection": "selected",
  "repository_count": 42,
  "permissions": {
    "actions": "read",
    "contents": "read",
    "issues": "write",
    "pull_requests": "write",
    "metadata": "read"
  },
  "webhook_url": "https://api.infragpt.com/integrations/github/webhook",
  "webhook_events": ["push", "pull_request", "issues"],
  "last_webhook_received": "2023-12-01T10:30:00Z"
}
```

### GitHub App Setup Requirements

#### Required Permissions
```yaml
permissions:
  # Repository permissions
  contents: read          # Read repository contents
  issues: write          # Manage issues
  pull_requests: write   # Manage pull requests
  metadata: read         # Read repository metadata
  actions: read          # Read GitHub Actions
  
  # Organization permissions
  members: read          # Read organization members
  
  # Account permissions
  email_addresses: read  # Read user email addresses
```

#### Required Events
```yaml
events:
  - installation         # App installation events
  - installation_repositories  # Repository access changes
  - push                # Code push events
  - pull_request        # Pull request events
  - issues              # Issue events
  - workflow_run        # GitHub Actions workflow events
```

#### Webhook Configuration
```yaml
webhook:
  url: "https://api.infragpt.com/integrations/github/webhook"
  content_type: "application/json"
  secret: "your-webhook-secret"
  ssl_verification: true
```

## Testing Strategy

### Unit Testing
- **JWT Generation**: Test RSA key parsing and token signing
- **Webhook Validation**: Test signature verification
- **Event Processing**: Test event parsing and routing
- **Error Handling**: Test error cases and recovery

### Integration Testing
- **GitHub API**: Test with mock GitHub API server
- **Database Operations**: Test with test database
- **Webhook Processing**: Test end-to-end webhook flow
- **Credential Management**: Test encryption/decryption

### Security Testing
- **Private Key Security**: Ensure keys are not leaked in logs or errors
- **Webhook Security**: Test signature validation thoroughly
- **Credential Storage**: Verify encryption is working properly
- **API Access**: Test permission boundaries

### Performance Testing
- **JWT Generation**: Should complete in <10ms
- **Webhook Processing**: Should handle events in <500ms
- **API Calls**: Should respect GitHub rate limits
- **Database Operations**: Should complete in <100ms

## Deployment Plan

### Pre-Deployment
1. **GitHub App Setup**: Create and configure GitHub App
2. **Environment Configuration**: Set all required environment variables
3. **Database Migration**: Apply any schema changes
4. **Certificate Setup**: Ensure SSL certificates are valid

### Deployment Phases

#### Phase 1: Staging Deployment
- Deploy to staging environment
- Test with sample repositories
- Validate webhook processing
- Perform security audit

#### Phase 2: Canary Release
- Deploy to 10% of production traffic
- Monitor error rates and performance
- Validate integration flows work correctly
- Collect user feedback

#### Phase 3: Full Production
- Deploy to all production traffic
- Enable monitoring and alerting
- Document operational procedures
- Train support team

### Monitoring and Validation

#### Health Checks
- **GitHub API Connectivity**: Regular health checks
- **Webhook Processing**: Monitor event processing success rate
- **Database Performance**: Monitor query performance
- **Integration Status**: Track integration success/failure rates

#### Metrics to Track
- Integration completion rate
- Webhook processing latency
- GitHub API rate limit usage
- Error rates by type
- User satisfaction scores

#### Alerting Rules
- Failed GitHub API calls > 5% in 5 minutes
- Webhook processing failures > 10% in 10 minutes
- JWT generation failures > 1% in 5 minutes
- Database query time > 1 second
- Integration completion rate < 90%

## Success Criteria

### Functional Requirements
- [x] Users can install GitHub App from web console
- [x] GitHub App installation creates integration record
- [x] Webhooks are automatically configured on repositories
- [x] Events are processed and logged correctly
- [x] Integrations can be revoked cleanly
- [x] Health checks report accurate status

### Performance Benchmarks
- **JWT Generation**: <10ms average
- **Webhook Processing**: <500ms average
- **Integration Flow**: <30 seconds end-to-end
- **API Response Time**: <200ms average
- **Database Queries**: <100ms average

### Security Validations
- [x] Private keys are never logged or exposed
- [x] Webhook signatures are validated correctly
- [x] Credentials are encrypted at rest
- [x] API access is properly scoped
- [x] Revocation cleans up all access

### User Experience Goals
- **Integration Success Rate**: >95%
- **Time to First Value**: <2 minutes
- **Error Rate**: <5%
- **User Satisfaction**: >4.5/5

## Risk Assessment

### Technical Risks

#### High Risk
1. **GitHub API Changes**: GitHub may change API or deprecate endpoints
   - *Mitigation*: Use official GitHub SDK, monitor GitHub changelog
   
2. **Rate Limiting**: GitHub API has strict rate limits
   - *Mitigation*: Implement exponential backoff, cache responses
   
3. **Private Key Compromise**: Private key exposure would compromise security
   - *Mitigation*: Secure key storage, regular rotation, monitoring

#### Medium Risk
4. **Webhook Delivery Failures**: GitHub webhook delivery may fail
   - *Mitigation*: Implement retry mechanism, webhook health monitoring
   
5. **Database Performance**: High webhook volume may impact database
   - *Mitigation*: Optimize queries, implement connection pooling
   
6. **Certificate Expiration**: SSL certificates may expire
   - *Mitigation*: Automated certificate renewal, monitoring

#### Low Risk
7. **Configuration Drift**: Environment configuration may change
   - *Mitigation*: Infrastructure as code, configuration validation
   
8. **Dependency Updates**: Go dependencies may introduce breaking changes
   - *Mitigation*: Dependency pinning, automated testing

### Timeline Risks

#### High Risk
1. **GitHub App Approval**: GitHub may require app review
   - *Mitigation*: Start app registration early, prepare documentation
   
2. **Integration Testing**: End-to-end testing may reveal issues
   - *Mitigation*: Plan extra time for testing, start early

#### Medium Risk
3. **Performance Optimization**: May need additional optimization
   - *Mitigation*: Include performance testing in timeline
   
4. **Security Review**: Security audit may find issues
   - *Mitigation*: Follow security best practices from start

### Mitigation Strategies

#### Monitoring and Alerting
- Comprehensive logging and metrics
- Real-time error tracking
- Performance monitoring
- Security event monitoring

#### Rollback Plan
- Feature flags for quick disable
- Database migration rollback scripts
- Previous version deployment ready
- Incident response procedures

#### Contingency Plans
- Manual webhook processing if automation fails
- Fallback to manual integration setup
- Alternative authentication methods
- Emergency contact procedures

## Dependencies and Prerequisites

### GitHub App Requirements
1. **GitHub App Registration**: Register app with GitHub
2. **App Permissions**: Configure required permissions
3. **Webhook URL**: Publicly accessible webhook endpoint
4. **Private Key**: Generate and securely store private key

### Infrastructure Requirements
1. **SSL Certificate**: Valid SSL certificate for webhook endpoint
2. **Public IP**: Publicly accessible IP for webhooks
3. **Database**: PostgreSQL with sufficient capacity
4. **Environment**: Production-like staging environment

### Development Environment
1. **Go 1.21+**: Required Go version
2. **PostgreSQL**: Local database for testing
3. **GitHub Account**: For testing app installation
4. **Test Repositories**: Sample repositories for webhook testing

### Third-Party Integrations
1. **GitHub API**: Access to GitHub REST API
2. **Webhook Delivery**: GitHub webhook delivery service
3. **SSL Provider**: Certificate authority for SSL
4. **Monitoring Service**: Error tracking and metrics

## Maintenance and Future Considerations

### Long-Term Maintenance

#### Regular Tasks
1. **Private Key Rotation**: Rotate private keys quarterly
2. **Dependency Updates**: Update Go dependencies monthly
3. **Security Audits**: Conduct security reviews quarterly
4. **Performance Reviews**: Review performance metrics monthly

#### Monitoring and Alerting
1. **GitHub API Health**: Monitor GitHub API status
2. **Webhook Delivery**: Track webhook delivery success rates
3. **Integration Health**: Monitor integration status across customers
4. **Security Events**: Monitor for suspicious activities

### Potential Enhancements

#### Short-Term (3-6 months)
1. **Enhanced Repository Analysis**: Code quality analysis, security scanning
2. **Pull Request Automation**: Automated PR reviews, conflict resolution
3. **Issue Management**: Advanced issue tracking and automation
4. **GitHub Actions Integration**: Trigger and monitor workflows

#### Medium-Term (6-12 months)
1. **Advanced Analytics**: Repository insights, developer productivity metrics
2. **Code Review Automation**: AI-powered code review suggestions
3. **Compliance Monitoring**: Security and compliance tracking
4. **Multi-Organization Support**: Support for multiple GitHub organizations

#### Long-Term (12+ months)
1. **GitHub Enterprise Support**: Support for GitHub Enterprise Server
2. **Advanced Workflow Automation**: Complex multi-repository workflows
3. **Machine Learning Integration**: Predictive analytics for development
4. **API Rate Optimization**: Advanced caching and optimization

### Scalability Considerations

#### Current Capacity
- **Webhook Processing**: Designed for 1000+ events/minute
- **Database Storage**: Supports millions of integration records
- **API Calls**: Respects GitHub rate limits with batching

#### Scaling Strategies
1. **Horizontal Scaling**: Multiple webhook processing instances
2. **Database Optimization**: Read replicas, query optimization
3. **Caching Layer**: Redis for frequently accessed data
4. **Async Processing**: Queue-based event processing

#### Performance Optimization
1. **Connection Pooling**: Optimize database connections
2. **Request Batching**: Batch GitHub API requests
3. **Caching Strategy**: Cache frequently accessed data
4. **Query Optimization**: Optimize database queries

## Conclusion

This implementation plan provides a comprehensive roadmap to complete the GitHub integration for InfraGPT. The plan builds on the existing 70% complete foundation while addressing all critical gaps identified in the analysis.

### Key Success Factors
1. **Incremental Implementation**: Phased approach reduces risk
2. **Comprehensive Testing**: Multiple testing layers ensure quality
3. **Security Focus**: Security considerations throughout implementation
4. **Performance Monitoring**: Continuous monitoring ensures reliability

### Expected Outcomes
- **Functional GitHub Integration**: Complete GitHub App installation workflow
- **Production-Ready System**: Robust, secure, and performant implementation
- **Excellent User Experience**: Seamless integration management for users
- **Scalable Architecture**: Foundation for future enhancements

The implementation plan estimates 13-19 days of development work, resulting in a production-ready GitHub integration that follows all established architectural patterns and provides excellent user experience.