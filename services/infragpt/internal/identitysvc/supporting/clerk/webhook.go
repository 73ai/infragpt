package clerk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/generic/httperrors"
	"log/slog"
	"net"
	"net/http"

	"github.com/priyanshujain/infragpt/services/infragpt"
)

type user struct {
	ID             string `json:"id"`
	EmailAddresses []struct {
		EmailAddress string `json:"email_address"`
	} `json:"email_addresses"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type organization struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedBy string `json:"created_by"`
}

type organizationMembership struct {
	Organization struct {
		ID string `json:"id"`
	} `json:"organization"`
	PublicUserData struct {
		UserID string `json:"user_id"`
	} `json:"public_user_data"`
	Role string `json:"role"`
}

type webhookHandler struct {
	http.ServeMux
	callbackHandlerFunc func(ctx context.Context, event any) error
}

type webhookServerConfig struct {
	port                int
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
		Handler:     panicHandler(h),
	}

	return httpServer.ListenAndServe()
}

func (h *webhookHandler) init() {
	h.HandleFunc("/webhooks/clerk", h.handler())
}

func (h *webhookHandler) handler() func(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	type response struct{}

	return ApiHandlerFunc(func(ctx context.Context, r request) (response, error) {
		switch r.Type {
		case "user.created":
			return response{}, h.handleUserCreated(ctx, r.Data)
		case "user.updated":
			return response{}, h.handleUserUpdated(ctx, r.Data)
		case "organization.created":
			return response{}, h.handleOrganizationCreated(ctx, r.Data)
		case "organization.updated":
			return response{}, h.handleOrganizationUpdated(ctx, r.Data)
		case "organizationMembership.created":
			return response{}, h.handleOrganizationMemberAdded(ctx, r.Data)
		case "organizationMembership.deleted":
			return response{}, h.handleOrganizationMemberRemoved(ctx, r.Data)
		default:
			return response{}, fmt.Errorf("unsupported webhook event type: %s", r.Type)
		}
	})
}

func (h *webhookHandler) handleUserCreated(ctx context.Context, data json.RawMessage) error {
	var user user
	if err := json.Unmarshal(data, &user); err != nil {
		return fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	email := ""
	if len(user.EmailAddresses) > 0 {
		email = user.EmailAddresses[0].EmailAddress
	}

	event := infragpt.UserCreatedEvent{
		ClerkUserID: user.ID,
		Email:       email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
	}

	return h.callbackHandlerFunc(ctx, event)
}

func (h *webhookHandler) handleUserUpdated(ctx context.Context, data json.RawMessage) error {
	var user user
	if err := json.Unmarshal(data, &user); err != nil {
		return fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	email := ""
	if len(user.EmailAddresses) > 0 {
		email = user.EmailAddresses[0].EmailAddress
	}

	event := infragpt.UserUpdatedEvent{
		ClerkUserID: user.ID,
		Email:       email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
	}

	return h.callbackHandlerFunc(ctx, event)
}

func (h *webhookHandler) handleOrganizationCreated(ctx context.Context, data json.RawMessage) error {
	var org organization
	if err := json.Unmarshal(data, &org); err != nil {
		return fmt.Errorf("failed to unmarshal organization data: %w", err)
	}

	event := infragpt.OrganizationCreatedEvent{
		ClerkOrgID:      org.ID,
		Name:            org.Name,
		Slug:            org.Slug,
		CreatedByUserID: org.CreatedBy,
	}

	return h.callbackHandlerFunc(ctx, event)
}

func (h *webhookHandler) handleOrganizationUpdated(ctx context.Context, data json.RawMessage) error {
	var org organization
	if err := json.Unmarshal(data, &org); err != nil {
		return fmt.Errorf("failed to unmarshal organization data: %w", err)
	}

	event := infragpt.OrganizationUpdatedEvent{
		ClerkOrgID: org.ID,
		Name:       org.Name,
		Slug:       org.Slug,
	}

	return h.callbackHandlerFunc(ctx, event)
}

func (h *webhookHandler) handleOrganizationMemberAdded(ctx context.Context, data json.RawMessage) error {
	var membership organizationMembership
	if err := json.Unmarshal(data, &membership); err != nil {
		return fmt.Errorf("failed to unmarshal membership data: %w", err)
	}

	event := infragpt.OrganizationMemberAddedEvent{
		ClerkUserID: membership.PublicUserData.UserID,
		ClerkOrgID:  membership.Organization.ID,
		Role:        membership.Role,
	}

	return h.callbackHandlerFunc(ctx, event)
}

func (h *webhookHandler) handleOrganizationMemberRemoved(ctx context.Context, data json.RawMessage) error {
	var membership organizationMembership
	if err := json.Unmarshal(data, &membership); err != nil {
		return fmt.Errorf("failed to unmarshal membership data: %w", err)
	}

	event := infragpt.OrganizationMemberDeletedEvent{
		ClerkUserID: membership.PublicUserData.UserID,
		ClerkOrgID:  membership.Organization.ID,
	}

	return h.callbackHandlerFunc(ctx, event)
}

func ApiHandlerFunc[T any, R any](handler func(context.Context, T) (R, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var request T
		if r.Method == http.MethodPost && r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
		}

		response, err := handler(ctx, request)
		if err != nil {
			slog.Error("error in clerk webhook api handler", "path", r.URL, "request", request, "err", err)
			var httpError = httperrors.From(err)
			w.WriteHeader(httpError.HttpStatus)
			_ = json.NewEncoder(w).Encode(httpError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

func panicHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("clerk: panic while handling http request", "recover", r)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
