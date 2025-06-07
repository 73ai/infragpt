package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(context.Context, User) error
	UserByClerkID(ctx context.Context, clerkUserID string) (*User, error)
	Update(ctx context.Context, clerkUserID string, user User) error
}

type User struct {
	ID          uuid.UUID
	ClerkUserID string
	Email       string
	FirstName   string
	LastName    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}