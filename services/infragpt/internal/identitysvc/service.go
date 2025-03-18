package identitysvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/domain"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt/identity"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	sessionRepo       domain.SessionRepository
	userRepo          domain.UserRepository
	emailSVC          EmailService
	tokenManager      TokenManager
	googleAuthGateway domain.GoogleAuthGateway
}

var _ identity.Service = (*service)(nil)

func (s service) CreateUser(ctx context.Context, cmd identity.CreateUserCommand) (identity.Credentials, error) {
	if err := ValidateEmail(cmd.Email); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate email: %w", err)
	}
	if err := ValidatePassword(cmd.Password); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate password: %w", err)
	}
	if err := ValidateIP(cmd.IP); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate ip: %w", err)
	}
	if err := ValidateUserAgent(cmd.UserAgent); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate user agent: %w", err)
	}
	if err := ValidateTimezone(cmd.Timezone); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate timezone: %w", err)
	}
	if err := ValidateOS(cmd.OS); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate os: %w", err)
	}
	if err := ValidateDeviceFingerprint(cmd.DeviceFingerprint); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate device fingerprint: %w", err)
	}

	if cmd.Brand == "" {
		cmd.Brand = "Unknown Brand"
	}

	// TODO: move this bcrypt hash to the repository
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("hash password: %w", err)
	}

	uid := newUserID()
	verificationID, err := s.userRepo.CreateUser(ctx, domain.User{
		UserID:       uid,
		Email:        cmd.Email,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("create user: %w", err)
	}

	// send email verification
	err = s.emailSVC.SendEmailVerification(ctx, cmd.Email, verificationID)
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("send email verification: %w", err)
	}

	sessionID := newSessionID()
	credentials, err := s.sessionRepo.StartUserSession(ctx, identity.UserSession{
		UserID:         uid,
		SessionID:      sessionID,
		IPAddress:      cmd.IP,
		UserAgent:      cmd.UserAgent,
		IPCountryISO:   cmd.IPCountryISO,
		LastActivityAt: time.Now(),
		Timezone:       cmd.Timezone,
		Device: identity.UserDevice{
			UserID:            uid,
			DeviceFingerprint: cmd.DeviceFingerprint,
			OS:                cmd.OS,
			Name:              fmt.Sprintf("%s %s", cmd.Brand, cmd.OS),
			Brand:             cmd.Brand,
		},
	})
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("start user session: %w", err)
	}

	return credentials, nil
}

func (s service) Login(ctx context.Context, cmd identity.LoginCommand) (identity.Credentials, error) {
	if err := ValidateEmail(cmd.Email); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate email: %w", err)
	}
	if err := ValidatePassword(cmd.Password); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate password: %w", err)
	}
	if err := ValidateIP(cmd.IP); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate ip: %w", err)
	}
	if err := ValidateUserAgent(cmd.UserAgent); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate user agent: %w", err)
	}
	if err := ValidateTimezone(cmd.Timezone); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate timezone: %w", err)
	}
	if err := ValidateOS(cmd.OS); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate os: %w", err)
	}
	if err := ValidateDeviceFingerprint(cmd.DeviceFingerprint); err != nil {
		return identity.Credentials{}, fmt.Errorf("validate device fingerprint: %w", err)
	}

	if cmd.Brand == "" {
		cmd.Brand = "Unknown Brand"
	}

	user, err := s.userRepo.UserByEmail(ctx, cmd.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return identity.Credentials{}, identity.ErrInvalidEmailOrPassword
		}
		return identity.Credentials{}, fmt.Errorf("get user by email: %w", err)
	}

	// TODO: move this bcrypt compare to the repository
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cmd.Password))
	if err != nil {
		return identity.Credentials{}, identity.ErrInvalidEmailOrPassword
	}

	sessionID := newSessionID()
	credentials, err := s.sessionRepo.StartUserSession(ctx, identity.UserSession{
		UserID:         user.UserID,
		SessionID:      sessionID,
		IPAddress:      cmd.IP,
		UserAgent:      cmd.UserAgent,
		IPCountryISO:   cmd.IPCountryISO,
		LastActivityAt: time.Now(),
		Timezone:       cmd.Timezone,
		Device: identity.UserDevice{
			UserID:            user.UserID,
			DeviceFingerprint: cmd.DeviceFingerprint,
			OS:                cmd.OS,
			Name:              fmt.Sprintf("%s %s", cmd.Brand, cmd.OS),
			Brand:             cmd.Brand,
		},
	})
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("start user session: %w", err)
	}

	return credentials, nil
}

func (s service) RefreshCredentials(ctx context.Context, cmd identity.RefreshCredentialsCommand) (identity.Credentials, error) {
	if cmd.RefreshToken == "" {
		return identity.Credentials{}, identity.ErrRefreshTokenCannotBeEmpty
	}
	tokenID, err := s.tokenManager.ValidateRefreshToken(cmd.RefreshToken)
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("validate refresh token: %w", err)
	}

	creds, err := s.sessionRepo.RefreshToken(ctx, tokenID)
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("refresh token: %w", err)
	}

	return creds, nil
}

func (s service) ValidateCredentials(ctx context.Context, query identity.ValidateCredentialsQuery) error {
	if query.AuthorizationHeader == "" {
		return identity.ErrAuthorizationHeaderCannotBeEmpty
	}
	accessToken := strings.TrimPrefix(query.AuthorizationHeader, "Bearer ")
	if accessToken == "" {
		return identity.ErrAccessTokenCannotBeEmpty
	}

	_, err := s.tokenManager.ValidateAccessToken(accessToken)
	if err != nil {
		return fmt.Errorf("validate access token: %w", err)
	}

	return nil
}

func (s service) VerifyEmail(ctx context.Context, cmd identity.VerifyEmailCommand) error {
	if cmd.VerificationID == "" {
		return identity.ErrEmailVerificationIDCannotBeEmpty
	}
	if err := s.userRepo.VerifyUserEmail(ctx, cmd.VerificationID); err != nil {
		return fmt.Errorf("verify email: %w", err)
	}

	return nil
}

func (s service) RequestResetPassword(ctx context.Context, cmd identity.RequestResetPasswordCommand) error {
	if cmd.Email == "" {
		return identity.ErrEmailCannotBeEmpty
	}
	if err := ValidateEmail(cmd.Email); err != nil {
		return fmt.Errorf("validate email: %w", err)
	}
	user, err := s.userRepo.UserByEmail(ctx, cmd.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return identity.ErrEmailNotRegistered
		}
		return fmt.Errorf("get user by email: %w", err)
	}

	resetID, err := s.userRepo.RequestResetPassword(ctx, user.UserID)
	if err != nil {
		return fmt.Errorf("request password reset: %w", err)
	}

	// send reset password email
	err = s.emailSVC.SendResetPassword(ctx, cmd.Email, resetID)
	if err != nil {
		return fmt.Errorf("send reset password email: %w", err)
	}

	return nil
}

func (s service) ValidateResetPasswordToken(ctx context.Context, query identity.ValidateResetPasswordTokenQuery) error {
	if query.ResetID == "" {
		return identity.ErrResetPasswordIDCannotBeEmpty
	}
	err := s.userRepo.ValidateResetPasswordToken(ctx, query.ResetID)
	if err != nil {
		return fmt.Errorf("validate password reset token: %w", err)
	}
	return nil
}

// TODO: add uuid validation check for most uuid type fields before sending to repository
func (s service) ResetPassword(ctx context.Context, cmd identity.ResetPasswordCommand) error {
	if cmd.ResetID == "" {
		return identity.ErrResetPasswordIDCannotBeEmpty
	}

	if err := s.userRepo.ValidateResetPasswordToken(ctx, cmd.ResetID); err != nil {
		return fmt.Errorf("validate password reset token: %w", err)
	}
	if err := ValidatePassword(cmd.Password); err != nil {
		return fmt.Errorf("validate password: %w", err)
	}
	err := s.userRepo.ResetPassword(ctx, cmd.ResetID, cmd.Password)
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	return nil
}

func (s service) UserSessions(ctx context.Context, query identity.UserSessionsQuery) ([]identity.UserSession, error) {
	if query.AuthorizationHeader == "" {
		return nil, identity.ErrAuthorizationHeaderCannotBeEmpty
	}
	// TODO: move this entire access token validation logic to middleware
	accessToken := strings.TrimPrefix(query.AuthorizationHeader, "Bearer ")
	if accessToken == "" {
		return nil, identity.ErrAccessTokenCannotBeEmpty
	}

	sid, err := s.tokenManager.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("validate access token: %w", err)
	}

	session, err := s.sessionRepo.UserSession(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("user session: %w", err)
	}
	sessions, err := s.sessionRepo.UserSessions(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user sessions: %w", err)
	}

	return sessions, nil
}

func (s service) InitiateGoogleAuth(ctx context.Context) (string, error) {
	url, err := s.googleAuthGateway.AuthURL(ctx)
	if err != nil {
		return "", fmt.Errorf("get Google Auth URL: %w", err)
	}

	return url, nil
}

func (s service) CompleteGoogleAuth(ctx context.Context, cmd identity.CompleteGoogleAuthCommand) (identity.Credentials, error) {
	code := cmd.Code
	state := cmd.State

	if code == "" {
		return identity.Credentials{}, identity.ErrGoogleAuthCodeCannotBeEmpty
	}
	if state == "" {
		return identity.Credentials{}, identity.ErrGoogleAuthStateCannotBeEmpty
	}

	profile, err := s.googleAuthGateway.CompleteAuth(ctx, code, state)
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("complete Google Auth: %w", err)
	}

	// check if user exists
	user, err := s.userRepo.UserByEmail(ctx, profile.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return identity.Credentials{}, fmt.Errorf("get user by email: %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		// create user
		uid := newUserID()
		verificationID, err := s.userRepo.CreateUser(ctx, domain.User{
			UserID: uid,
			Email:  profile.Email,
		})
		if err != nil {
			return identity.Credentials{}, fmt.Errorf("create user: %w", err)
		}
		// send email verification
		err = s.userRepo.VerifyUserEmail(ctx, verificationID)
		if err != nil {
			return identity.Credentials{}, fmt.Errorf("email verification failed: %w", err)
		}
		user = domain.User{
			UserID: uid,
			Email:  profile.Email,
		}
	}

	sessionID := newSessionID()
	credentials, err := s.sessionRepo.StartUserSession(ctx, identity.UserSession{
		UserID:    user.UserID,
		SessionID: sessionID,
	})
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("start user session: %w", err)
	}

	return credentials, nil
}

func newUserID() string {
	return uuid.New().String()
}

func newSessionID() string {
	return uuid.New().String()
}
