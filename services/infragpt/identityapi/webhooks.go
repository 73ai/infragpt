package identityapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/priyanshujain/infragpt/services/infragpt"
)

type ClerkWebhookEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type ClerkUser struct {
	ID         string `json:"id"`
	EmailAddresses []struct {
		EmailAddress string `json:"email_address"`
	} `json:"email_addresses"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ClerkOrganization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	CreatedBy   string `json:"created_by"`
}

type ClerkOrganizationMembership struct {
	Organization struct {
		ID string `json:"id"`
	} `json:"organization"`
	PublicUserData struct {
		UserID string `json:"user_id"`
	} `json:"public_user_data"`
	Role string `json:"role"`
}

func (h *Handler) HandleClerkWebhook(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	type response struct{}

	ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
		switch req.Type {
		case "user.created":
			return response{}, h.handleUserCreated(ctx, req.Data)
		case "user.updated":
			return response{}, h.handleUserUpdated(ctx, req.Data)
		case "organization.created":
			return response{}, h.handleOrganizationCreated(ctx, req.Data)
		case "organization.updated":
			return response{}, h.handleOrganizationUpdated(ctx, req.Data)
		case "organizationMembership.created":
			return response{}, h.handleOrganizationMemberAdded(ctx, req.Data)
		case "organizationMembership.deleted":
			return response{}, h.handleOrganizationMemberRemoved(ctx, req.Data)
		default:
			return response{}, fmt.Errorf("unsupported webhook event type: %s", req.Type)
		}
	})(w, r)
}

func (h *Handler) handleUserCreated(ctx context.Context, data json.RawMessage) error {
	var user ClerkUser
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

	return h.identityService.SubscribeUserCreated(ctx, event)
}

func (h *Handler) handleUserUpdated(ctx context.Context, data json.RawMessage) error {
	var user ClerkUser
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

	return h.identityService.SubscribeUserUpdated(ctx, event)
}

func (h *Handler) handleOrganizationCreated(ctx context.Context, data json.RawMessage) error {
	var org ClerkOrganization
	if err := json.Unmarshal(data, &org); err != nil {
		return fmt.Errorf("failed to unmarshal organization data: %w", err)
	}

	event := infragpt.OrganizationCreatedEvent{
		ClerkOrgID:      org.ID,
		Name:            org.Name,
		Slug:            org.Slug,
		CreatedByUserID: org.CreatedBy,
	}

	return h.identityService.SubscribeOrganizationCreated(ctx, event)
}

func (h *Handler) handleOrganizationUpdated(ctx context.Context, data json.RawMessage) error {
	var org ClerkOrganization
	if err := json.Unmarshal(data, &org); err != nil {
		return fmt.Errorf("failed to unmarshal organization data: %w", err)
	}

	event := infragpt.OrganizationUpdatedEvent{
		ClerkOrgID: org.ID,
		Name:       org.Name,
		Slug:       org.Slug,
	}

	return h.identityService.SubscribeOrganizationUpdated(ctx, event)
}

func (h *Handler) handleOrganizationMemberAdded(ctx context.Context, data json.RawMessage) error {
	var membership ClerkOrganizationMembership
	if err := json.Unmarshal(data, &membership); err != nil {
		return fmt.Errorf("failed to unmarshal membership data: %w", err)
	}

	event := infragpt.OrganizationMemberAddedEvent{
		ClerkUserID: membership.PublicUserData.UserID,
		ClerkOrgID:  membership.Organization.ID,
		Role:        membership.Role,
	}

	return h.identityService.SubscribeOrganizationMemberAdded(ctx, event)
}

func (h *Handler) handleOrganizationMemberRemoved(ctx context.Context, data json.RawMessage) error {
	var membership ClerkOrganizationMembership
	if err := json.Unmarshal(data, &membership); err != nil {
		return fmt.Errorf("failed to unmarshal membership data: %w", err)
	}

	event := infragpt.OrganizationMemberRemovedEvent{
		ClerkUserID: membership.PublicUserData.UserID,
		ClerkOrgID:  membership.Organization.ID,
	}

	return h.identityService.SubscribeOrganizationMemberRemoved(ctx, event)
}