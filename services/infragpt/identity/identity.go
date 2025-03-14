package identity

import (
	"context"
	"errors"
	"time"
)

var (
	ErrEmailCannotBeEmpty               = errors.New("email_cannot_be_empty")
	ErrInvalidEmail                     = errors.New("invalid_email")
	ErrPasswordCannotBeEmpty            = errors.New("password_cannot_be_empty")
	ErrInvalidPassword                  = errors.New("invalid_password")
	ErrIPCannotBeEmpty                  = errors.New("ip_cannot_be_empty")
	ErrInvalidIP                        = errors.New("invalid_ip")
	ErrUserAgentCannotBeEmpty           = errors.New("user_agent_cannot_be_empty")
	ErrInvalidUserAgent                 = errors.New("invalid_user_agent")
	ErrInvalidOS                        = errors.New("invalid_os")
	ErrDeviceFingerprintCannotBeEmpty   = errors.New("device_fingerprint_cannot_be_empty")
	ErrTimezoneCannotBeEmpty            = errors.New("timezone_cannot_be_empty")
	ErrInvalidEmailOrPassword           = errors.New("invalid_email_or_password")
	ErrInvalidRefreshToken              = errors.New("invalid_refresh_token")
	ErrRefreshTokenCannotBeEmpty        = errors.New("refresh_token_cannot_be_empty")
	ErrAccessTokenCannotBeEmpty         = errors.New("access_token_cannot_be_empty")
	ErrInvalidAccessToken               = errors.New("invalid_access_token")
	ErrAuthorizationHeaderCannotBeEmpty = errors.New("authorization_header_cannot_be_empty")
	ErrAccessTokenExpired               = errors.New("access_token_expired")
	ErrRefreshTokenExpired              = errors.New("refresh_token_expired")
	ErrInvalidEmailVerificationID       = errors.New("invalid_email_verification_id")
	ErrEmailVerificationIDCannotBeEmpty = errors.New("email_verification_id_cannot_be_empty")
	ErrorEmailVerificationExpired       = errors.New("email_verification_expired")
	ErrEmailAlreadyVerified             = errors.New("email_already_verified")
	ErrEmailNotRegistered               = errors.New("email_not_registered")
	ErrResetPasswordIDCannotBeEmpty     = errors.New("reset_password_id_cannot_be_empty")
	ErrInvalidResetPasswordToken        = errors.New("invalid_reset_password_token")
	ErrResetPasswordTokenExpired        = errors.New("reset_password_token_expired")
	ErrGoogleAuthCodeCannotBeEmpty      = errors.New("google_auth_code_cannot_be_empty")
	ErrGoogleAuthStateCannotBeEmpty     = errors.New("google_auth_state_cannot_be_empty")
)

type UserSession struct {
	UserID         string
	SessionID      string
	Device         UserDevice
	UserAgent      string
	IPAddress      string
	IPCountryISO   string
	LastActivityAt time.Time
	Timezone       *time.Location
	IsExpired      bool
}

type UserDevice struct {
	UserID            string
	DeviceFingerprint string
	Name              string
	OS                OperatingSystem
	Brand             string
}

type OperatingSystem string

var (
	OperatingSystemAndroid OperatingSystem = "android"
	OperatingSystemIOS     OperatingSystem = "ios"
	OperatingSystemWindows OperatingSystem = "windows"
	OperatingSystemMacOS   OperatingSystem = "macos"
	OperatingSystemLinux   OperatingSystem = "linux"
)

type Service interface {
	CreateUser(context.Context, CreateUserCommand) (Credentials, error)
	Login(context.Context, LoginCommand) (Credentials, error)
	RefreshCredentials(context.Context, RefreshCredentialsCommand) (Credentials, error)
	ValidateCredentials(context.Context, ValidateCredentialsQuery) error

	VerifyEmail(context.Context, VerifyEmailCommand) error

	RequestResetPassword(context.Context, RequestResetPasswordCommand) error
	ValidateResetPasswordToken(context.Context, ValidateResetPasswordTokenQuery) error
	ResetPassword(context.Context, ResetPasswordCommand) error

	UserSessions(context.Context, UserSessionsQuery) ([]UserSession, error)

	InitiateGoogleAuth(context.Context) (string, error)
	CompleteGoogleAuth(context.Context, string, string) (Credentials, error)
}

type CreateUserCommand struct {
	Email             string
	Password          string
	IP                string
	UserAgent         string
	IPCountryISO      string
	OS                OperatingSystem
	Brand             string
	DeviceFingerprint string
	Timezone          *time.Location
}

type LoginCommand struct {
	Email             string
	Password          string
	IP                string
	UserAgent         string
	IPCountryISO      string
	OS                OperatingSystem
	Brand             string
	DeviceFingerprint string
	Timezone          *time.Location
}

type Credentials struct {
	AccessToken  string
	RefreshToken string
}

type RefreshCredentialsCommand struct {
	RefreshToken string
}

type ValidateCredentialsQuery struct {
	AuthorizationHeader string
}

type VerifyEmailCommand struct {
	VerificationID string
}

type RequestResetPasswordCommand struct {
	Email string
}

type ValidateResetPasswordTokenQuery struct {
	Token string
}

type ResetPasswordCommand struct {
	Token    string
	Password string
}

type UserSessionsQuery struct {
	AuthorizationHeader string
}

// Future work:
// - Google login support
// - User email update
// - Two factor authentication support (email and totp)
// - Phone number support
// - GeoIP Reader location support (https://github.com/oschwald/geoip2-golang)
