package identitytest

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/supporting/token"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt/identity"
)

func Ensure(t *testing.T, svc identity.Service, externalActions ExternalActions) {
	t.Run("CreateUser", func(t *testing.T) {
		ctx := context.Background()
		t.Run("Fails if email is empty", func(t *testing.T) {
			cmd := identity.CreateUserCommand{
				Email:    "",
				Password: "password",
			}
			_, err := svc.CreateUser(ctx, cmd)
			if !errors.Is(err, identity.ErrEmailCannotBeEmpty) {
				t.Errorf("CreateUser(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrEmailCannotBeEmpty)
			}
		})
		t.Run("Fails if email is invalid", func(t *testing.T) {
			cmd := identity.CreateUserCommand{
				Email:    "invalid",
				Password: "password",
			}
			_, err := svc.CreateUser(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidEmail) {
				t.Errorf("CreateUser(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidEmail)
			}
		})
		t.Run("Fails if password is empty", func(t *testing.T) {
			cmd := identity.CreateUserCommand{
				Email:    "team@balancewise.app",
				Password: "",
			}
			_, err := svc.CreateUser(ctx, cmd)
			if !errors.Is(err, identity.ErrPasswordCannotBeEmpty) {
				t.Errorf("CreateUser(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrPasswordCannotBeEmpty)
			}
		})
		t.Run("Fails if password is invalid", func(t *testing.T) {
			cmd := identity.CreateUserCommand{
				Email:    "team@balanewise.app",
				Password: "short",
			}
			_, err := svc.CreateUser(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidPassword) {
				t.Errorf("CreateUser(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidPassword)
			}
		})
		t.Run("Fails if IP is empty", func(t *testing.T) {
			cmd := identity.CreateUserCommand{
				Email:    "team@balancewise.app",
				Password: "123@Balancewise",
			}
			_, err := svc.CreateUser(ctx, cmd)
			if !errors.Is(err, identity.ErrIPCannotBeEmpty) {
				t.Errorf("CreateUser(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrIPCannotBeEmpty)
			}
		})
		t.Run("Fails if IP is invalid", func(t *testing.T) {
			cmd := identity.CreateUserCommand{
				Email:    "team@balancewise.app",
				Password: "123@Balancewise",
				IP:       "1234",
			}
			_, err := svc.CreateUser(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidIP) {
				t.Errorf("CreateUser(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidIP)
			}
		})
		t.Run("Fails if UserAgent is empty", func(t *testing.T) {
			cmd := identity.CreateUserCommand{
				Email:    "team@balancewise.app",
				Password: "123@Balancewise",
				IP:       "1.1.1.1",
			}
			_, err := svc.CreateUser(ctx, cmd)
			if !errors.Is(err, identity.ErrUserAgentCannotBeEmpty) {
				t.Errorf("CreateUser(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrUserAgentCannotBeEmpty)
			}
		})
		t.Run("Fails if UserAgent is invalid", func(t *testing.T) {
			cmd := identity.CreateUserCommand{
				Email:     "team@balancewise.app",
				Password:  "123@Balancewise",
				IP:        "1.1.1.1",
				UserAgent: "invalid",
			}
			_, err := svc.CreateUser(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidUserAgent) {
				t.Errorf("CreateUser(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidUserAgent)
			}
		})
		t.Run("Fails if Timezone is empty", func(t *testing.T) {
			cmd := identity.CreateUserCommand{
				Email:     "team@balancewise.app",
				Password:  "123@Balancewise",
				IP:        "1.1.1.1",
				UserAgent: "Mozilla/5.0",
				Timezone:  &time.Location{},
			}
			_, err := svc.CreateUser(ctx, cmd)
			if !errors.Is(err, identity.ErrTimezoneCannotBeEmpty) {
				t.Errorf("CreateUser(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrTimezoneCannotBeEmpty)
			}
		})
		t.Run("Fails if OS is invalid", func(t *testing.T) {
			cmd := identity.CreateUserCommand{
				Email:     "team@balancewise.app",
				Password:  "123@Balancewise",
				IP:        "1.1.1.1",
				UserAgent: "Mozilla/5.0",
				Timezone:  time.UTC,
			}
			_, err := svc.CreateUser(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidOS) {
				t.Errorf("CreateUser(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidOS)
			}
		})
		t.Run("Fails if DeviceFingerprint is empty", func(t *testing.T) {
			cmd := identity.CreateUserCommand{
				Email:     "team@balancewise.app",
				Password:  "123@Balancewise",
				IP:        "1.1.1.1",
				UserAgent: "Mozilla/5.0",
				Timezone:  time.UTC,
				OS:        identity.OperatingSystemAndroid,
			}
			_, err := svc.CreateUser(ctx, cmd)
			if !errors.Is(err, identity.ErrDeviceFingerprintCannotBeEmpty) {
				t.Errorf("CreateUser(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrDeviceFingerprintCannotBeEmpty)
			}
		})
		t.Run("Success", func(t *testing.T) {
			cmd := identity.CreateUserCommand{
				Email:             "team@balancewise.app",
				Password:          "123@Balancewise",
				IP:                "1.1.1.1",
				UserAgent:         "Mozilla/5.0",
				Timezone:          time.UTC,
				OS:                identity.OperatingSystemAndroid,
				DeviceFingerprint: someValidDeviceFingerprint(),
			}
			_, err := svc.CreateUser(ctx, cmd)
			if err != nil {
				t.Errorf("CreateUser(%v) error = %v, wantErr nil", cmd, err)
			}
		})
	})

	t.Run("Login", func(t *testing.T) {
		ctx := context.Background()
		t.Run("Fails if email is empty", func(t *testing.T) {
			cmd := identity.LoginCommand{
				Email:    "",
				Password: "password",
			}
			_, err := svc.Login(ctx, cmd)
			if !errors.Is(err, identity.ErrEmailCannotBeEmpty) {
				t.Errorf("Login(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrEmailCannotBeEmpty)
			}
		})
		t.Run("Fails if email is invalid", func(t *testing.T) {
			cmd := identity.LoginCommand{
				Email:    "invalid",
				Password: "password",
			}
			_, err := svc.Login(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidEmail) {
				t.Errorf("Login(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidEmail)
			}
		})
		t.Run("Fails if password is empty", func(t *testing.T) {
			cmd := identity.LoginCommand{
				Email:    "team@balancewise.app",
				Password: "",
			}
			_, err := svc.Login(ctx, cmd)
			if !errors.Is(err, identity.ErrPasswordCannotBeEmpty) {
				t.Errorf("Login(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrPasswordCannotBeEmpty)
			}
		})
		t.Run("Fails if ip is empty", func(t *testing.T) {
			cmd := identity.LoginCommand{
				Email:    someValidEmail(),
				Password: someValidPassword(),
				IP:       "",
			}
			_, err := svc.Login(ctx, cmd)
			if !errors.Is(err, identity.ErrIPCannotBeEmpty) {
				t.Errorf("Login(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrIPCannotBeEmpty)
			}
		})
		t.Run("Fails if ip is invalid", func(t *testing.T) {
			cmd := identity.LoginCommand{
				Email:    someValidEmail(),
				Password: someValidPassword(),
				IP:       "invalid",
			}
			_, err := svc.Login(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidIP) {
				t.Errorf("Login(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidIP)
			}
		})
		t.Run("Fails if user agent is empty", func(t *testing.T) {
			cmd := identity.LoginCommand{
				Email:     someValidEmail(),
				Password:  someValidPassword(),
				IP:        someValidIP(),
				UserAgent: "",
			}
			_, err := svc.Login(ctx, cmd)
			if !errors.Is(err, identity.ErrUserAgentCannotBeEmpty) {
				t.Errorf("Login(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrUserAgentCannotBeEmpty)
			}
		})
		t.Run("Fails if user agent is invalid", func(t *testing.T) {
			cmd := identity.LoginCommand{
				Email:     someValidEmail(),
				Password:  someValidPassword(),
				IP:        someValidIP(),
				UserAgent: "invalid",
			}
			_, err := svc.Login(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidUserAgent) {
				t.Errorf("Login(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidUserAgent)
			}
		})
		t.Run("Fails if timezone is empty", func(t *testing.T) {
			cmd := identity.LoginCommand{
				Email:     someValidEmail(),
				Password:  someValidPassword(),
				IP:        someValidIP(),
				UserAgent: someValidUserAgent(),
				Timezone:  &time.Location{},
			}
			_, err := svc.Login(ctx, cmd)
			if !errors.Is(err, identity.ErrTimezoneCannotBeEmpty) {
				t.Errorf("Login(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrTimezoneCannotBeEmpty)
			}
		})
		t.Run("Fails if os is invalid", func(t *testing.T) {
			cmd := identity.LoginCommand{
				Email:     someValidEmail(),
				Password:  someValidPassword(),
				IP:        someValidIP(),
				UserAgent: someValidUserAgent(),
				Timezone:  time.UTC,
				OS:        "invalid",
			}
			_, err := svc.Login(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidOS) {
				t.Errorf("Login(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidOS)
			}
		})
		t.Run("Fails if device fingerprint is empty", func(t *testing.T) {
			cmd := identity.LoginCommand{
				Email:             someValidEmail(),
				Password:          someValidPassword(),
				IP:                someValidIP(),
				UserAgent:         someValidUserAgent(),
				Timezone:          time.UTC,
				OS:                identity.OperatingSystemAndroid,
				DeviceFingerprint: "",
			}
			_, err := svc.Login(ctx, cmd)
			if !errors.Is(err, identity.ErrDeviceFingerprintCannotBeEmpty) {
				t.Errorf("Login(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrDeviceFingerprintCannotBeEmpty)
			}
		})
		t.Run("Success", func(t *testing.T) {
			email := someValidEmail()
			password := someValidPassword()
			signupUser(t, svc, email, password)
			cmd := identity.LoginCommand{
				Email:             email,
				Password:          password,
				IP:                someValidIP(),
				UserAgent:         someValidUserAgent(),
				Timezone:          time.UTC,
				OS:                identity.OperatingSystemAndroid,
				DeviceFingerprint: someValidDeviceFingerprint(),
			}
			_, err := svc.Login(ctx, cmd)
			if err != nil {
				t.Errorf("Login(%v) error = %v, wantErr nil", cmd, err)
			}
		})
	})

	t.Run("RefreshCredentials", func(t *testing.T) {
		ctx := context.Background()
		t.Run("Fails if refresh token is empty", func(t *testing.T) {
			cmd := identity.RefreshCredentialsCommand{
				RefreshToken: "",
			}
			_, err := svc.RefreshCredentials(ctx, cmd)
			if !errors.Is(err, identity.ErrRefreshTokenCannotBeEmpty) {
				t.Errorf("RefreshCredentials(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrRefreshTokenCannotBeEmpty)
			}
		})
		t.Run("Fails if refresh token is some random string", func(t *testing.T) {
			cmd := identity.RefreshCredentialsCommand{
				RefreshToken: "invalid",
			}
			_, err := svc.RefreshCredentials(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidRefreshToken) {
				t.Errorf("RefreshCredentials(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidRefreshToken)
			}
		})
		t.Run("Fails if refresh token is invalid token", func(t *testing.T) {
			cmd := identity.RefreshCredentialsCommand{
				RefreshToken: someInvalidRefreshToken(t),
			}
			_, err := svc.RefreshCredentials(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidRefreshToken) {
				t.Errorf("RefreshCredentials(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidRefreshToken)
			}
		})
		// TODO: Add test for expired refresh token
		t.Run("Success", func(t *testing.T) {
			email := someValidEmail()
			password := someValidPassword()
			signupUser(t, svc, email, password)
			loginCmd := identity.LoginCommand{
				Email:             email,
				Password:          password,
				IP:                someValidIP(),
				UserAgent:         someValidUserAgent(),
				Timezone:          time.UTC,
				OS:                identity.OperatingSystemAndroid,
				DeviceFingerprint: someValidDeviceFingerprint(),
			}
			creds, err := svc.Login(ctx, loginCmd)
			if err != nil {
				t.Fatalf("Login(%v) error = %v", loginCmd, err)
			}
			refreshCmd := identity.RefreshCredentialsCommand{
				RefreshToken: creds.RefreshToken,
			}
			_, err = svc.RefreshCredentials(ctx, refreshCmd)
			if err != nil {
				t.Errorf("RefreshCredentials(%v) error = %v, wantErr nil", refreshCmd, err)
			}
		})
	})

	t.Run("ValidateCredentials", func(t *testing.T) {
		ctx := context.Background()
		t.Run("Fails if authorization header is empty", func(t *testing.T) {
			query := identity.ValidateCredentialsQuery{
				AuthorizationHeader: "",
			}
			err := svc.ValidateCredentials(ctx, query)
			if !errors.Is(err, identity.ErrAuthorizationHeaderCannotBeEmpty) {
				t.Errorf("ValidateCredentials(%v) error = %v, wantErr %v", query,
					err, identity.ErrAuthorizationHeaderCannotBeEmpty)
			}
		})
		t.Run("Fails if authorization header is invalid", func(t *testing.T) {
			query := identity.ValidateCredentialsQuery{
				AuthorizationHeader: "invalid",
			}
			err := svc.ValidateCredentials(ctx, query)
			if !errors.Is(err, identity.ErrInvalidAccessToken) {
				t.Errorf("ValidateCredentials(%v) error = %v, wantErr %v", query,
					err, identity.ErrInvalidAccessToken)
			}
		})
		// TODO: Fails if access token is expired
		t.Run("Success", func(t *testing.T) {
			email := someValidEmail()
			password := someValidPassword()
			signupUser(t, svc, email, password)
			loginCmd := identity.LoginCommand{
				Email:             email,
				Password:          password,
				IP:                someValidIP(),
				UserAgent:         someValidUserAgent(),
				Timezone:          time.UTC,
				OS:                identity.OperatingSystemAndroid,
				DeviceFingerprint: someValidDeviceFingerprint(),
			}
			creds, err := svc.Login(ctx, loginCmd)
			if err != nil {
				t.Fatalf("Login(%v) error = %v", loginCmd, err)
			}
			query := identity.ValidateCredentialsQuery{
				AuthorizationHeader: fmt.Sprintf("Bearer %s", creds.AccessToken),
			}
			err = svc.ValidateCredentials(ctx, query)
			if err != nil {
				t.Errorf("ValidateCredentials(%v) error = %v, wantErr nil", query, err)
			}
		})
	})

	t.Run("VerifyEmail", func(t *testing.T) {
		ctx := context.Background()
		t.Run("Fails if verification id is empty", func(t *testing.T) {
			cmd := identity.VerifyEmailCommand{
				VerificationID: "",
			}
			err := svc.VerifyEmail(ctx, cmd)
			if !errors.Is(err, identity.ErrEmailVerificationIDCannotBeEmpty) {
				t.Errorf("VerifyEmail(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrEmailVerificationIDCannotBeEmpty)
			}
		})
		t.Run("Fails if verification id is invalid", func(t *testing.T) {
			cmd := identity.VerifyEmailCommand{
				VerificationID: "invalid",
			}
			err := svc.VerifyEmail(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidEmailVerificationID) {
				t.Errorf("VerifyEmail(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidEmailVerificationID)
			}
		})
		//TODO: Fails if email verification is expired
		// TODO: Fails if email is already verified
		// TODO: figure out how to test success case
		t.Run("success", func(t *testing.T) {
			email := someValidEmail()
			password := someValidPassword()
			signupUser(t, svc, email, password)
			externalActions.VerifyEmailExternally(t, email)
		})
		// We need to create fixtures to test success scenario
		// TODO: Figure out a way to test timeout scenario when token gets expired
	})

	t.Run("RequestResetPassword", func(t *testing.T) {
		ctx := context.Background()
		t.Run("Fails if email is empty", func(t *testing.T) {
			cmd := identity.RequestResetPasswordCommand{
				Email: "",
			}
			err := svc.RequestResetPassword(ctx, cmd)
			if !errors.Is(err, identity.ErrEmailCannotBeEmpty) {
				t.Errorf("RequestResetPassword(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrEmailCannotBeEmpty)
			}
		})
		t.Run("Fails if email is invalid", func(t *testing.T) {
			cmd := identity.RequestResetPasswordCommand{
				Email: "invalid",
			}
			err := svc.RequestResetPassword(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidEmail) {
				t.Errorf("RequestResetPassword(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidEmail)
			}
		})
		t.Run("Fails if email is not registered", func(t *testing.T) {
			cmd := identity.RequestResetPasswordCommand{
				Email: someValidEmail(),
			}
			err := svc.RequestResetPassword(ctx, cmd)
			if !errors.Is(err, identity.ErrEmailNotRegistered) {
				t.Errorf("RequestResetPassword(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrEmailNotRegistered)
			}
		})
		t.Run("Success", func(t *testing.T) {
			email := someValidEmail()
			password := someValidPassword()
			signupUser(t, svc, email, password)
			cmd := identity.RequestResetPasswordCommand{
				Email: email,
			}
			err := svc.RequestResetPassword(ctx, cmd)
			if err != nil {
				t.Errorf("RequestResetPassword(%v) error = %v, wantErr nil", cmd, err)
			}
		})
	})

	t.Run("ValidateResetPasswordToken", func(t *testing.T) {
		ctx := context.Background()
		t.Run("Fails if token is empty", func(t *testing.T) {
			query := identity.ValidateResetPasswordTokenQuery{
				Token: "",
			}
			err := svc.ValidateResetPasswordToken(ctx, query)
			if !errors.Is(err, identity.ErrResetPasswordIDCannotBeEmpty) {
				t.Errorf("ValidateResetPasswordToken(%v) error = %v, wantErr %v", query,
					err, identity.ErrResetPasswordIDCannotBeEmpty)
			}
		})
		t.Run("Fails if token is invalid", func(t *testing.T) {
			query := identity.ValidateResetPasswordTokenQuery{
				Token: "invalid",
			}
			err := svc.ValidateResetPasswordToken(ctx, query)
			if !errors.Is(err, identity.ErrInvalidResetPasswordToken) {
				t.Errorf("ValidateResetPasswordToken(%v) error = %v, wantErr %v", query,
					err, identity.ErrInvalidResetPasswordToken)
			}
		})
		// TODO: Fails if token is expired
		// TODO: Success
		// These two will require access to email verification token
	})

	t.Run("ResetPassword", func(t *testing.T) {
		ctx := context.Background()
		t.Run("Fails if token is empty", func(t *testing.T) {
			cmd := identity.ResetPasswordCommand{
				Token:    "",
				Password: someValidPassword(),
			}
			err := svc.ResetPassword(ctx, cmd)
			if !errors.Is(err, identity.ErrResetPasswordIDCannotBeEmpty) {
				t.Errorf("ResetPassword(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrResetPasswordIDCannotBeEmpty)
			}
		})
		t.Run("Fails if token is invalid", func(t *testing.T) {
			cmd := identity.ResetPasswordCommand{
				Token:    someInvalidResetPasswordToken(),
				Password: someValidPassword(),
			}
			err := svc.ResetPassword(ctx, cmd)
			if !errors.Is(err, identity.ErrInvalidResetPasswordToken) {
				t.Errorf("ResetPassword(%v) error = %v, wantErr %v", cmd,
					err, identity.ErrInvalidResetPasswordToken)
			}
		})
		// TODO: Fails if token is expired
		// TODO: Fails if password is empty as it requires access to valid password reset token
		// TODO: Fails if password is invalid as it requires access to valid password reset token
		// TODO: Success case requires access to email verification token
	})

	t.Run("UserSessions", func(t *testing.T) {
		ctx := context.Background()
		t.Run("Fails if authorization header is empty", func(t *testing.T) {
			query := identity.UserSessionsQuery{
				AuthorizationHeader: "",
			}
			_, err := svc.UserSessions(ctx, query)
			if !errors.Is(err, identity.ErrAuthorizationHeaderCannotBeEmpty) {
				t.Errorf("UserSessions(%v) error = %v, wantErr %v", query,
					err, identity.ErrAuthorizationHeaderCannotBeEmpty)
			}
		})
		t.Run("Fails if authorization header is invalid", func(t *testing.T) {
			query := identity.UserSessionsQuery{
				AuthorizationHeader: "invalid",
			}
			_, err := svc.UserSessions(ctx, query)
			if !errors.Is(err, identity.ErrInvalidAccessToken) {
				t.Errorf("UserSessions(%v) error = %v, wantErr %v", query,
					err, identity.ErrInvalidAccessToken)
			}
		})
		t.Run("Success", func(t *testing.T) {
			email := someValidEmail()
			password := someValidPassword()
			signupUser(t, svc, email, password)
			loginCmd := identity.LoginCommand{
				Email:             email,
				Password:          password,
				IP:                someValidIP(),
				UserAgent:         someValidUserAgent(),
				Timezone:          time.UTC,
				OS:                identity.OperatingSystemAndroid,
				DeviceFingerprint: someValidDeviceFingerprint(),
			}
			creds, err := svc.Login(ctx, loginCmd)
			if err != nil {
				t.Fatalf("Login(%v) error = %v", loginCmd, err)
			}
			query := identity.UserSessionsQuery{
				AuthorizationHeader: fmt.Sprintf("Bearer %s", creds.AccessToken),
			}
			_, err = svc.UserSessions(ctx, query)
			if err != nil {
				t.Errorf("UserSessions(%v) error = %v, wantErr nil", query, err)
			}
		})
	})
}

func someValidDeviceFingerprint() string {
	return uuid.New().String()
}

func someValidEmail() string {
	return fmt.Sprintf("%s@balancewise.app", uuid.New().String())
}

func someValidPassword() string {
	return "123@Balancewise"
}

func someValidIP() string {
	return "1.1.1.1"
}

func someValidUserAgent() string {
	return "Mozilla/5.0"
}

func someValidBrand() string {
	return "Samsung"
}

func someValidIPCountryISO() string {
	return "US"
}

func someValidOS() identity.OperatingSystem {
	return identity.OperatingSystemAndroid
}

func someValidTimezone() *time.Location {
	return time.UTC
}

func signupUser(t *testing.T, svc identity.Service, email, password string) {
	t.Helper()
	ctx := context.Background()
	cmd := identity.CreateUserCommand{
		Email:             email,
		Password:          password,
		IP:                someValidIP(),
		IPCountryISO:      someValidIPCountryISO(),
		UserAgent:         someValidUserAgent(),
		Timezone:          someValidTimezone(),
		OS:                someValidOS(),
		DeviceFingerprint: someValidDeviceFingerprint(),
		Brand:             someValidBrand(),
	}
	_, _ = svc.CreateUser(ctx, cmd)
	return
}

func someInvalidRefreshToken(t *testing.T) string {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	tm := token.NewManager(privateKey)
	refreshToken, err := tm.NewRefreshToken(uuid.New().String())
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	return refreshToken.TokenString
}

func someInvalidResetPasswordToken() string {
	return uuid.New().String()
}

type ExternalActions interface {
	VerifyEmailExternally(t *testing.T, email string)
	ResetPasswordExternally(t *testing.T, email, password string)
}
