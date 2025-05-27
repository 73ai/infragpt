package domain

import (
	"context"
)

type User struct {
	UserID       string
	Email        string
	PasswordHash string
}

type UserRepository interface {
	CreateUser(ctx context.Context, user User) (verificationID string, err error)
	UserByEmail(ctx context.Context, email string) (User, error)

	VerifyUserEmail(ctx context.Context, verificationID string) error
	RequestResetPassword(ctx context.Context, userID string) (resetID string, err error)

	ValidateResetPasswordToken(ctx context.Context, resetID string) error

	ResetPassword(ctx context.Context, token, password string) error
}
