package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

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

	// Handle installation and repository events
	switch webhookEvent.EventType {
	case EventTypeInstallation:
		return g.handleInstallationEvent(ctx, webhookEvent)
	case "installation_repositories":
		return g.handleInstallationRepositoriesEvent(ctx, webhookEvent)
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


func (g *githubConnector) computeSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return fmt.Sprintf("sha256=%s", hex.EncodeToString(h.Sum(nil)))
}

func (g *githubConnector) handleInstallationEvent(ctx context.Context, event WebhookEvent) error {
	slog.Info("handling GitHub installation event",
		"action", event.InstallationAction,
		"installation_id", event.InstallationID,
		"account_login", event.SenderLogin,
		"repositories_added", len(event.RepositoriesAdded),
		"repositories_removed", len(event.RepositoriesRemoved))

	// Parse the raw payload into proper Installation Event structure
	installationEvent, err := g.parseInstallationEvent(event.RawPayload)
	if err != nil {
		return fmt.Errorf("failed to parse installation event: %w", err)
	}

	// Handle different installation actions
	switch installationEvent.Action {
	case "created":
		return g.handleInstallationCreated(ctx, installationEvent)
	case "deleted":
		return g.handleInstallationDeleted(ctx, installationEvent)
	case "suspend":
		return g.handleInstallationSuspended(ctx, installationEvent)
	case "unsuspend":
		return g.handleInstallationUnsuspended(ctx, installationEvent)
	case "new_permissions_accepted":
		return g.handlePermissionsUpdated(ctx, installationEvent)
	default:
		slog.Debug("unhandled installation action", "action", installationEvent.Action)
		return nil
	}
}

func (g *githubConnector) parseInstallationEvent(rawPayload map[string]any) (*InstallationEvent, error) {
	payloadBytes, err := json.Marshal(rawPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw payload: %w", err)
	}

	var installationEvent InstallationEvent
	if err := json.Unmarshal(payloadBytes, &installationEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal installation event: %w", err)
	}

	installationEvent.RawPayload = rawPayload
	return &installationEvent, nil
}

func (g *githubConnector) handleInstallationCreated(ctx context.Context, event *InstallationEvent) error {
	slog.Info("GitHub App installation created",
		"installation_id", event.Installation.ID,
		"account", event.Installation.Account.Login,
		"repository_selection", event.Installation.RepositorySelection,
		"repository_count", len(event.Repositories))

	// Store in unclaimed_installations table for later processing
	unclaimedInstallation := UnclaimedInstallation{
		GitHubInstallationID: event.Installation.ID,
		GitHubAppID:          event.Installation.AppID,
		GitHubAccountID:      event.Installation.Account.ID,
		GitHubAccountLogin:   event.Installation.Account.Login,
		GitHubAccountType:    event.Installation.Account.Type,
		RepositorySelection:  event.Installation.RepositorySelection,
		Permissions:          event.Installation.Permissions,
		Events:               event.Installation.Events,
		AccessTokensURL:      event.Installation.AccessTokensURL,
		RepositoriesURL:      event.Installation.RepositoriesURL,
		HTMLURL:              event.Installation.HTMLURL,
		AppSlug:              event.Installation.AppSlug,
		SuspendedAt:          event.Installation.SuspendedAt,
		SuspendedBy:          event.Installation.SuspendedBy,
		WebhookSender:        &event.Sender,
		RawWebhookPayload:    event.RawPayload,
		GitHubCreatedAt:      event.Installation.CreatedAt,
		GitHubUpdatedAt:      event.Installation.UpdatedAt,
	}

	// Store unclaimed installation using repository service
	if err := g.repositoryService.StoreUnclaimedInstallation(ctx, &unclaimedInstallation); err != nil {
		return fmt.Errorf("failed to store unclaimed installation: %w", err)
	}

	return nil
}

func (g *githubConnector) handleInstallationDeleted(ctx context.Context, event *InstallationEvent) error {
	slog.Info("GitHub App installation deleted",
		"installation_id", event.Installation.ID,
		"account", event.Installation.Account.Login)

	// TODO: Mark installation as deleted in database
	// TODO: Revoke any active integrations for this installation
	// TODO: Clean up repository permissions tracking

	return nil
}

func (g *githubConnector) handleInstallationSuspended(ctx context.Context, event *InstallationEvent) error {
	slog.Info("GitHub App installation suspended",
		"installation_id", event.Installation.ID,
		"account", event.Installation.Account.Login,
		"suspended_by", event.Installation.SuspendedBy)

	// TODO: Update installation status in database
	// TODO: Disable webhook processing for this installation

	return nil
}

func (g *githubConnector) handleInstallationUnsuspended(ctx context.Context, event *InstallationEvent) error {
	slog.Info("GitHub App installation unsuspended",
		"installation_id", event.Installation.ID,
		"account", event.Installation.Account.Login)

	// TODO: Update installation status in database
	// TODO: Re-enable webhook processing for this installation

	return nil
}

func (g *githubConnector) handlePermissionsUpdated(ctx context.Context, event *InstallationEvent) error {
	slog.Info("GitHub App permissions updated",
		"installation_id", event.Installation.ID,
		"account", event.Installation.Account.Login,
		"permissions", event.Installation.Permissions)

	// TODO: Update permissions in database
	// TODO: Sync repository access based on new permissions

	return nil
}

func (g *githubConnector) handleInstallationRepositoriesEvent(ctx context.Context, event WebhookEvent) error {
	slog.Info("handling GitHub installation repositories event",
		"action", event.Action,
		"installation_id", event.InstallationID,
		"repositories_added", len(event.RepositoriesAdded),
		"repositories_removed", len(event.RepositoriesRemoved))

	// Parse the raw payload into proper Installation Event structure
	installationEvent, err := g.parseInstallationEvent(event.RawPayload)
	if err != nil {
		return fmt.Errorf("failed to parse installation repositories event: %w", err)
	}

	// Handle repository additions and removals
	switch installationEvent.Action {
	case "added":
		return g.handleRepositoriesAdded(ctx, installationEvent)
	case "removed":
		return g.handleRepositoriesRemoved(ctx, installationEvent)
	default:
		slog.Debug("unhandled installation repositories action", "action", installationEvent.Action)
		return nil
	}
}

func (g *githubConnector) handleRepositoriesAdded(ctx context.Context, event *InstallationEvent) error {
	slog.Info("GitHub App repositories added",
		"installation_id", event.Installation.ID,
		"account", event.Installation.Account.Login,
		"repositories_count", len(event.RepositoriesAdded))

	for _, repo := range event.RepositoriesAdded {
		slog.Debug("repository added",
			"installation_id", event.Installation.ID,
			"repository_id", repo.ID,
			"repository_name", repo.FullName)
	}

	// Update repository permissions tracking using repository service
	// TODO: Get integration ID from installation ID
	// integrationID := getIntegrationIDFromInstallationID(event.Installation.ID)
	// if integrationID != "" {
	//     if err := g.repositoryService.AddRepositories(ctx, integrationID, event.RepositoriesAdded); err != nil {
	//         return fmt.Errorf("failed to add repositories: %w", err)
	//     }
	// }

	return nil
}

func (g *githubConnector) handleRepositoriesRemoved(ctx context.Context, event *InstallationEvent) error {
	slog.Info("GitHub App repositories removed",
		"installation_id", event.Installation.ID,
		"account", event.Installation.Account.Login,
		"repositories_count", len(event.RepositoriesRemoved))

	for _, repo := range event.RepositoriesRemoved {
		slog.Debug("repository removed",
			"installation_id", event.Installation.ID,
			"repository_id", repo.ID,
			"repository_name", repo.FullName)
	}

	// Remove repository permissions using repository service
	// TODO: Get integration ID from installation ID
	// integrationID := getIntegrationIDFromInstallationID(event.Installation.ID)
	// if integrationID != "" {
	//     var repoIDs []int64
	//     for _, repo := range event.RepositoriesRemoved {
	//         repoIDs = append(repoIDs, repo.ID)
	//     }
	//     if err := g.repositoryService.RemoveRepositories(ctx, integrationID, repoIDs); err != nil {
	//         return fmt.Errorf("failed to remove repositories: %w", err)
	//     }
	// }

	return nil
}

// UnclaimedInstallation represents an installation waiting to be claimed
type UnclaimedInstallation struct {
	GitHubInstallationID int64
	GitHubAppID          int64
	GitHubAccountID      int64
	GitHubAccountLogin   string
	GitHubAccountType    string
	RepositorySelection  string
	Permissions          map[string]string
	Events               []string
	AccessTokensURL      string
	RepositoriesURL      string
	HTMLURL              string
	AppSlug              string
	SuspendedAt          *time.Time
	SuspendedBy          *User
	WebhookSender        *User
	RawWebhookPayload    map[string]any
	GitHubCreatedAt      time.Time
	GitHubUpdatedAt      time.Time
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