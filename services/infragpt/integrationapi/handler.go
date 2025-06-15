package integrationapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/generic/httperrors"
)

type httpHandler struct {
	http.ServeMux
	svc infragpt.IntegrationService
}

func (h *httpHandler) init() {
	h.HandleFunc("/integrations/authorize/", h.authorize())
	h.HandleFunc("/integrations/callback/", h.callback())
	h.HandleFunc("/integrations/list/", h.list())
	h.HandleFunc("/integrations/revoke/", h.revoke())
	h.HandleFunc("/integrations/refresh/", h.refresh())
	h.HandleFunc("/integrations/status/", h.status())
}

func NewHandler(integrationService infragpt.IntegrationService,
	authMiddleware func(handler http.Handler) http.Handler) http.Handler {
	h := &httpHandler{
		svc: integrationService,
	}

	h.init()
	return authMiddleware(h)
}

func (h *httpHandler) authorize() func(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OrganizationID string `json:"organization_id"`
		UserID         string `json:"user_id"`
		ConnectorType  string `json:"connector_type"`
	}
	type response struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	}

	return ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
		cmd := infragpt.NewIntegrationCommand{
			OrganizationID: req.OrganizationID,
			UserID:         req.UserID,
			ConnectorType:  infragpt.ConnectorType(req.ConnectorType),
		}

		intent, err := h.svc.NewIntegration(ctx, cmd)
		if err != nil {
			return response{}, err
		}

		return response{
			Type: string(intent.Type),
			URL:  intent.URL,
		}, nil
	})
}

func (h *httpHandler) callback() func(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OrganizationID string `json:"organization_id"`
		ConnectorType  string `json:"connector_type"`
		Code           string `json:"code"`
		State          string `json:"state"`
		InstallationID string `json:"installation_id"`
	}
	type response struct {
		ID                      string            `json:"id"`
		OrganizationID          string            `json:"organization_id"`
		UserID                  string            `json:"user_id"`
		ConnectorType           string            `json:"connector_type"`
		Status                  string            `json:"status"`
		BotID                   string            `json:"bot_id,omitempty"`
		ConnectorUserID         string            `json:"connector_user_id,omitempty"`
		ConnectorOrganizationID string            `json:"connector_organization_id,omitempty"`
		Metadata                map[string]string `json:"metadata"`
		CreatedAt               string            `json:"created_at"`
		UpdatedAt               string            `json:"updated_at"`
		LastUsedAt              string            `json:"last_used_at,omitempty"`
	}

	return ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
		cmd := infragpt.AuthorizeIntegrationCommand{
			OrganizationID: req.OrganizationID,
			ConnectorType:  infragpt.ConnectorType(req.ConnectorType),
			Code:           req.Code,
			State:          req.State,
			InstallationID: req.InstallationID,
		}

		integration, err := h.svc.AuthorizeIntegration(ctx, cmd)
		if err != nil {
			return response{}, err
		}

		resp := response{
			ID:                      integration.ID,
			OrganizationID:          integration.OrganizationID,
			UserID:                  integration.UserID,
			ConnectorType:           string(integration.ConnectorType),
			Status:                  string(integration.Status),
			BotID:                   integration.BotID,
			ConnectorUserID:         integration.ConnectorUserID,
			ConnectorOrganizationID: integration.ConnectorOrganizationID,
			Metadata:                integration.Metadata,
			CreatedAt:               integration.CreatedAt.Format(time.RFC3339),
			UpdatedAt:               integration.UpdatedAt.Format(time.RFC3339),
		}

		if integration.LastUsedAt != nil {
			resp.LastUsedAt = integration.LastUsedAt.Format(time.RFC3339)
		}

		return resp, nil
	})
}

func (h *httpHandler) list() func(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OrganizationID string `json:"organization_id"`
		ConnectorType  string `json:"connector_type,omitempty"`
	}
	type integration struct {
		ID                      string            `json:"id"`
		OrganizationID          string            `json:"organization_id"`
		UserID                  string            `json:"user_id"`
		ConnectorType           string            `json:"connector_type"`
		Status                  string            `json:"status"`
		BotID                   string            `json:"bot_id,omitempty"`
		ConnectorUserID         string            `json:"connector_user_id,omitempty"`
		ConnectorOrganizationID string            `json:"connector_organization_id,omitempty"`
		Metadata                map[string]string `json:"metadata"`
		CreatedAt               string            `json:"created_at"`
		UpdatedAt               string            `json:"updated_at"`
		LastUsedAt              string            `json:"last_used_at,omitempty"`
	}
	type response struct {
		Integrations []integration `json:"integrations"`
	}

	return ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
		query := infragpt.IntegrationsQuery{
			OrganizationID: req.OrganizationID,
		}

		if req.ConnectorType != "" {
			query.ConnectorType = infragpt.ConnectorType(req.ConnectorType)
		}

		integrations, err := h.svc.Integrations(ctx, query)
		if err != nil {
			return response{}, err
		}

		resp := response{
			Integrations: make([]integration, len(integrations)),
		}

		for i, integ := range integrations {
			resp.Integrations[i] = integration{
				ID:                      integ.ID,
				OrganizationID:          integ.OrganizationID,
				UserID:                  integ.UserID,
				ConnectorType:           string(integ.ConnectorType),
				Status:                  string(integ.Status),
				BotID:                   integ.BotID,
				ConnectorUserID:         integ.ConnectorUserID,
				ConnectorOrganizationID: integ.ConnectorOrganizationID,
				Metadata:                integ.Metadata,
				CreatedAt:               integ.CreatedAt.Format(time.RFC3339),
				UpdatedAt:               integ.UpdatedAt.Format(time.RFC3339),
			}

			if integ.LastUsedAt != nil {
				resp.Integrations[i].LastUsedAt = integ.LastUsedAt.Format(time.RFC3339)
			}
		}

		return resp, nil
	})
}

func (h *httpHandler) revoke() func(w http.ResponseWriter, r *http.Request) {
	type request struct {
		IntegrationID  string `json:"integration_id"`
		OrganizationID string `json:"organization_id"`
	}
	type response struct{}

	return ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
		cmd := infragpt.RevokeIntegrationCommand{
			IntegrationID:  req.IntegrationID,
			OrganizationID: req.OrganizationID,
		}

		err := h.svc.RevokeIntegration(ctx, cmd)
		return response{}, err
	})
}

func (h *httpHandler) refresh() func(w http.ResponseWriter, r *http.Request) {
	type request struct {
		IntegrationID  string `json:"integration_id"`
		OrganizationID string `json:"organization_id"`
	}
	type response struct {
		Message string `json:"message"`
	}

	return ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
		// TODO: Implement credential refresh logic
		// This would involve:
		// 1. Get integration by ID and validate organization
		// 2. Get credentials for integration
		// 3. Call connector.RefreshCredentials()
		// 4. Update stored credentials

		return response{
			Message: "Credential refresh not implemented yet",
		}, nil
	})
}

func (h *httpHandler) status() func(w http.ResponseWriter, r *http.Request) {
	type request struct {
		IntegrationID  string `json:"integration_id"`
		OrganizationID string `json:"organization_id"`
	}
	type response struct {
		ID                      string            `json:"id"`
		OrganizationID          string            `json:"organization_id"`
		UserID                  string            `json:"user_id"`
		ConnectorType           string            `json:"connector_type"`
		Status                  string            `json:"status"`
		BotID                   string            `json:"bot_id,omitempty"`
		ConnectorUserID         string            `json:"connector_user_id,omitempty"`
		ConnectorOrganizationID string            `json:"connector_organization_id,omitempty"`
		Metadata                map[string]string `json:"metadata"`
		CreatedAt               string            `json:"created_at"`
		UpdatedAt               string            `json:"updated_at"`
		LastUsedAt              string            `json:"last_used_at,omitempty"`
		HealthStatus            string            `json:"health_status"`
	}

	return ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
		query := infragpt.IntegrationQuery{
			IntegrationID:  req.IntegrationID,
			OrganizationID: req.OrganizationID,
		}

		integration, err := h.svc.Integration(ctx, query)
		if err != nil {
			return response{}, err
		}

		// TODO: Implement health check logic
		// This would involve calling connector.ValidateCredentials()
		healthStatus := "unknown"

		resp := response{
			ID:                      integration.ID,
			OrganizationID:          integration.OrganizationID,
			UserID:                  integration.UserID,
			ConnectorType:           string(integration.ConnectorType),
			Status:                  string(integration.Status),
			BotID:                   integration.BotID,
			ConnectorUserID:         integration.ConnectorUserID,
			ConnectorOrganizationID: integration.ConnectorOrganizationID,
			Metadata:                integration.Metadata,
			CreatedAt:               integration.CreatedAt.Format(time.RFC3339),
			UpdatedAt:               integration.UpdatedAt.Format(time.RFC3339),
			HealthStatus:            healthStatus,
		}

		if integration.LastUsedAt != nil {
			resp.LastUsedAt = integration.LastUsedAt.Format(time.RFC3339)
		}

		return resp, nil
	})
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
			slog.Error("error in integration api handler", "path", r.URL, "request", request, "err", err)
			var httpError = httperrors.From(err)
			w.WriteHeader(httpError.HttpStatus)
			_ = json.NewEncoder(w).Encode(httpError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
