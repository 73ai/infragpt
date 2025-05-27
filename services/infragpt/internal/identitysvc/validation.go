package identitysvc

import (
	"regexp"
	"time"

	"github.com/priyanshujain/infragpt/services/infragpt/identity"
)

func ValidateEmail(email string) error {
	if email == "" {
		return identity.ErrEmailCannotBeEmpty
	}
	isValid := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email)
	if !isValid {
		return identity.ErrInvalidEmail
	}

	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return identity.ErrPasswordCannotBeEmpty
	}
	if len(password) < 8 {
		return identity.ErrInvalidPassword
	}

	isValid := regexp.MustCompile(`[A-Z]`).MatchString(password) &&
		regexp.MustCompile(`[a-z]`).MatchString(password) &&
		regexp.MustCompile(`\d`).MatchString(password) &&
		regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	if !isValid {
		return identity.ErrInvalidPassword
	}

	return nil
}

func ValidateIP(ip string) error {
	if ip == "" {
		return identity.ErrIPCannotBeEmpty
	}
	isValid := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`).MatchString(ip)
	if !isValid {
		return identity.ErrInvalidIP
	}

	return nil
}

func ValidateUserAgent(userAgent string) error {
	if userAgent == "" {
		return identity.ErrUserAgentCannotBeEmpty
	}
	isValid := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+/[a-zA-Z0-9.-]+$`).MatchString(userAgent)
	if !isValid {
		return identity.ErrInvalidUserAgent
	}

	return nil
}

func ValidateOS(os identity.OperatingSystem) error {
	if os == "" {
		return identity.ErrInvalidOS
	}

	// os should be of type identity.OperatingSystem
	switch os {
	case identity.OperatingSystemAndroid, identity.OperatingSystemIOS,
		identity.OperatingSystemWindows, identity.OperatingSystemMacOS,
		identity.OperatingSystemLinux:
		return nil
	default:
		return identity.ErrInvalidOS
	}
}

func ValidateDeviceFingerprint(deviceFingerprint string) error {
	if deviceFingerprint == "" {
		return identity.ErrDeviceFingerprintCannotBeEmpty
	}
	return nil
}

func ValidateTimezone(timezone *time.Location) error {
	if timezone.String() == "" {
		return identity.ErrTimezoneCannotBeEmpty
	}
	return nil
}
