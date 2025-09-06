package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/backend"
	"github.com/priyanshujain/infragpt/services/backend/internal/integrationsvc/domain"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

// ServiceAccountKey represents the structure of a GCP service account JSON key
type ServiceAccountKey struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain,omitempty"`
}


// ValidationResult represents the result of service account validation
type ValidationResult struct {
	Valid       bool     `json:"valid"`
	ProjectID   string   `json:"project_id"`
	ClientEmail string   `json:"client_email"`
	HasViewer   bool     `json:"has_viewer_role"`
	Errors      []string `json:"errors,omitempty"`
}

// Connector implements the domain.Connector interface for GCP
type Connector struct {
	integrationRepository domain.IntegrationRepository
	credentialRepository  domain.CredentialRepository
}


// InitiateAuthorization initiates the authorization process for GCP
func (c *Connector) InitiateAuthorization(organizationID string, userID string) (backend.IntegrationAuthorizationIntent, error) {
	// For GCP with service account, we use a special flow
	// The frontend will handle the JSON input directly
	return backend.IntegrationAuthorizationIntent{
		Type: backend.AuthorizationTypeAPIKey,
		URL:  "gcp-service-account", // Special marker for frontend
	}, nil
}

// ParseState parses the authorization state
func (c *Connector) ParseState(state string) (organizationID uuid.UUID, userID uuid.UUID, err error) {
	parts := strings.Split(state, ":")
	if len(parts) != 2 {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid state format")
	}

	orgID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid organization ID in state: %w", err)
	}

	uID, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid user ID in state: %w", err)
	}

	return orgID, uID, nil
}

// CompleteAuthorization completes the authorization process with service account JSON
func (c *Connector) CompleteAuthorization(authData backend.AuthorizationData) (backend.Credentials, error) {
	// The service account JSON will be in the Code field
	if authData.Code == "" {
		return backend.Credentials{}, fmt.Errorf("service account JSON is required")
	}

	// Basic validation - just ensure it's valid JSON
	var jsonCheck map[string]interface{}
	if err := json.Unmarshal([]byte(authData.Code), &jsonCheck); err != nil {
		return backend.Credentials{}, fmt.Errorf("invalid JSON format")
	}

	// Store the credentials as a simple string - validation will happen in ValidateCredentials
	// The credential repository will handle encryption
	creds := backend.Credentials{
		Type: backend.CredentialTypeServiceAccount,
		Data: map[string]string{
			"service_account_json": authData.Code, // This will be encrypted by the credential repository
		},
	}

	return creds, nil
}

// ValidateCredentials validates GCP service account credentials and checks for Viewer role
func (c *Connector) ValidateCredentials(creds backend.Credentials) error {
	saJSON, exists := creds.Data["service_account_json"]
	if !exists {
		return fmt.Errorf("service account JSON not found in credentials")
	}

	validation, err := ValidateServiceAccountWithViewer([]byte(saJSON))
	if err != nil {
		return fmt.Errorf("credential validation failed - please check your service account JSON format and permissions")
	}

	if !validation.Valid {
		return fmt.Errorf("invalid service account - please check your credentials")
	}

	if !validation.HasViewer {
		return fmt.Errorf("service account does not have required Viewer permissions - please grant the Viewer role in the GCP Console")
	}

	return nil
}

// ValidateServiceAccountWithViewer validates the service account and checks for Viewer role
func ValidateServiceAccountWithViewer(jsonData []byte) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  false,
		Errors: []string{},
	}

	// Parse the service account JSON
	var sa ServiceAccountKey
	if err := json.Unmarshal(jsonData, &sa); err != nil {
		result.Errors = append(result.Errors, "invalid service account JSON format")
		return result, nil
	}

	// Validate required fields
	if sa.Type != "service_account" {
		result.Errors = append(result.Errors, "invalid service account type")
		return result, nil
	}

	if sa.ProjectID == "" {
		result.Errors = append(result.Errors, "project_id is required")
		return result, nil
	}

	if sa.ClientEmail == "" {
		result.Errors = append(result.Errors, "client_email is required")
		return result, nil
	}

	if sa.PrivateKey == "" {
		result.Errors = append(result.Errors, "private_key is required")
		return result, nil
	}

	result.ProjectID = sa.ProjectID
	result.ClientEmail = sa.ClientEmail

	// Create a context for API calls
	ctx := context.Background()

	// Test authentication by creating clients with service account credentials
	option := option.WithCredentialsJSON(jsonData)

	// Test Resource Manager API access
	service, err := cloudresourcemanager.NewService(ctx, option)
	if err != nil {
		result.Errors = append(result.Errors, "failed to authenticate with service account")
		return result, nil
	}

	// Get project to verify access
	project, err := service.Projects.Get(sa.ProjectID).Context(ctx).Do()
	if err != nil {
		result.Errors = append(result.Errors, "failed to access project - please verify permissions")
		return result, nil
	}

	if project == nil {
		result.Errors = append(result.Errors, "project not found or no access")
		return result, nil
	}

	// Get IAM policy for the project
	policy, err := service.Projects.GetIamPolicy(sa.ProjectID, &cloudresourcemanager.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		result.Errors = append(result.Errors, "failed to retrieve IAM policy")
		return result, nil
	}

	// Check if the service account has Viewer role or higher
	hasViewer := false
	viewerRoles := []string{"roles/viewer", "roles/owner", "roles/editor"}
	memberIdentity := fmt.Sprintf("serviceAccount:%s", sa.ClientEmail)

	for _, binding := range policy.Bindings {
		// Check if this binding grants a viewer role or higher
		for _, role := range viewerRoles {
			if binding.Role == role {
				// Check if our service account is in the members list
				for _, member := range binding.Members {
					if member == memberIdentity {
						hasViewer = true
						break
					}
				}
			}
			if hasViewer {
				break
			}
		}
		if hasViewer {
			break
		}
	}

	if !hasViewer {
		result.Errors = append(result.Errors, "service account does not have Viewer role")
	}

	result.HasViewer = hasViewer
	result.Valid = len(result.Errors) == 0

	return result, nil
}

// RefreshCredentials refreshes the credentials (not applicable for service accounts)
func (c *Connector) RefreshCredentials(creds backend.Credentials) (backend.Credentials, error) {
	// Service account credentials don't need refreshing
	return creds, nil
}

// RevokeCredentials revokes the credentials
func (c *Connector) RevokeCredentials(creds backend.Credentials) error {
	// Service account credentials are revoked by deleting them from storage
	// The actual key revocation should be done in GCP Console
	return nil
}

// ConfigureWebhooks configures webhooks (not applicable for GCP)
func (c *Connector) ConfigureWebhooks(integrationID string, creds backend.Credentials) error {
	// GCP doesn't use webhooks for this integration
	return nil
}

// ValidateWebhookSignature validates webhook signatures (not applicable for GCP)
func (c *Connector) ValidateWebhookSignature(payload []byte, signature string, secret string) error {
	return fmt.Errorf("webhooks not supported for GCP connector")
}

// Subscribe subscribes to events (not applicable for GCP)
func (c *Connector) Subscribe(ctx context.Context, handler func(ctx context.Context, event any) error) error {
	// GCP connector doesn't have event subscription
	<-ctx.Done()
	return ctx.Err()
}

// ProcessEvent processes events (not applicable for GCP)
func (c *Connector) ProcessEvent(ctx context.Context, event any) error {
	return fmt.Errorf("event processing not supported for GCP connector")
}

// Sync performs synchronization operations
func (c *Connector) Sync(ctx context.Context, integration backend.Integration, params map[string]string) error {
	// For GCP, sync could involve refreshing project metadata or permissions
	// For now, we'll just validate the credentials are still valid
	credRecord, err := c.credentialRepository.FindByIntegration(ctx, integration.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	// Convert domain.IntegrationCredential to backend.Credentials
	creds := backend.Credentials{
		Type:      credRecord.CredentialType,
		Data:      credRecord.Data,
		ExpiresAt: credRecord.ExpiresAt,
	}

	return c.ValidateCredentials(creds)
}
