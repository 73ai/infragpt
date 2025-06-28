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
	ClaimInstallation(ctx context.Context, installationID string, organizationID, userID string) (*infragpt.Integration, error)
	GetUnclaimedInstallations(ctx context.Context) ([]UnclaimedInstallation, error)
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

func (g *githubConnector) ParseState(state string) (organizationID string, userID string, err error) {
	stateJSON, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return "", "", fmt.Errorf("invalid state format, failed to decode base64: %w", err)
	}

	var stateData map[string]interface{}
	if err := json.Unmarshal(stateJSON, &stateData); err != nil {
		return "", "", fmt.Errorf("invalid state format, failed to parse JSON: %w", err)
	}

	orgID, exists := stateData["organization_id"]
	if !exists {
		return "", "", fmt.Errorf("organization_id not found in state")
	}
	organizationID, ok := orgID.(string)
	if !ok {
		return "", "", fmt.Errorf("organization_id must be a string")
	}

	uID, exists := stateData["user_id"]
	if !exists {
		return "", "", fmt.Errorf("user_id not found in state")
	}
	userID, ok = uID.(string)
	if !ok {
		return "", "", fmt.Errorf("user_id must be a string")
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
	if err == nil {
		slog.Info("automatically claimed unclaimed GitHub installation during authorization",
			"installation_id", authData.InstallationID,
			"organization_id", organizationID,
			"integration_id", integration.ID)

		// Return credentials with organization info to indicate successful claiming
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

	// If claiming fails, continue with normal authorization flow
	slog.Debug("could not claim unclaimed installation, proceeding with normal authorization flow",
		"installation_id", authData.InstallationID,
		"error", err)

	jwt, err := g.generateJWT()
	if err != nil {
		return infragpt.Credentials{}, fmt.Errorf("failed to generate JWT: %w", err)
	}

	accessToken, err := g.getInstallationAccessToken(jwt, authData.InstallationID)
	if err != nil {
		return infragpt.Credentials{}, fmt.Errorf("failed to get installation access token: %w", err)
	}

	installationDetails, err := g.getInstallationDetails(jwt, authData.InstallationID)
	if err != nil {
		return infragpt.Credentials{}, fmt.Errorf("failed to get installation details: %w", err)
	}

	credentialData := map[string]string{
		"installation_id": authData.InstallationID,
		"access_token":    accessToken.Token,
		"account_login":   installationDetails.Account.Login,
		"account_id":      strconv.FormatInt(installationDetails.Account.ID, 10),
		"account_type":    installationDetails.Account.Type,
		"target_type":     installationDetails.TargetType,
		"permissions":     g.formatPermissions(installationDetails.Permissions),
	}

	var expiresAt *time.Time
	if !accessToken.ExpiresAt.IsZero() {
		expiresAt = &accessToken.ExpiresAt
	}

	return infragpt.Credentials{
		Type:      infragpt.CredentialTypeToken,
		Data:      credentialData,
		ExpiresAt: expiresAt,
		OrganizationInfo: &infragpt.OrganizationInfo{
			ExternalID: strconv.FormatInt(installationDetails.Account.ID, 10),
			Name:       installationDetails.Account.Login,
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

func (g *githubConnector) ClaimInstallation(ctx context.Context, installationID string, organizationID, userID string) (*infragpt.Integration, error) {
	// Get unclaimed installation from database
	unclaimed, err := g.config.UnclaimedInstallationRepo.GetByInstallationID(ctx, installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unclaimed installation: %w", err)
	}
	if unclaimed.ID == uuid.Nil {
		return nil, fmt.Errorf("unclaimed installation not found for ID %s", installationID)
	}

	// Create integration record
	connectorOrgID := strconv.FormatInt(unclaimed.GitHubAccountID, 10)
	integration := &infragpt.Integration{
		ID:                      uuid.New().String(),
		OrganizationID:          organizationID,
		UserID:                  userID,
		ConnectorType:           infragpt.ConnectorTypeGithub,
		Status:                  infragpt.IntegrationStatusActive,
		BotID:                   installationID,
		ConnectorUserID:         unclaimed.GitHubAccountLogin,
		ConnectorOrganizationID: connectorOrgID,
		Metadata: map[string]string{
			"github_installation_id": unclaimed.GitHubInstallationID,
			"github_app_id":          strconv.FormatInt(unclaimed.GitHubAppID, 10),
			"github_account_id":      strconv.FormatInt(unclaimed.GitHubAccountID, 10),
			"github_account_login":   unclaimed.GitHubAccountLogin,
			"github_account_type":    unclaimed.GitHubAccountType,
			"repository_selection":   unclaimed.RepositorySelection,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store integration
	if err := g.config.IntegrationRepository.Store(ctx, *integration); err != nil {
		return nil, fmt.Errorf("failed to store integration: %w", err)
	}

	// Generate and store credentials
	jwt, err := g.generateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	accessToken, err := g.getInstallationAccessToken(jwt, installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	credentialData := map[string]string{
		"installation_id": installationID,
		"access_token":    accessToken.Token,
		"account_login":   unclaimed.GitHubAccountLogin,
		"account_id":      strconv.FormatInt(unclaimed.GitHubAccountID, 10),
		"account_type":    unclaimed.GitHubAccountType,
	}

	var expiresAt *time.Time
	if !accessToken.ExpiresAt.IsZero() {
		expiresAt = &accessToken.ExpiresAt
	}

	credentialRecord := domain.IntegrationCredential{
		ID:              uuid.New().String(),
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
	integrationUUID, err := uuid.Parse(integration.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse integration ID: %w", err)
	}

	if err := g.syncRepositories(ctx, integrationUUID, installationID); err != nil {
		slog.Error("failed to sync repositories during installation claim",
			"integration_id", integration.ID,
			"installation_id", installationID,
			"error", err)
	}

	// Mark installation as claimed
	orgUUID := uuid.MustParse(organizationID)
	userUUID := uuid.MustParse(userID)
	if err := g.config.UnclaimedInstallationRepo.MarkAsClaimed(ctx, installationID, orgUUID, userUUID); err != nil {
		slog.Error("failed to mark installation as claimed",
			"installation_id", installationID,
			"error", err)
	}

	return integration, nil
}

func (g *githubConnector) GetUnclaimedInstallations(ctx context.Context) ([]UnclaimedInstallation, error) {
	return g.config.UnclaimedInstallationRepo.List(ctx, 100)
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
	integrationUUID, err := uuid.Parse(integration.ID)
	if err != nil {
		return fmt.Errorf("failed to parse integration ID: %w", err)
	}

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

	// This handles the case where a user is redirected back from GitHub
	// after installing the app and needs to claim the installation
	return g.claimInstallationForExistingIntegration(ctx, integration, installationID)
}

func (g *githubConnector) syncRepositoriesForIntegration(ctx context.Context, integration infragpt.Integration) error {
	integrationUUID, err := uuid.Parse(integration.ID)
	if err != nil {
		return fmt.Errorf("failed to parse integration ID: %w", err)
	}

	installationID := integration.BotID
	if installationID == "" {
		return fmt.Errorf("installation ID not found in integration")
	}

	return g.syncRepositories(ctx, integrationUUID, installationID)
}

func (g *githubConnector) claimInstallationForExistingIntegration(ctx context.Context, integration infragpt.Integration, installationID string) error {
	// Get unclaimed installation from database
	unclaimed, err := g.config.UnclaimedInstallationRepo.GetByInstallationID(ctx, installationID)
	if err != nil {
		return fmt.Errorf("failed to get unclaimed installation: %w", err)
	}
	if unclaimed.ID == uuid.Nil {
		return fmt.Errorf("unclaimed installation not found for ID %s", installationID)
	}

	// Update existing integration with installation details
	connectorOrgID := strconv.FormatInt(unclaimed.GitHubAccountID, 10)
	updatedIntegration := integration
	updatedIntegration.BotID = installationID
	updatedIntegration.ConnectorUserID = unclaimed.GitHubAccountLogin
	updatedIntegration.ConnectorOrganizationID = connectorOrgID
	updatedIntegration.Status = infragpt.IntegrationStatusActive
	updatedIntegration.UpdatedAt = time.Now()

	// Update metadata
	if updatedIntegration.Metadata == nil {
		updatedIntegration.Metadata = make(map[string]string)
	}
	updatedIntegration.Metadata["github_installation_id"] = unclaimed.GitHubInstallationID
	updatedIntegration.Metadata["github_app_id"] = strconv.FormatInt(unclaimed.GitHubAppID, 10)
	updatedIntegration.Metadata["github_account_id"] = strconv.FormatInt(unclaimed.GitHubAccountID, 10)
	updatedIntegration.Metadata["github_account_login"] = unclaimed.GitHubAccountLogin
	updatedIntegration.Metadata["github_account_type"] = unclaimed.GitHubAccountType
	updatedIntegration.Metadata["repository_selection"] = unclaimed.RepositorySelection

	// Update integration in database
	if err := g.config.IntegrationRepository.Update(ctx, updatedIntegration); err != nil {
		return fmt.Errorf("failed to update integration: %w", err)
	}

	// Generate and store/update credentials
	jwt, err := g.generateJWT()
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}

	accessToken, err := g.getInstallationAccessToken(jwt, installationID)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	credentialData := map[string]string{
		"installation_id": installationID,
		"access_token":    accessToken.Token,
		"account_login":   unclaimed.GitHubAccountLogin,
		"account_id":      strconv.FormatInt(unclaimed.GitHubAccountID, 10),
		"account_type":    unclaimed.GitHubAccountType,
	}

	var expiresAt *time.Time
	if !accessToken.ExpiresAt.IsZero() {
		expiresAt = &accessToken.ExpiresAt
	}

	// Try to find existing credential first
	existingCredential, err := g.config.CredentialRepository.FindByIntegration(ctx, integration.ID)
	if err == nil {
		// Update existing credential
		existingCredential.Data = credentialData
		existingCredential.ExpiresAt = expiresAt
		existingCredential.UpdatedAt = time.Now()

		if err := g.config.CredentialRepository.Update(ctx, existingCredential); err != nil {
			return fmt.Errorf("failed to update credentials: %w", err)
		}
	} else {
		// Create new credential
		credentialRecord := domain.IntegrationCredential{
			ID:              uuid.New().String(),
			IntegrationID:   integration.ID,
			CredentialType:  infragpt.CredentialTypeToken,
			Data:            credentialData,
			ExpiresAt:       expiresAt,
			EncryptionKeyID: "v1",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if err := g.config.CredentialRepository.Store(ctx, credentialRecord); err != nil {
			return fmt.Errorf("failed to store credentials: %w", err)
		}
	}

	// Sync repositories
	integrationUUID, err := uuid.Parse(integration.ID)
	if err != nil {
		return fmt.Errorf("failed to parse integration ID: %w", err)
	}

	if err := g.syncRepositories(ctx, integrationUUID, installationID); err != nil {
		slog.Error("failed to sync repositories during installation claim",
			"integration_id", integration.ID,
			"installation_id", installationID,
			"error", err)
	}

	// Mark installation as claimed
	orgUUID := uuid.MustParse(integration.OrganizationID)
	userUUID := uuid.MustParse(integration.UserID)
	if err := g.config.UnclaimedInstallationRepo.MarkAsClaimed(ctx, installationID, orgUUID, userUUID); err != nil {
		slog.Error("failed to mark installation as claimed",
			"installation_id", installationID,
			"error", err)
	}

	return nil
}
