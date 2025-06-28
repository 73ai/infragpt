package github

import (
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
	"github.com/priyanshujain/infragpt/services/infragpt"
)

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
	if g.config.RedirectURL != "" {
		params.Set("redirect_url", g.config.RedirectURL)
	}

	installURL := fmt.Sprintf("https://github.com/apps/%s/installations/new?%s", g.config.AppID, params.Encode())

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

	installationID, err := strconv.ParseInt(authData.InstallationID, 10, 64)
	if err != nil {
		return infragpt.Credentials{}, fmt.Errorf("invalid installation ID: %w", err)
	}

	jwt, err := g.generateJWT()
	if err != nil {
		return infragpt.Credentials{}, fmt.Errorf("failed to generate JWT: %w", err)
	}

	accessToken, err := g.getInstallationAccessToken(jwt, installationID)
	if err != nil {
		return infragpt.Credentials{}, fmt.Errorf("failed to get installation access token: %w", err)
	}

	installationDetails, err := g.getInstallationDetails(jwt, installationID)
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

	installationIDInt, err := strconv.ParseInt(installationID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid installation ID: %w", err)
	}

	_, err = g.getInstallationDetails(jwt, installationIDInt)
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

	installationIDInt, err := strconv.ParseInt(installationID, 10, 64)
	if err != nil {
		return infragpt.Credentials{}, fmt.Errorf("invalid installation ID: %w", err)
	}

	jwt, err := g.generateJWT()
	if err != nil {
		return infragpt.Credentials{}, fmt.Errorf("failed to generate JWT: %w", err)
	}

	accessToken, err := g.getInstallationAccessToken(jwt, installationIDInt)
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
		return "", fmt.Errorf("private key not configured")
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

func (g *githubConnector) getInstallationAccessToken(jwt string, installationID int64) (*accessTokenResponse, error) {
	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID)

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

func (g *githubConnector) getInstallationDetails(jwt string, installationID int64) (*installationResponse, error) {
	url := fmt.Sprintf("https://api.github.com/app/installations/%d", installationID)

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


// Type definitions for GitHub API responses

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

