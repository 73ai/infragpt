package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/connectors/github"
)

func main() {
	fmt.Println("🔍 Testing GitHub App Authorization Flow Components...")
	
	// Test 1: State Creation and Parsing
	fmt.Println("\n1. Testing State Creation and Parsing...")
	testStateHandling()
	
	// Test 2: Authorization Intent Generation
	fmt.Println("\n2. Testing Authorization Intent Generation...")
	testAuthorizationIntent()
	
	// Test 3: Webhook Configuration
	fmt.Println("\n3. Testing Webhook Configuration...")
	testWebhookConfig()
	
	// Test 4: Credential Data Structures
	fmt.Println("\n4. Testing Credential Data Structures...")
	testCredentialStructures()
	
	fmt.Println("\n✅ All GitHub integration flow tests completed successfully!")
}

func testStateHandling() {
	// Create test data
	organizationID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	
	// Create state manually (simulating InitiateAuthorization)
	stateData := map[string]interface{}{
		"organization_id": organizationID,
		"user_id":         userID,
		"timestamp":       time.Now().Unix(),
	}
	
	stateJSON, err := json.Marshal(stateData)
	if err != nil {
		log.Fatalf("❌ Failed to marshal state data: %v", err)
	}
	
	state := base64.URLEncoding.EncodeToString(stateJSON)
	fmt.Printf("   ✓ State created: %s...\n", state[:20])
	
	// Create connector and test parsing
	config := github.Config{
		AppID:       "123456",
		PrivateKey:  "",  // Empty for this test
		RedirectURL: "https://app-local.infragpt.io/integrations/github/callback",
	}
	
	connector := config.NewConnector()
	parsedOrgID, parsedUserID, err := connector.ParseState(state)
	if err != nil {
		log.Fatalf("❌ ParseState failed: %v", err)
	}
	
	if parsedOrgID != organizationID {
		log.Fatalf("❌ Organization ID mismatch: expected %s, got %s", organizationID, parsedOrgID)
	}
	
	if parsedUserID != userID {
		log.Fatalf("❌ User ID mismatch: expected %s, got %s", userID, parsedUserID)
	}
	
	fmt.Printf("   ✓ State parsed correctly - Org: %s, User: %s\n", parsedOrgID[:8], parsedUserID[:8])
}

func testAuthorizationIntent() {
	config := github.Config{
		AppID:       "123456",
		PrivateKey:  "",  // Empty for this test
		RedirectURL: "https://app-local.infragpt.io/integrations/github/callback",
	}
	
	connector := config.NewConnector()
	
	organizationID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	
	intent, err := connector.InitiateAuthorization(organizationID, userID)
	if err != nil {
		log.Fatalf("❌ InitiateAuthorization failed: %v", err)
	}
	
	// Validate intent structure
	if intent.Type != infragpt.AuthorizationTypeInstallation {
		log.Fatalf("❌ Expected authorization type installation, got %s", intent.Type)
	}
	
	if intent.URL == "" {
		log.Fatalf("❌ Expected URL to be non-empty")
	}
	
	expectedURLPrefix := fmt.Sprintf("https://github.com/apps/%s/installations/new", config.AppID)
	if len(intent.URL) < len(expectedURLPrefix) || intent.URL[:len(expectedURLPrefix)] != expectedURLPrefix {
		log.Fatalf("❌ URL format incorrect: expected to start with %s, got %s", expectedURLPrefix, intent.URL)
	}
	
	fmt.Printf("   ✓ Authorization intent generated correctly\n")
	fmt.Printf("   ✓ URL: %s\n", intent.URL)
}

func testWebhookConfig() {
	config := github.Config{
		AppID:         "123456",
		PrivateKey:    "",
		WebhookSecret: "test-webhook-secret",
		RedirectURL:   "https://app-local.infragpt.io/integrations/github/callback",
	}
	
	connector := config.NewConnector()
	
	// Test webhook signature validation (simplified)
	payload := []byte(`{"action":"created","installation":{"id":12345}}`)
	
	// Test with empty signature (should fail gracefully)
	err := connector.ValidateWebhookSignature(payload, "", "test-webhook-secret")
	if err == nil {
		log.Fatalf("❌ Expected empty signature to fail validation")
	}
	
	fmt.Printf("   ✓ Webhook signature validation works correctly\n")
	
	// Test webhook URL building
	testCredentials := infragpt.Credentials{
		Type: infragpt.CredentialTypeToken,
		Data: map[string]string{
			"installation_id": "123456",
			"access_token":    "test-token",
		},
	}
	
	err = connector.ConfigureWebhooks("test-integration-id", testCredentials)
	if err != nil {
		log.Fatalf("❌ ConfigureWebhooks failed: %v", err)
	}
	
	fmt.Printf("   ✓ Webhook configuration works correctly\n")
}

func testCredentialStructures() {
	// Test authorization data structure
	authData := infragpt.AuthorizationData{
		Code:           "test-code",
		State:          "test-state", 
		InstallationID: "123456",
	}
	
	if authData.InstallationID == "" {
		log.Fatalf("❌ AuthorizationData structure not working correctly")
	}
	
	fmt.Printf("   ✓ AuthorizationData structure valid\n")
	
	// Test credentials structure
	creds := infragpt.Credentials{
		Type: infragpt.CredentialTypeToken,
		Data: map[string]string{
			"installation_id": "123456",
			"access_token":    "test-token",
			"account_login":   "test-org",
		},
		ExpiresAt: nil,
	}
	
	if creds.Type != infragpt.CredentialTypeToken {
		log.Fatalf("❌ Credentials structure not working correctly")
	}
	
	fmt.Printf("   ✓ Credentials structure valid\n")
	
	// Test integration structure
	integration := infragpt.Integration{
		ID:                      "test-id",
		OrganizationID:          "org-id",
		UserID:                  "user-id",
		ConnectorType:           infragpt.ConnectorTypeGithub,
		Status:                  infragpt.IntegrationStatusActive,
		BotID:                   "123456",
		ConnectorOrganizationID: "github-org-id",
		Metadata:                make(map[string]string),
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}
	
	if integration.ConnectorType != infragpt.ConnectorTypeGithub {
		log.Fatalf("❌ Integration structure not working correctly")
	}
	
	fmt.Printf("   ✓ Integration structure valid\n")
}