package identitytest

import (
	"context"
	"testing"

	"github.com/priyanshujain/infragpt/services/infragpt"
)

func Ensure(t *testing.T, f fixture) {
	t.Run("SubscribeUserCreated", func(t *testing.T) {
		t.Run("creates user successfully", func(t *testing.T) {
			ctx := context.Background()

			event := infragpt.UserCreatedEvent{
				ClerkUserID: "user_test123",
				Email:       "test@example.com",
				FirstName:   "John",
				LastName:    "Doe",
			}

			err := f.Service().SubscribeUserCreated(ctx, event)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	})

	t.Run("Organization", func(t *testing.T) {
		t.Run("full workflow with metadata", func(t *testing.T) {
			ctx := context.Background()
			svc := f.Service()
			
			// 1. Create user
			userEvent := infragpt.UserCreatedEvent{
				ClerkUserID: "user_workflow123",
				Email:       "workflow@example.com",
				FirstName:   "Jane",
				LastName:    "Smith",
			}
			err := svc.SubscribeUserCreated(ctx, userEvent)
			if err != nil {
				t.Fatalf("failed to create user: %v", err)
			}
			
			// 2. Create organization
			orgEvent := infragpt.OrganizationCreatedEvent{
				ClerkOrgID:      "org_workflow123",
				Name:            "Test Org",
				Slug:            "test-org",
				CreatedByUserID: "user_workflow123",
			}
			err = svc.SubscribeOrganizationCreated(ctx, orgEvent)
			if err != nil {
				t.Fatalf("failed to create organization: %v", err)
			}
			
			// 3. Get organization without metadata
			query := infragpt.OrganizationQuery{ClerkOrgID: "org_workflow123"}
			org, err := svc.Organization(ctx, query)
			if err != nil {
				t.Fatalf("failed to get organization: %v", err)
			}
			
			if org.Name != "Test Org" {
				t.Errorf("expected org name 'Test Org', got '%s'", org.Name)
			}
			
			// 4. Set metadata
			cmd := infragpt.OrganizationMetadataCommand{
				OrganizationID:     org.ID,
				CompanySize:        infragpt.CompanySizeStartup,
				TeamSize:           infragpt.TeamSize1To5,
				UseCases:           []infragpt.UseCase{infragpt.UseCaseInfrastructureMonitoring},
				ObservabilityStack: []infragpt.ObservabilityStack{infragpt.ObservabilityStackDatadog},
			}
			err = svc.SetOrganizationMetadata(ctx, cmd)
			if err != nil {
				t.Fatalf("failed to set metadata: %v", err)
			}
			
			// 5. Get organization with metadata
			org, err = svc.Organization(ctx, query)
			if err != nil {
				t.Fatalf("failed to get organization with metadata: %v", err)
			}
			
			if org.Metadata.CompanySize != infragpt.CompanySizeStartup {
				t.Errorf("expected company size '%s', got '%s'", infragpt.CompanySizeStartup, org.Metadata.CompanySize)
			}
			
			if len(org.Metadata.UseCases) != 1 || org.Metadata.UseCases[0] != infragpt.UseCaseInfrastructureMonitoring {
				t.Errorf("expected use cases [%s], got %v", infragpt.UseCaseInfrastructureMonitoring, org.Metadata.UseCases)
			}
		})
	})

	t.Run("SetOrganizationMetadata", func(t *testing.T) {
		t.Run("sets metadata successfully", func(t *testing.T) {
			t.Skip("skipping - needs organization setup")
		})
	})
}

type fixture interface {
	Service() infragpt.IdentityService
}
