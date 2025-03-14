package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/domain"
	db2 "github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/supporting/postgres/db"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt/identity"
	"golang.org/x/crypto/bcrypt"
)

type userRepository struct {
	db      *sql.DB
	queries *db2.Queries
}

func (u userRepository) ResetPassword(ctx context.Context, token, password string) error {
	tx, err := u.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback()
	qtx := u.queries.WithTx(tx)

	resetID, _ := uuid.Parse(token)
	resetPassword, err := qtx.ResetPassword(ctx, resetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return identity.ErrInvalidResetPasswordToken
		}
		return fmt.Errorf("get password reset: %w", err)
	}

	if resetPassword.ExpiryAt.Before(time.Now()) {
		return identity.ErrResetPasswordTokenExpired
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	err = qtx.SetNewPassword(ctx, db2.SetNewPasswordParams{
		UserID:       resetPassword.UserID,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (u userRepository) CreateUser(ctx context.Context, user domain.User) (string, error) {
	tx, err := u.db.Begin()
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback()
	qtx := u.queries.WithTx(tx)

	uid, _ := uuid.Parse(user.UserID)
	err = qtx.CreateUser(ctx, db2.CreateUserParams{
		UserID:       uid,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	})
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}

	// create email verification token
	vid := newEmailVerificationID()
	err = qtx.CreateEmailVerification(ctx, db2.CreateEmailVerificationParams{
		VerificationID: vid,
		UserID:         uid,
		Email:          user.Email,
		ExpiryAt:       newEmailVerificationExpiry(),
	})
	if err != nil {
		return "", fmt.Errorf("create email verification: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("commit transaction: %w", err)
	}

	return vid.String(), nil
}

func (u userRepository) UserByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := u.queries.UserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return domain.User{
		UserID:       user.UserID.String(),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}, nil
}

func (u userRepository) VerifyUserEmail(ctx context.Context, verificationID string) error {
	tx, err := u.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback()
	qtx := u.queries.WithTx(tx)

	vid, _ := uuid.Parse(verificationID)
	emailVerification, err := qtx.EmailVerification(ctx, vid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return identity.ErrInvalidEmailVerificationID
		}
		return fmt.Errorf("get email verification: %w", err)
	}

	if emailVerification.ExpiryAt.Before(time.Now()) {
		return identity.ErrorEmailVerificationExpired
	}

	user, err := qtx.UserByID(ctx, emailVerification.UserID)
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	if user.IsEmailVerified {
		return identity.ErrEmailAlreadyVerified
	}

	err = qtx.VerifyEmail(ctx, vid)
	if err != nil {
		return fmt.Errorf("verify email: %w", err)
	}

	err = qtx.MarkEmailVerificationAsExpired(ctx, emailVerification.VerificationID)
	if err != nil {
		return fmt.Errorf("mark email verification as expired: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (u userRepository) RequestResetPassword(ctx context.Context, userID string) (string, error) {
	tx, err := u.db.Begin()
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback()
	qtx := u.queries.WithTx(tx)

	uid := uuid.MustParse(userID)
	rid := newResetID()
	err = qtx.CreateResetPassword(ctx, db2.CreateResetPasswordParams{
		UserID:   uid,
		ResetID:  rid,
		ExpiryAt: time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		return "", fmt.Errorf("create password reset: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("commit transaction: %w", err)
	}

	return rid.String(), nil
}

func (u userRepository) ValidateResetPasswordToken(ctx context.Context, token string) error {
	tx, err := u.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback()
	qtx := u.queries.WithTx(tx)

	resetID, _ := uuid.Parse(token)
	resetPassword, err := qtx.ResetPassword(ctx, resetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return identity.ErrInvalidResetPasswordToken
		}
		return fmt.Errorf("get password reset: %w", err)
	}

	if resetPassword.ExpiryAt.Before(time.Now()) {
		return identity.ErrResetPasswordTokenExpired
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func NewUserRepository(sqlDB *sql.DB) domain.UserRepository {
	return &userRepository{
		db:      sqlDB,
		queries: db2.New(sqlDB),
	}
}

func newEmailVerificationExpiry() time.Time {
	return time.Now().Add(24 * time.Hour)
}

func newEmailVerificationID() uuid.UUID {
	return uuid.New()
}

func newResetID() uuid.UUID {
	return uuid.New()
}
