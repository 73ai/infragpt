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

// ValidateWebhookSignature validates the GitHub webhook signature
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

// ProcessEvent processes a GitHub webhook event
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

// Subscribe starts the webhook server and processes incoming webhook events
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


// computeSignature computes the HMAC signature for webhook validation
func (g *githubConnector) computeSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return fmt.Sprintf("sha256=%s", hex.EncodeToString(h.Sum(nil)))
}

// handleInstallationEvent processes GitHub installation webhook events
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

// Webhook server configuration and implementation
type webhookServerConfig struct {
	port                int
	webhookSecret       string
	callbackHandlerFunc func(ctx context.Context, event any) error
}

// startWebhookServer starts the HTTP server for handling GitHub webhooks
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

// webhookHandler handles incoming GitHub webhook requests
type webhookHandler struct {
	http.ServeMux
	callbackHandlerFunc func(ctx context.Context, event any) error
}

// init initializes the webhook handler with HTTP routes
func (wh *webhookHandler) init() {
	wh.HandleFunc("/webhooks/github", wh.handler())
}

// handler returns the HTTP handler function for GitHub webhooks
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

// convertToWebhookEvent converts raw GitHub webhook payload to WebhookEvent
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

// panicMiddleware recovers from panics and returns a 500 error
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

// webhookValidationMiddleware validates GitHub webhook signatures
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

// validateGitHubSignature validates the GitHub webhook signature
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