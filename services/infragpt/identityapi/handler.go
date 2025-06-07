package identityapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
)

type Handler struct {
	identityService infragpt.IdentityService
}

func NewHandler(identityService infragpt.IdentityService) *Handler {
	return &Handler{
		identityService: identityService,
	}
}

func (h *Handler) GetOrganization(w http.ResponseWriter, r *http.Request) {
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

	ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
		query := infragpt.OrganizationQuery{
			ClerkOrgID: req.ClerkOrgID,
		}

		org, err := h.identityService.Organization(ctx, query)
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
	})(w, r)
}

func (h *Handler) SetOrganizationMetadata(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OrganizationID     string   `json:"organization_id"`
		CompanySize        string   `json:"company_size"`
		TeamSize           string   `json:"team_size"`
		UseCases           []string `json:"use_cases"`
		ObservabilityStack []string `json:"observability_stack"`
	}
	type response struct{}

	ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
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

		err = h.identityService.SetOrganizationMetadata(ctx, cmd)
		return response{}, err
	})(w, r)
}

func ApiHandlerFunc[T any, R any](handler func(context.Context, T) (R, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req T
		if r.Method == http.MethodPost && r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
		}

		response, err := handler(ctx, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}