package identitysvctest

import (
	"context"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/domain"
	"sync"
	"testing"
)

type emailService struct {
	emails sync.Map
	domain.UserRepository
}

func (e *emailService) SendEmailVerification(ctx context.Context, email, verificationID string) error {
	key := fmt.Sprintf("email_verification|%s", email)
	e.emails.Store(key, verificationID)
	return nil
}

func (e *emailService) SendResetPassword(ctx context.Context, email, resetID string) error {
	key := fmt.Sprintf("reset_password|%s", email)
	e.emails.Store(key, resetID)
	return nil
}

func (e *emailService) VerifyEmailExternally(t *testing.T, email string) {
	key := fmt.Sprintf("email_verification|%s", email)
	var verificationID string
	if val, ok := e.emails.Load(key); ok {
		verificationID = val.(string)
	} else {
		t.Fatalf("verificationID not found for email %s", email)
	}
	err := e.UserRepository.VerifyUserEmail(context.Background(), verificationID)
	if err != nil {
		t.Fatalf("failed to verify email %s: %v", email, err)
	}
}

func (e *emailService) ResetPasswordExternally(t *testing.T, email, password string) {
	key := fmt.Sprintf("reset_password|%s", email)
	var resetID string
	if val, ok := e.emails.Load(key); ok {
		resetID = val.(string)
	} else {
		t.Fatalf("resetID not found for email %s", email)
	}
	err := e.UserRepository.ValidateResetPasswordToken(context.Background(), resetID)
	if err != nil {
		t.Fatalf("failed to validate reset password resetID for email %s: %v", email, err)
	}

	err = e.UserRepository.ResetPassword(context.Background(), resetID, password)
	if err != nil {
		t.Fatalf("failed to reset password for email %s: %v", email, err)
	}
}
