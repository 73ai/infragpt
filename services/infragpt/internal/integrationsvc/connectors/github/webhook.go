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

type webhookHandler struct {
	http.ServeMux
	callbackHandlerFunc func(ctx context.Context, event any) error
}

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

func (wh *webhookHandler) init() {
	wh.HandleFunc("/webhooks/github", wh.handler())
}

func (wh *webhookHandler) handler() func(w http.ResponseWriter, r *http.Request) {
	type response struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get event type from GitHub headers
		eventType := r.Header.Get("X-GitHub-Event")
		if eventType == "" {
			http.Error(w, "Missing X-GitHub-Event header", http.StatusBadRequest)
			return
		}

		// Read payload
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read payload", http.StatusBadRequest)
			return
		}

		// Parse payload as generic JSON
		var rawPayload map[string]interface{}
		if err := json.Unmarshal(payload, &rawPayload); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Convert to our WebhookEvent format
		webhookEvent, err := wh.convertToWebhookEvent(eventType, rawPayload)
		if err != nil {
			slog.Error("failed to convert GitHub webhook event", "event_type", eventType, "error", err)
			http.Error(w, "Failed to process event", http.StatusInternalServerError)
			return
		}

		// Call the handler
		if err := wh.callbackHandlerFunc(ctx, webhookEvent); err != nil {
			slog.Error("error handling GitHub webhook event", "event_type", eventType, "error", err)
			http.Error(w, "Failed to handle event", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response{})
	}
}

func (wh *webhookHandler) convertToWebhookEvent(eventType string, rawPayload map[string]interface{}) (WebhookEvent, error) {
	event := WebhookEvent{
		EventType:  EventType(eventType),
		RawPayload: rawPayload,
		CreatedAt:  time.Now(),
	}

	// Extract common fields
	if installation, ok := rawPayload["installation"].(map[string]interface{}); ok {
		if id, ok := installation["id"].(float64); ok {
			event.InstallationID = int64(id)
		}
	}

	if sender, ok := rawPayload["sender"].(map[string]interface{}); ok {
		if id, ok := sender["id"].(float64); ok {
			event.SenderID = int64(id)
		}
		if login, ok := sender["login"].(string); ok {
			event.SenderLogin = login
		}
	}

	if repository, ok := rawPayload["repository"].(map[string]interface{}); ok {
		if id, ok := repository["id"].(float64); ok {
			event.RepositoryID = int64(id)
		}
		if name, ok := repository["full_name"].(string); ok {
			event.RepositoryName = name
		}
	}

	if action, ok := rawPayload["action"].(string); ok {
		event.Action = action
	}

	// Event-specific field extraction
	switch EventType(eventType) {
	case EventTypePush:
		if ref, ok := rawPayload["ref"].(string); ok {
			event.Ref = ref
			if strings.HasPrefix(ref, "refs/heads/") {
				event.Branch = strings.TrimPrefix(ref, "refs/heads/")
			}
		}
		if after, ok := rawPayload["after"].(string); ok {
			event.CommitSHA = after
		}

	case EventTypePullRequest:
		if pr, ok := rawPayload["pull_request"].(map[string]interface{}); ok {
			if number, ok := pr["number"].(float64); ok {
				event.PullRequestNumber = int(number)
			}
			if title, ok := pr["title"].(string); ok {
				event.PullRequestTitle = title
			}
			if state, ok := pr["state"].(string); ok {
				event.PullRequestState = state
			}
			if head, ok := pr["head"].(map[string]interface{}); ok {
				if sha, ok := head["sha"].(string); ok {
					event.CommitSHA = sha
				}
			}
		}

	case EventTypeIssues:
		if issue, ok := rawPayload["issue"].(map[string]interface{}); ok {
			if number, ok := issue["number"].(float64); ok {
				event.IssueNumber = int(number)
			}
			if title, ok := issue["title"].(string); ok {
				event.IssueTitle = title
			}
			if state, ok := issue["state"].(string); ok {
				event.IssueState = state
			}
		}

	case EventTypeInstallation:
		if action, ok := rawPayload["action"].(string); ok {
			event.InstallationAction = action
		}
		
		// Handle repository changes
		if repositories, ok := rawPayload["repositories"].([]interface{}); ok {
			for _, repo := range repositories {
				if repoMap, ok := repo.(map[string]interface{}); ok {
					if fullName, ok := repoMap["full_name"].(string); ok {
						if event.InstallationAction == "created" {
							event.RepositoriesAdded = append(event.RepositoriesAdded, fullName)
						}
					}
				}
			}
		}
		
		if repositoriesRemoved, ok := rawPayload["repositories_removed"].([]interface{}); ok {
			for _, repo := range repositoriesRemoved {
				if repoMap, ok := repo.(map[string]interface{}); ok {
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
			// Skip validation if no secret configured
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

		// Restore body for downstream handlers
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		next.ServeHTTP(w, r)
	})
}

func validateGitHubSignature(payload []byte, signature string, secret string) bool {
	// GitHub sends signature as sha256=<hash>
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	
	expectedHash := strings.TrimPrefix(signature, "sha256=")
	
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	actualHash := hex.EncodeToString(h.Sum(nil))
	
	return hmac.Equal([]byte(expectedHash), []byte(actualHash))
}