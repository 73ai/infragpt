package identityapi

import (
	"context"
	"encoding/json"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/generic/httperrors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
)

type httpHandler struct {
	http.ServeMux
	svc infragpt.IdentityService
}

func (h *httpHandler) init() {
	h.HandleFunc("/identity/organization", h.organization())
	h.HandleFunc("/identity/organization/set-metadata", h.setOrganizationMetadata())
}

func NewHandler(identityService infragpt.IdentityService) http.Handler {
	return &httpHandler{
		svc: identityService,
	}
}

func (h *httpHandler) organization() func(w http.ResponseWriter, r *http.Request) {
	type request struct {
		ClerkOrgID string `json:"clerk_org_id"`
	}
	type response struct {
		ID         string `json:"id"`
		ClerkOrgID string `json:"clerk_org_id"`
		Name       string `json:"name"`
		Slug       string `json:"slug"`
		CreatedAt  string `json:"created_at"`
		Metadata   struct {
			CompanySize        string   `json:"company_size"`
			TeamSize           string   `json:"team_size"`
			UseCases           []string `json:"use_cases"`
			ObservabilityStack []string `json:"observability_stack"`
			CompletedAt        string   `json:"completed_at"`
		} `json:"metadata"`
	}

	return ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
		query := infragpt.OrganizationQuery{
			ClerkOrgID: req.ClerkOrgID,
		}

		org, err := h.svc.Organization(ctx, query)
		if err != nil {
			return response{}, err
		}

		useCases := make([]string, len(org.Metadata.UseCases))
		for i, uc := range org.Metadata.UseCases {
			useCases[i] = string(uc)
		}

		stack := make([]string, len(org.Metadata.ObservabilityStack))
		for i, s := range org.Metadata.ObservabilityStack {
			stack[i] = string(s)
		}

		resp := response{
			ID:         org.ID.String(),
			ClerkOrgID: org.ClerkOrgID,
			Name:       org.Name,
			Slug:       org.Slug,
			CreatedAt:  org.CreatedAt.Format(time.RFC3339),
			Metadata: struct {
				CompanySize        string   `json:"company_size"`
				TeamSize           string   `json:"team_size"`
				UseCases           []string `json:"use_cases"`
				ObservabilityStack []string `json:"observability_stack"`
				CompletedAt        string   `json:"completed_at"`
			}{
				CompanySize:        string(org.Metadata.CompanySize),
				TeamSize:           string(org.Metadata.TeamSize),
				UseCases:           useCases,
				ObservabilityStack: stack,
				CompletedAt:        org.Metadata.CompletedAt.Format(time.RFC3339),
			},
		}

		return resp, nil
	})
}

func (h *httpHandler) setOrganizationMetadata() func(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OrganizationID     string   `json:"organization_id"`
		CompanySize        string   `json:"company_size"`
		TeamSize           string   `json:"team_size"`
		UseCases           []string `json:"use_cases"`
		ObservabilityStack []string `json:"observability_stack"`
	}
	type response struct{}

	return ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
		orgID, err := uuid.Parse(req.OrganizationID)
		if err != nil {
			return response{}, err
		}

		useCases := make([]infragpt.UseCase, len(req.UseCases))
		for i, uc := range req.UseCases {
			useCases[i] = infragpt.UseCase(uc)
		}

		stack := make([]infragpt.ObservabilityStack, len(req.ObservabilityStack))
		for i, s := range req.ObservabilityStack {
			stack[i] = infragpt.ObservabilityStack(s)
		}

		cmd := infragpt.OrganizationMetadataCommand{
			OrganizationID:     orgID,
			CompanySize:        infragpt.CompanySize(req.CompanySize),
			TeamSize:           infragpt.TeamSize(req.TeamSize),
			UseCases:           useCases,
			ObservabilityStack: stack,
		}

		err = h.svc.SetOrganizationMetadata(ctx, cmd)
		return response{}, err
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
			slog.Error("error in identity api handler", "path", r.URL, "request", request, "err", err)
			var httpError = httperrors.From(err)
			w.WriteHeader(httpError.HttpStatus)
			_ = json.NewEncoder(w).Encode(httpError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
