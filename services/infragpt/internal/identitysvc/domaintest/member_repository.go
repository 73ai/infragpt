package domaintest

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/domain"
)

type memberRepository struct {
	mu      sync.RWMutex
	members map[string]domain.OrganizationMember
}

func NewMemberRepository() domain.MemberRepository {
	return &memberRepository{
		members: make(map[string]domain.OrganizationMember),
	}
}

func (r *memberRepository) Create(ctx context.Context, member domain.OrganizationMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", member.ClerkUserID, member.ClerkOrgID)
	if _, exists := r.members[key]; exists {
		return fmt.Errorf("member relationship already exists for user %s in org %s", member.ClerkUserID, member.ClerkOrgID)
	}

	r.members[key] = member
	return nil
}

func (r *memberRepository) DeleteByClerkIDs(ctx context.Context, clerkUserID string, clerkOrgID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", clerkUserID, clerkOrgID)
	if _, exists := r.members[key]; !exists {
		return fmt.Errorf("member relationship not found for user %s in org %s", clerkUserID, clerkOrgID)
	}

	delete(r.members, key)
	return nil
}

func (r *memberRepository) MembersByOrganizationID(ctx context.Context, organizationID uuid.UUID) ([]*domain.OrganizationMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.OrganizationMember, 0)
	for _, member := range r.members {
		if member.OrganizationID == organizationID {
			memberCopy := member
			result = append(result, &memberCopy)
		}
	}

	return result, nil
}

func (r *memberRepository) MembersByUserClerkID(ctx context.Context, clerkUserID string) ([]*domain.OrganizationMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.OrganizationMember, 0)
	for _, member := range r.members {
		if member.ClerkUserID == clerkUserID {
			memberCopy := member
			result = append(result, &memberCopy)
		}
	}

	return result, nil
}