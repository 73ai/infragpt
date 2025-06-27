package github

import (
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
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
	// Create state as base64 encoded JSON
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
	// Decode base64 state
	stateJSON, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return "", "", fmt.Errorf("invalid state format, failed to decode base64: %w", err)
	}

	// Parse JSON state data
	var stateData map[string]interface{}
	if err := json.Unmarshal(stateJSON, &stateData); err != nil {
		return "", "", fmt.Errorf("invalid state format, failed to parse JSON: %w", err)
	}

	// Extract organization_id
	orgID, exists := stateData["organization_id"]
	if !exists {
		return "", "", fmt.Errorf("organization_id not found in state")
	}
	organizationID, ok := orgID.(string)
	if !ok {
		return "", "", fmt.Errorf("organization_id must be a string")
	}

	// Extract user_id
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

func (g *githubConnector) ValidateWebhookSignature(payload []byte, signature string, secret string) error {
	if secret == "" {
		secret = g.config.WebhookSecret
	}

	expectedSignature := g.computeSignature(payload, secret)

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("webhook signature validation failed")
	}

	return nil
}

func (g *githubConnector) ProcessEvent(ctx context.Context, event any) error {
	webhookEvent, ok := event.(WebhookEvent)
	if !ok {
		return fmt.Errorf("invalid event type: expected WebhookEvent")
	}

	// Only handle installation events
	switch webhookEvent.EventType {
	case EventTypeInstallation:
		return g.handleInstallationEvent(ctx, webhookEvent)
	default:
		slog.Debug("ignoring non-installation event",
			"event_type", webhookEvent.EventType,
			"installation_id", webhookEvent.InstallationID)
		return nil
	}
}

func (g *githubConnector) Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error {
	if g.config.WebhookPort == 0 {
		return fmt.Errorf("github: webhook port is required for webhook server")
	}

	webhookConfig := webhookServerConfig{
		port:                g.config.WebhookPort,
		webhookSecret:       g.config.WebhookSecret,
		callbackHandlerFunc: handler,
	}

	return webhookConfig.startWebhookServer(ctx)
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

func (g *githubConnector) computeSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return fmt.Sprintf("sha256=%s", hex.EncodeToString(h.Sum(nil)))
}

func (g *githubConnector) buildWebhookURL(integrationID string) string {
	baseURL := g.config.RedirectURL
	if baseURL == "" {
		return ""
	}
	
	baseURL = strings.TrimSuffix(baseURL, "/")
	return fmt.Sprintf("%s/webhooks/github", baseURL)
}

func (g *githubConnector) handleInstallationEvent(ctx context.Context, event WebhookEvent) error {
	slog.Info("handling GitHub installation event",
		"action", event.InstallationAction,
		"installation_id", event.InstallationID,
		"account_login", event.SenderLogin,
		"repositories_added", len(event.RepositoriesAdded),
		"repositories_removed", len(event.RepositoriesRemoved))

	// Simple event processing - just log the installation event
	switch event.InstallationAction {
	case "created":
		slog.Info("GitHub App installed",
			"installation_id", event.InstallationID,
			"account", event.SenderLogin)
	case "deleted":
		slog.Info("GitHub App uninstalled",
			"installation_id", event.InstallationID,
			"account", event.SenderLogin)
	case "added":
		slog.Info("GitHub App repository access added",
			"installation_id", event.InstallationID,
			"repositories", event.RepositoriesAdded)
	case "removed":
		slog.Info("GitHub App repository access removed",
			"installation_id", event.InstallationID,
			"repositories", event.RepositoriesRemoved)
	}

	return nil
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

// Webhook server configuration and implementation
type webhookServerConfig struct {
	port                int
	webhookSecret       string
	callbackHandlerFunc func(ctx context.Context, event any) error
}

func (c webhookServerConfig) startWebhookServer(ctx context.Context) error {
	h := &webhookHandler{
		callbackHandlerFunc: c.callbackHandlerFunc,
	}
	h.init()

	httpServer := &http.Server{
		Addr:        fmt.Sprintf(":%d", c.port),
		BaseContext: func(net.Listener) context.Context { return ctx },
		Handler:     panicMiddleware(webhookValidationMiddleware(c.webhookSecret, h)),
	}

	return httpServer.ListenAndServe()
}

type webhookHandler struct {
	http.ServeMux
	callbackHandlerFunc func(ctx context.Context, event any) error
}

func (wh *webhookHandler) init() {
	wh.HandleFunc("/webhooks/github", wh.handler())
}

func (wh *webhookHandler) handler() func(w http.ResponseWriter, r *http.Request) {
	type response struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		eventType := r.Header.Get("X-GitHub-Event")
		if eventType == "" {
			http.Error(w, "Missing X-GitHub-Event header", http.StatusBadRequest)
			return
		}

		// Only process installation events
		if eventType != "installation" && eventType != "installation_repositories" {
			slog.Debug("ignoring non-installation event", "event_type", eventType)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response{})
			return
		}

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read payload", http.StatusBadRequest)
			return
		}

		var rawPayload map[string]any
		if err := json.Unmarshal(payload, &rawPayload); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		webhookEvent, err := wh.convertToWebhookEvent(eventType, rawPayload)
		if err != nil {
			slog.Error("failed to convert GitHub webhook event", "event_type", eventType, "error", err)
			http.Error(w, "Failed to process event", http.StatusInternalServerError)
			return
		}

		if err := wh.callbackHandlerFunc(ctx, webhookEvent); err != nil {
			slog.Error("error handling GitHub webhook event", "event_type", eventType, "error", err)
			http.Error(w, "Failed to handle event", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response{})
	}
}

func (wh *webhookHandler) convertToWebhookEvent(eventType string, rawPayload map[string]any) (WebhookEvent, error) {
	event := WebhookEvent{
		EventType:  EventType(eventType),
		RawPayload: rawPayload,
		CreatedAt:  time.Now(),
	}

	// Extract common fields
	if installation, ok := rawPayload["installation"].(map[string]any); ok {
		if id, ok := installation["id"].(float64); ok {
			event.InstallationID = int64(id)
		}
	}

	if sender, ok := rawPayload["sender"].(map[string]any); ok {
		if id, ok := sender["id"].(float64); ok {
			event.SenderID = int64(id)
		}
		if login, ok := sender["login"].(string); ok {
			event.SenderLogin = login
		}
	}

	if action, ok := rawPayload["action"].(string); ok {
		event.Action = action
		event.InstallationAction = action
	}

	// Handle repository changes for installation events
	if eventType == "installation" || eventType == "installation_repositories" {
		if repositories, ok := rawPayload["repositories"].([]any); ok {
			for _, repo := range repositories {
				if repoMap, ok := repo.(map[string]any); ok {
					if fullName, ok := repoMap["full_name"].(string); ok {
						event.RepositoriesAdded = append(event.RepositoriesAdded, fullName)
					}
				}
			}
		}

		if repositoriesRemoved, ok := rawPayload["repositories_removed"].([]any); ok {
			for _, repo := range repositoriesRemoved {
				if repoMap, ok := repo.(map[string]any); ok {
					if fullName, ok := repoMap["full_name"].(string); ok {
						event.RepositoriesRemoved = append(event.RepositoriesRemoved, fullName)
					}
				}
			}
		}
	}

	return event, nil
}

func panicMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("github: panic while handling http request", "recover", r)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func webhookValidationMiddleware(webhookSecret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if webhookSecret == "" {
			next.ServeHTTP(w, r)
			return
		}

		signature := r.Header.Get("X-Hub-Signature-256")
		if signature == "" {
			slog.Info("github: missing webhook signature")
			http.Error(w, "Missing webhook signature", http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		if !validateGitHubSignature(body, signature, webhookSecret) {
			slog.Info("github: webhook validation failed", "signature", signature)
			http.Error(w, "Invalid webhook signature", http.StatusUnauthorized)
			return
		}

		r.Body = io.NopCloser(strings.NewReader(string(body)))
		next.ServeHTTP(w, r)
	})
}

func validateGitHubSignature(payload []byte, signature string, secret string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	expectedHash := strings.TrimPrefix(signature, "sha256=")
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	actualHash := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(expectedHash), []byte(actualHash))
}
