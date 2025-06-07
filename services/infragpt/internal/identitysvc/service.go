package identitysvc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/domain"
)

type Service struct {
	userRepo         domain.UserRepository
	organizationRepo domain.OrganizationRepository
	memberRepo       domain.MemberRepository
}

func New(
	userRepo domain.UserRepository,
	organizationRepo domain.OrganizationRepository,
	memberRepo domain.MemberRepository,
) *Service {
	return &Service{
		userRepo:         userRepo,
		organizationRepo: organizationRepo,
		memberRepo:       memberRepo,
	}
}

func (s *Service) SubscribeUserCreated(ctx context.Context, event infragpt.UserCreatedEvent) error {
	user := domain.User{
		ID:          uuid.New(),
		ClerkUserID: event.ClerkUserID,
		Email:       event.Email,
		FirstName:   event.FirstName,
		LastName:    event.LastName,
	}

	return s.userRepo.Create(ctx, user)
}

func (s *Service) SubscribeUserUpdated(ctx context.Context, event infragpt.UserUpdatedEvent) error {
	user := domain.User{
		Email:     event.Email,
		FirstName: event.FirstName,
		LastName:  event.LastName,
	}

	return s.userRepo.Update(ctx, event.ClerkUserID, user)
}

func (s *Service) SubscribeOrganizationCreated(ctx context.Context, event infragpt.OrganizationCreatedEvent) error {
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
		return err
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

func (s *Service) SubscribeOrganizationUpdated(ctx context.Context, event infragpt.OrganizationUpdatedEvent) error {
	org := domain.Organization{
		Name: event.Name,
		Slug: event.Slug,
	}

	return s.organizationRepo.Update(ctx, event.ClerkOrgID, org)
}

func (s *Service) SubscribeOrganizationMemberAdded(ctx context.Context, event infragpt.OrganizationMemberAddedEvent) error {
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

func (s *Service) SubscribeOrganizationMemberRemoved(ctx context.Context, event infragpt.OrganizationMemberRemovedEvent) error {
	return s.memberRepo.DeleteByClerkIDs(ctx, event.ClerkUserID, event.ClerkOrgID)
}

func (s *Service) SetOrganizationMetadata(ctx context.Context, cmd infragpt.OrganizationMetadataCommand) error {
	metadata := domain.OrganizationMetadata{
		OrganizationID:     cmd.OrganizationID,
		CompanySize:        cmd.CompanySize,
		TeamSize:           cmd.TeamSize,
		UseCases:           cmd.UseCases,
		ObservabilityStack: cmd.ObservabilityStack,
	}

	return s.organizationRepo.SetMetadata(ctx, cmd.OrganizationID, metadata)
}

func (s *Service) Organization(ctx context.Context, query infragpt.OrganizationQuery) (infragpt.Organization, error) {
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