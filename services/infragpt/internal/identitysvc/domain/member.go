package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MemberRepository interface {
	Create(context.Context, OrganizationMember) error
	DeleteByClerkIDs(ctx context.Context, clerkUserID string, clerkOrgID string) error
	MembersByOrganizationID(ctx context.Context, organizationID uuid.UUID) ([]*OrganizationMember, error)
	MembersByUserClerkID(ctx context.Context, clerkUserID string) ([]*OrganizationMember, error)
}

type OrganizationMember struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	ClerkUserID    string
	ClerkOrgID     string
	Role           string
	JoinedAt       time.Time
}