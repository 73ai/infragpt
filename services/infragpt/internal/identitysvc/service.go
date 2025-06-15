package identitysvc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/domain"
)

type service struct {
	userRepo         domain.UserRepository
	organizationRepo domain.OrganizationRepository
	memberRepo       domain.MemberRepository
	authService      domain.AuthService
}

func (s *service) Subscribe(ctx context.Context) error {
	return s.authService.Subscribe(ctx, func(ctx context.Context, event any) error {
		switch e := event.(type) {
		case infragpt.UserCreatedEvent:
			return s.reconcileUserCreated(ctx, e)
		case infragpt.UserUpdatedEvent:
			return s.reconcileUserUpdated(ctx, e)
		case infragpt.UserDeletedEvent:
			return s.reconcileUserDeleted(ctx, e)
		case infragpt.OrganizationCreatedEvent:
			return s.reconcileOrganizationCreated(ctx, e)
		case infragpt.OrganizationUpdatedEvent:
			return s.reconcileOrganizationUpdated(ctx, e)
		case infragpt.OrganizationDeletedEvent:
			return s.reconcileOrganizationDeleted(ctx, e)
		case infragpt.OrganizationMemberAddedEvent:
			return s.reconcileOrganizationMemberAdded(ctx, e)
		case infragpt.OrganizationMemberUpdatedEvent:
			return s.reconcileOrganizationMemberUpdated(ctx, e)
		case infragpt.OrganizationMemberDeletedEvent:
			return s.reconcileOrganizationMemberDeleted(ctx, e)
		default:
			return fmt.Errorf("unknown event type: %T", e)
		}

	})
}

func (s *service) reconcileUserCreated(ctx context.Context, event infragpt.UserCreatedEvent) error {
	user := domain.User{
		ID:          uuid.New(),
		ClerkUserID: event.ClerkUserID,
		Email:       event.Email,
		FirstName:   event.FirstName,
		LastName:    event.LastName,
	}

	return s.userRepo.Create(ctx, user)
}

func (s *service) SubscribeUserCreated(ctx context.Context, event infragpt.UserCreatedEvent) error {
	panic("not allowed")
}

func (s *service) reconcileUserUpdated(ctx context.Context, event infragpt.UserUpdatedEvent) error {
	user := domain.User{
		Email:     event.Email,
		FirstName: event.FirstName,
		LastName:  event.LastName,
	}

	return s.userRepo.Update(ctx, event.ClerkUserID, user)
}

func (s *service) SubscribeUserUpdated(ctx context.Context, event infragpt.UserUpdatedEvent) error {
	panic("not allowed")
}

func (s *service) reconcileUserDeleted(ctx context.Context, event infragpt.UserDeletedEvent) error {
	return s.userRepo.DeleteByClerkID(ctx, event.ClerkUserID)
}

func (s *service) SubscribeUserDeleted(ctx context.Context, event infragpt.UserDeletedEvent) error {
	panic("not allowed")
}

func (s *service) reconcileOrganizationCreated(ctx context.Context, event infragpt.OrganizationCreatedEvent) error {
	createdByUser, err := s.userRepo.UserByClerkID(ctx, event.CreatedByUserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	org := domain.Organization{
		ID:              uuid.New(),
		ClerkOrgID:      event.ClerkOrgID,
		Name:            event.Name,
		Slug:            event.Slug,
		CreatedByUserID: createdByUser.ID,
	}

	err = s.organizationRepo.Create(ctx, org)
	if err != nil {
		return fmt.Errorf("organization created: %w", err)
	}

	member := domain.OrganizationMember{
		UserID:         createdByUser.ID,
		OrganizationID: org.ID,
		ClerkUserID:    event.CreatedByUserID,
		ClerkOrgID:     event.ClerkOrgID,
		Role:           "admin",
	}

	return s.memberRepo.Create(ctx, member)
}

func (s *service) SubscribeOrganizationCreated(ctx context.Context, event infragpt.OrganizationCreatedEvent) error {
	panic("not allowed")
}

func (s *service) reconcileOrganizationUpdated(ctx context.Context, event infragpt.OrganizationUpdatedEvent) error {
	org := domain.Organization{
		Name: event.Name,
		Slug: event.Slug,
	}

	return s.organizationRepo.Update(ctx, event.ClerkOrgID, org)
}

func (s *service) SubscribeOrganizationUpdated(ctx context.Context, event infragpt.OrganizationUpdatedEvent) error {
	panic("not allowed")
}

func (s *service) reconcileOrganizationDeleted(ctx context.Context, event infragpt.OrganizationDeletedEvent) error {
	return s.organizationRepo.DeleteByClerkID(ctx, event.ClerkOrgID)
}

func (s *service) SubscribeOrganizationDeleted(ctx context.Context, event infragpt.OrganizationDeletedEvent) error {
	panic("not allowed")
}

func (s *service) reconcileOrganizationMemberAdded(ctx context.Context, event infragpt.OrganizationMemberAddedEvent) error {
	user, err := s.userRepo.UserByClerkID(ctx, event.ClerkUserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	org, err := s.organizationRepo.OrganizationByClerkID(ctx, event.ClerkOrgID)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	member := domain.OrganizationMember{
		UserID:         user.ID,
		OrganizationID: org.ID,
		ClerkUserID:    event.ClerkUserID,
		ClerkOrgID:     event.ClerkOrgID,
		Role:           event.Role,
	}

	return s.memberRepo.Create(ctx, member)
}

func (s *service) SubscribeOrganizationMemberAdded(ctx context.Context, event infragpt.OrganizationMemberAddedEvent) error {
	panic("not allowed")
}

func (s *service) reconcileOrganizationMemberUpdated(ctx context.Context, event infragpt.OrganizationMemberUpdatedEvent) error {
	return s.memberRepo.UpdateByClerkIDs(ctx, event.ClerkUserID, event.ClerkOrgID, event.Role)
}

func (s *service) SubscribeOrganizationMemberUpdated(ctx context.Context, event infragpt.OrganizationMemberUpdatedEvent) error {
	panic("not allowed")
}

func (s *service) reconcileOrganizationMemberDeleted(ctx context.Context, event infragpt.OrganizationMemberDeletedEvent) error {
	return s.memberRepo.DeleteByClerkIDs(ctx, event.ClerkUserID, event.ClerkOrgID)
}

func (s *service) SubscribeOrganizationMemberDeleted(ctx context.Context, event infragpt.OrganizationMemberDeletedEvent) error {
	panic("not allowed")
}

func (s *service) SetOrganizationMetadata(ctx context.Context, cmd infragpt.OrganizationMetadataCommand) error {
	metadata := domain.OrganizationMetadata{
		OrganizationID:     cmd.OrganizationID,
		CompanySize:        cmd.CompanySize,
		TeamSize:           cmd.TeamSize,
		UseCases:           cmd.UseCases,
		ObservabilityStack: cmd.ObservabilityStack,
	}

	return s.organizationRepo.SetMetadata(ctx, cmd.OrganizationID, metadata)
}

func (s *service) Organization(ctx context.Context, query infragpt.OrganizationQuery) (infragpt.Organization, error) {
	org, err := s.organizationRepo.OrganizationByClerkID(ctx, query.ClerkOrgID)
	if err != nil {
		return infragpt.Organization{}, err
	}

	return infragpt.Organization{
		ID:              org.ID,
		ClerkOrgID:      org.ClerkOrgID,
		Name:            org.Name,
		Slug:            org.Slug,
		CreatedByUserID: org.CreatedByUserID,
		CreatedAt:       org.CreatedAt,
		UpdatedAt:       org.UpdatedAt,
		Metadata: infragpt.OrganizationMetadata{
			OrganizationID:     org.Metadata.OrganizationID,
			CompanySize:        org.Metadata.CompanySize,
			TeamSize:           org.Metadata.TeamSize,
			UseCases:           org.Metadata.UseCases,
			ObservabilityStack: org.Metadata.ObservabilityStack,
			CompletedAt:        org.Metadata.CompletedAt,
			UpdatedAt:          org.Metadata.UpdatedAt,
		},
	}, nil
}
