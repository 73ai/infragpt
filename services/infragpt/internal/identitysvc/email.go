package identitysvc

import "context"

type EmailService interface {
	SendEmailVerification(ctx context.Context, email, verificationID string) error
	SendResetPassword(ctx context.Context, email, resetID string) error
}
