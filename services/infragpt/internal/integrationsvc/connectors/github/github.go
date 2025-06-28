package github

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/domain"
)

type GitHubConnector interface {
	ClaimInstallation(ctx context.Context, installationID string, organizationID, userID uuid.UUID) (*infragpt.Integration, error)
}

type githubConnector struct {
	config     Config
	client     *http.Client
	privateKey *rsa.PrivateKey
}

func (g *githubConnector) InitiateAuthorization(organizationID string, userID string) (infragpt.IntegrationAuthorizationIntent, error) {
	stateData := map[string]interface{}{
		"organization_id": organizationID,
		"user_id":         userID,
		"timestamp":       time.Now().Unix(),
	}

	stateJSON, err := json.Marshal(stateData)
	if err != nil {
		return infragpt.IntegrationAuthorizationIntent{}, fmt.Errorf("failed to marshal state data: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(stateJSON)

	params := url.Values{}
	params.Set("state", state)

	installURL := fmt.Sprintf("https://github.com/apps/%s/installations/new?%s", g.config.AppName, params.Encode())

	return infragpt.IntegrationAuthorizationIntent{
		Type: infragpt.AuthorizationTypeInstallation,
		URL:  installURL,
	}, nil
}

func (g *githubConnector) ParseState(state string) (organizationID uuid.UUID, userID uuid.UUID, err error) {
	stateJSON, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid state format, failed to decode base64: %w", err)
	}

	var stateData map[string]interface{}
	if err := json.Unmarshal(stateJSON, &stateData); err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid state format, failed to parse JSON: %w", err)
	}

	orgID, exists := stateData["organization_id"]
	if !exists {
		return uuid.Nil, uuid.Nil, fmt.Errorf("organization_id not found in state")
	}
	orgIDStr, ok := orgID.(string)
	if !ok {
		return uuid.Nil, uuid.Nil, fmt.Errorf("organization_id must be a string")
	}
	organizationID, err = uuid.Parse(orgIDStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid organization_id format: %w", err)
	}

	uID, exists := stateData["user_id"]
	if !exists {
		return uuid.Nil, uuid.Nil, fmt.Errorf("user_id not found in state")
	}
	userIDStr, ok := uID.(string)
	if !ok {
		return uuid.Nil, uuid.Nil, fmt.Errorf("user_id must be a string")
	}
	userID, err = uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid user_id format: %w", err)
	}

	return organizationID, userID, nil
}

func (g *githubConnector) CompleteAuthorization(authData infragpt.AuthorizationData) (infragpt.Credentials, error) {
	if authData.InstallationID == "" {
		return infragpt.Credentials{}, fmt.Errorf("installation ID is required for GitHub App")
	}

	// Parse state to get organization ID and user ID for claiming
	organizationID, userID, err := g.ParseState(authData.State)
	if err != nil {
		return infragpt.Credentials{}, fmt.Errorf("failed to parse state: %w", err)
	}

	// Try to claim installation if it exists as unclaimed
	ctx := context.Background()
	integration, err := g.ClaimInstallation(ctx, authData.InstallationID, organizationID, userID)
	if err != nil {
		slog.Error("failed to claim unclaimed GitHub installation",
			"installation_id", authData.InstallationID,
			"organization_id", organizationID,
			"integration_id", integration.ID,
			"error", err)
		return infragpt.Credentials{}, fmt.Errorf("failed to claim unclaimed GitHub installation: %w", err)
	}

	return infragpt.Credentials{
		Type: infragpt.CredentialTypeToken,
		Data: map[string]string{
			"installation_id": authData.InstallationID,
			"claimed":         "true",
		},
		OrganizationInfo: &infragpt.OrganizationInfo{
			ExternalID: integration.ConnectorOrganizationID,
			Name:       integration.ConnectorUserID,
			Metadata:   integration.Metadata,
		},
	}, nil
}

func (g *githubConnector) ValidateCredentials(creds infragpt.Credentials) error {
	installationID, exists := creds.Data["installation_id"]
	if !exists {
		return fmt.Errorf("installation ID not found in credentials")
	}

	jwt, err := g.generateJWT()
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}

	_, err = g.getInstallationDetails(jwt, installationID)
	if err != nil {
		return fmt.Errorf("installation validation failed: %w", err)
	}

	return nil
}

func (g *githubConnector) RefreshCredentials(creds infragpt.Credentials) (infragpt.Credentials, error) {
	installationID, exists := creds.Data["installation_id"]
	if !exists {
		return infragpt.Credentials{}, fmt.Errorf("installation ID not found in credentials")
	}

	jwt, err := g.generateJWT()
	if err != nil {
		return infragpt.Credentials{}, fmt.Errorf("failed to generate JWT: %w", err)
	}

	accessToken, err := g.getInstallationAccessToken(jwt, installationID)
	if err != nil {
		return infragpt.Credentials{}, fmt.Errorf("failed to refresh access token: %w", err)
	}

	newCreds := creds
	newCreds.Data["access_token"] = accessToken.Token
	if !accessToken.ExpiresAt.IsZero() {
		newCreds.ExpiresAt = &accessToken.ExpiresAt
	}

	return newCreds, nil
}

func (g *githubConnector) RevokeCredentials(creds infragpt.Credentials) error {
	// Simple stub implementation - just log the revocation
	installationID, exists := creds.Data["installation_id"]
	if !exists {
		return fmt.Errorf("installation ID not found in credentials")
	}

	slog.Info("GitHub credentials revoked", "installation_id", installationID)
	return nil
}

func (g *githubConnector) ConfigureWebhooks(integrationID string, creds infragpt.Credentials) error {
	// Simple stub implementation for installation webhook only
	installationID, exists := creds.Data["installation_id"]
	if !exists {
		return fmt.Errorf("installation ID not found in credentials")
	}

	webhookURL := g.buildWebhookURL(integrationID)
	if webhookURL == "" {
		return fmt.Errorf("webhook URL configuration missing: redirect_url is required")
	}

	slog.Info("GitHub installation webhook configured",
		"installation_id", installationID,
		"webhook_url", webhookURL)
	return nil
}

func (g *githubConnector) generateJWT() (string, error) {
	if g.privateKey == nil {
		return "", fmt.Errorf("GitHub private key not configured or failed to parse - check private_key configuration")
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"iss": g.config.AppID,
	})

	tokenString, err := token.SignedString(g.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, nil
}

func (g *githubConnector) getInstallationAccessToken(jwt string, installationID string) (*accessTokenResponse, error) {
	url := fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens", installationID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	var response accessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode access token response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("GitHub API error: %s", response.Message)
	}

	return &response, nil
}

func (g *githubConnector) getInstallationDetails(jwt string, installationID string) (*installationResponse, error) {
	url := fmt.Sprintf("https://api.github.com/app/installations/%s", installationID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get installation details: %w", err)
	}
	defer resp.Body.Close()

	var response installationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode installation response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	return &response, nil
}

func (g *githubConnector) formatPermissions(perms map[string]string) string {
	var parts []string
	for key, value := range perms {
		parts = append(parts, fmt.Sprintf("%s:%s", key, value))
	}
	return strings.Join(parts, ",")
}

func (g *githubConnector) buildWebhookURL(integrationID string) string {
	baseURL := g.config.RedirectURL
	if baseURL == "" {
		return ""
	}

	baseURL = strings.TrimSuffix(baseURL, "/")
	return fmt.Sprintf("%s/webhooks/github", baseURL)
}

func (g *githubConnector) ClaimInstallation(ctx context.Context, installationID string, organizationID, userID uuid.UUID) (*infragpt.Integration, error) {
	// Generate JWT to access GitHub API
	jwt, err := g.generateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	// Get installation details directly from GitHub API
	installationDetails, err := g.getInstallationDetails(jwt, installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get installation details from GitHub: %w", err)
	}

	// Create integration record with data from GitHub API
	connectorOrgID := strconv.FormatInt(installationDetails.Account.ID, 10)
	integration := &infragpt.Integration{
		ID:                      uuid.New(),
		OrganizationID:          organizationID,
		UserID:                  userID,
		ConnectorType:           infragpt.ConnectorTypeGithub,
		Status:                  infragpt.IntegrationStatusActive,
		BotID:                   installationID,
		ConnectorUserID:         installationDetails.Account.Login,
		ConnectorOrganizationID: connectorOrgID,
		Metadata: map[string]string{
			"github_installation_id": installationID,
			"github_app_id":          g.config.AppID,
			"github_account_id":      strconv.FormatInt(installationDetails.Account.ID, 10),
			"github_account_login":   installationDetails.Account.Login,
			"github_account_type":    installationDetails.Account.Type,
			"target_type":            installationDetails.TargetType,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store integration
	if err := g.config.IntegrationRepository.Store(ctx, *integration); err != nil {
		return nil, fmt.Errorf("failed to store integration: %w", err)
	}

	// Generate and store credentials
	accessToken, err := g.getInstallationAccessToken(jwt, installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	credentialData := map[string]string{
		"installation_id": installationID,
		"access_token":    accessToken.Token,
		"account_login":   installationDetails.Account.Login,
		"account_id":      strconv.FormatInt(installationDetails.Account.ID, 10),
		"account_type":    installationDetails.Account.Type,
	}

	var expiresAt *time.Time
	if !accessToken.ExpiresAt.IsZero() {
		expiresAt = &accessToken.ExpiresAt
	}

	credentialRecord := domain.IntegrationCredential{
		ID:              uuid.New(),
		IntegrationID:   integration.ID,
		CredentialType:  infragpt.CredentialTypeToken,
		Data:            credentialData,
		ExpiresAt:       expiresAt,
		EncryptionKeyID: "v1",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := g.config.CredentialRepository.Store(ctx, credentialRecord); err != nil {
		return nil, fmt.Errorf("failed to store credentials: %w", err)
	}

	// Sync repositories
	if err := g.syncRepositories(ctx, integration.ID, installationID); err != nil {
		slog.Error("failed to sync repositories during installation claim",
			"integration_id", integration.ID,
			"installation_id", installationID,
			"error", err)
	}

	return integration, nil
}

func (g *githubConnector) syncRepositories(ctx context.Context, integrationID uuid.UUID, installationID string) error {
	slog.Info("syncing repositories",
		"integration_id", integrationID,
		"installation_id", installationID)

	// Generate JWT and get installation access token
	jwt, err := g.generateJWT()
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}

	accessToken, err := g.getInstallationAccessToken(jwt, installationID)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Fetch repositories from GitHub API
	repositories, err := g.fetchInstallationRepositories(accessToken.Token)
	if err != nil {
		return fmt.Errorf("failed to fetch repositories: %w", err)
	}

	slog.Info("fetched repositories from GitHub",
		"integration_id", integrationID,
		"repository_count", len(repositories))

	// Store repositories in database
	for _, repo := range repositories {
		githubRepo := GitHubRepository{
			ID:                    uuid.New(),
			IntegrationID:         integrationID,
			GitHubRepositoryID:    repo.ID,
			RepositoryName:        repo.Name,
			RepositoryFullName:    repo.FullName,
			RepositoryURL:         repo.HTMLURL,
			IsPrivate:             repo.Private,
			DefaultBranch:         repo.DefaultBranch,
			PermissionAdmin:       false, // TODO: Extract from API response
			PermissionPush:        false, // TODO: Extract from API response
			PermissionPull:        true,  // Default permission for installations
			RepositoryDescription: repo.Description,
			RepositoryLanguage:    repo.Language,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
			LastSyncedAt:          time.Now(),
			GitHubCreatedAt:       repo.CreatedAt,
			GitHubUpdatedAt:       repo.UpdatedAt,
			GitHubPushedAt:        repo.PushedAt,
		}

		if err := g.config.GitHubRepositoryRepo.Store(ctx, githubRepo); err != nil {
			slog.Error("failed to store repository",
				"integration_id", integrationID,
				"repository_id", repo.ID,
				"repository_name", repo.FullName,
				"error", err)
			continue
		}
	}

	// Update last sync time
	if err := g.config.GitHubRepositoryRepo.UpdateLastSyncTime(ctx, integrationID, time.Now()); err != nil {
		slog.Error("failed to update last sync time", "integration_id", integrationID, "error", err)
	}

	return nil
}

func (g *githubConnector) addRepositories(ctx context.Context, integrationID uuid.UUID, repositories []Repository) error {
	slog.Info("adding repositories",
		"integration_id", integrationID,
		"repository_count", len(repositories))

	for _, repo := range repositories {
		githubRepo := GitHubRepository{
			ID:                    uuid.New(),
			IntegrationID:         integrationID,
			GitHubRepositoryID:    repo.ID,
			RepositoryName:        repo.Name,
			RepositoryFullName:    repo.FullName,
			RepositoryURL:         repo.HTMLURL,
			IsPrivate:             repo.Private,
			DefaultBranch:         repo.DefaultBranch,
			PermissionAdmin:       false,
			PermissionPush:        false,
			PermissionPull:        true,
			RepositoryDescription: repo.Description,
			RepositoryLanguage:    repo.Language,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
			LastSyncedAt:          time.Now(),
			GitHubCreatedAt:       repo.CreatedAt,
			GitHubUpdatedAt:       repo.UpdatedAt,
			GitHubPushedAt:        repo.PushedAt,
		}

		if err := g.config.GitHubRepositoryRepo.Store(ctx, githubRepo); err != nil {
			slog.Error("failed to add repository",
				"integration_id", integrationID,
				"repository_id", repo.ID,
				"repository_name", repo.FullName,
				"error", err)
			continue
		}
	}

	return nil
}

func (g *githubConnector) removeRepositories(ctx context.Context, integrationID uuid.UUID, repositoryIDs []int64) error {
	slog.Info("removing repositories",
		"integration_id", integrationID,
		"repository_count", len(repositoryIDs))

	if err := g.config.GitHubRepositoryRepo.BulkDelete(ctx, integrationID, repositoryIDs); err != nil {
		return fmt.Errorf("failed to bulk delete repositories: %w", err)
	}

	return nil
}

func (g *githubConnector) fetchInstallationRepositories(accessToken string) ([]Repository, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/installation/repositories", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	var response struct {
		TotalCount   int          `json:"total_count"`
		Repositories []Repository `json:"repositories"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode repositories response: %w", err)
	}

	return response.Repositories, nil
}

type accessTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Message   string    `json:"message,omitempty"`
}

type installationResponse struct {
	ID          int64             `json:"id"`
	Account     accountResponse   `json:"account"`
	TargetType  string            `json:"target_type"`
	Permissions map[string]string `json:"permissions"`
}

type accountResponse struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Type  string `json:"type"`
}

func (g *githubConnector) Sync(ctx context.Context, integration infragpt.Integration, params map[string]string) error {
	// Sync repositories - check for new repositories and update existing ones
	if err := g.syncRepositoriesForIntegration(ctx, integration); err != nil {
		return fmt.Errorf("failed to sync repositories: %w", err)
	}

	// Check and update repository permissions and status
	if err := g.syncRepositoryPermissions(ctx, integration); err != nil {
		return fmt.Errorf("failed to sync repository permissions: %w", err)
	}

	return nil
}

func (g *githubConnector) syncRepositoryPermissions(ctx context.Context, integration infragpt.Integration) error {
	integrationUUID := integration.ID

	installationID := integration.BotID
	if installationID == "" {
		return fmt.Errorf("installation ID not found in integration")
	}

	// Generate JWT and get installation access token
	jwt, err := g.generateJWT()
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}

	accessToken, err := g.getInstallationAccessToken(jwt, installationID)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Fetch current installation details to check permissions
	installationDetails, err := g.getInstallationDetails(jwt, installationID)
	if err != nil {
		return fmt.Errorf("failed to get installation details: %w", err)
	}

	// Fetch repositories from GitHub API with updated permissions
	repositories, err := g.fetchInstallationRepositories(accessToken.Token)
	if err != nil {
		return fmt.Errorf("failed to fetch repositories: %w", err)
	}

	// Update existing repositories with current permissions and status
	// For GitHub App installations, all repositories have the same permissions as defined in the installation
	defaultPermissions := RepositoryPermissions{
		Admin: false, // Apps typically don't get admin access
		Push:  true,  // Most installations need push access
		Pull:  true,  // All installations need pull access
	}

	for _, repo := range repositories {
		if err := g.config.GitHubRepositoryRepo.UpdatePermissions(ctx, integrationUUID, repo.ID, defaultPermissions); err != nil {
			slog.Error("failed to update repository permissions",
				"integration_id", integration.ID,
				"repository_id", repo.ID,
				"repository_name", repo.FullName,
				"error", err)
			continue
		}
	}

	slog.Info("synced repository permissions",
		"integration_id", integration.ID,
		"installation_id", installationID,
		"repository_count", len(repositories),
		"permissions", g.formatPermissions(installationDetails.Permissions))

	return nil
}

func (g *githubConnector) syncInstallation(ctx context.Context, integration infragpt.Integration, params map[string]string) error {
	installationID := params["installation_id"]
	if installationID == "" {
		installationID = integration.BotID
	}
	if installationID == "" {
		return fmt.Errorf("installation_id is required for GitHub installation sync")
	}

	// Simply sync repositories for the existing integration
	return g.syncRepositoriesForIntegration(ctx, integration)
}

func (g *githubConnector) syncRepositoriesForIntegration(ctx context.Context, integration infragpt.Integration) error {
	integrationUUID := integration.ID

	installationID := integration.BotID
	if installationID == "" {
		return fmt.Errorf("installation ID not found in integration")
	}

	return g.syncRepositories(ctx, integrationUUID, installationID)
}
