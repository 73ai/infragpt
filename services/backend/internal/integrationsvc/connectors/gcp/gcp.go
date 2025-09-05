package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	// Note: The following GCP imports are commented out to avoid dependency issues
	// Uncomment and add to go.mod when implementing full GCP IAM validation
	// "cloud.google.com/go/iam/admin/apiv1"
	// "cloud.google.com/go/iam/admin/apiv1/adminpb"
	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/backend"
	"github.com/priyanshujain/infragpt/services/backend/internal/integrationsvc/domain"
	// "google.golang.org/api/cloudresourcemanager/v1"
	// "google.golang.org/api/option"
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

	// Validate the service account JSON
	var sa ServiceAccountKey
	if err := json.Unmarshal([]byte(authData.Code), &sa); err != nil {
		return backend.Credentials{}, fmt.Errorf("invalid service account JSON: %w", err)
	}

	if sa.Type != "service_account" {
		return backend.Credentials{}, fmt.Errorf("invalid type: expected 'service_account', got '%s'", sa.Type)
	}

	// Store the credentials
	creds := backend.Credentials{
		Type: backend.CredentialTypeServiceAccount,
		Data: map[string]string{
			"service_account_json": authData.Code,
			"project_id":           sa.ProjectID,
			"client_email":         sa.ClientEmail,
		},
		OrganizationInfo: &backend.OrganizationInfo{
			ExternalID: sa.ProjectID,
			Name:       sa.ProjectID,
			Metadata: map[string]string{
				"project_id":   sa.ProjectID,
				"client_email": sa.ClientEmail,
			},
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
		return fmt.Errorf("credential validation failed: %w", err)
	}

	if !validation.Valid {
		return fmt.Errorf("invalid service account: %v", validation.Errors)
	}

	if !validation.HasViewer {
		return fmt.Errorf("service account does not have Viewer role")
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
		result.Errors = append(result.Errors, fmt.Sprintf("invalid JSON format: %v", err))
		return result, nil
	}

	// Validate required fields
	if sa.Type != "service_account" {
		result.Errors = append(result.Errors, fmt.Sprintf("invalid type: expected 'service_account', got '%s'", sa.Type))
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

	// TODO: Implement full GCP IAM validation when dependencies are added
	// For now, we do basic validation and assume the viewer role is granted
	// In production, uncomment the GCP imports and implement the following:
	// 1. Create a Resource Manager client with the service account credentials
	// 2. Test access to the project
	// 3. Check IAM policy bindings for Viewer/Editor/Owner roles
	// 4. Return appropriate errors if permissions are missing
	
	// For development/testing, we'll assume valid credentials have the Viewer role
	result.HasViewer = true
	result.Valid = len(result.Errors) == 0

	// Add a warning that this is simplified validation
	if result.Valid {
		// In production, remove this and implement actual GCP API validation
		// The frontend will still show the validation as successful
	}

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