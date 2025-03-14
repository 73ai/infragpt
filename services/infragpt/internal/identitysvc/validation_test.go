package identitysvc_test

import (
	"errors"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc"
	"testing"
	"time"

	"github.com/priyanshujain/infragpt/services/infragpt/identity"
)

func TestValidatePassword(t *testing.T) {
	t.Run("Fails if password is empty", func(t *testing.T) {
		err := identitysvc.ValidatePassword("")
		if !errors.Is(err, identity.ErrPasswordCannotBeEmpty) {
			t.Errorf("ValidatePassword('') want %v got %v", identity.ErrPasswordCannotBeEmpty, err)
		}
	})
	t.Run("Fails if password is less than 8 characters", func(t *testing.T) {
		err := identitysvc.ValidatePassword("pass")
		if !errors.Is(err, identity.ErrInvalidPassword) {
			t.Errorf("ValidatePassword('pass') want %v got %v", identity.ErrInvalidPassword, err)
		}
	})
	t.Run("Fails if password does not contain uppercase letter", func(t *testing.T) {
		err := identitysvc.ValidatePassword("password")
		if !errors.Is(err, identity.ErrInvalidPassword) {
			t.Errorf("ValidatePassword('password') want %v got %v", identity.ErrInvalidPassword, err)
		}
	})
	t.Run("Fails if password does not contain lowercase letter", func(t *testing.T) {
		err := identitysvc.ValidatePassword("PASSWORD")
		if !errors.Is(err, identity.ErrInvalidPassword) {
			t.Errorf("ValidatePassword('PASSWORD') want %v got %v", identity.ErrInvalidPassword, err)
		}
	})
	t.Run("Fails if password does not contain digit", func(t *testing.T) {
		err := identitysvc.ValidatePassword("Password")
		if !errors.Is(err, identity.ErrInvalidPassword) {
			t.Errorf("ValidatePassword('Password') want %v got %v", identity.ErrInvalidPassword, err)
		}
	})
	t.Run("Fails if password does not contain special character", func(t *testing.T) {
		err := identitysvc.ValidatePassword("Password1")
		if !errors.Is(err, identity.ErrInvalidPassword) {
			t.Errorf("ValidatePassword('Password1') want %v got %v", identity.ErrInvalidPassword, err)
		}
	})
	t.Run("Succeeds if password is valid", func(t *testing.T) {
		err := identitysvc.ValidatePassword("Password1!")
		if err != nil {
			t.Errorf("ValidatePassword('Password1!') want nil got %v", err)
		}
	})
}

func TestValidateEmail(t *testing.T) {
	t.Run("Fails if email is empty", func(t *testing.T) {
		err := identitysvc.ValidateEmail("")
		if !errors.Is(err, identity.ErrEmailCannotBeEmpty) {
			t.Errorf("ValidateEmail('') want %v got %v", identity.ErrEmailCannotBeEmpty, err)
		}
	})
	t.Run("Fails if email is invalid", func(t *testing.T) {
		err := identitysvc.ValidateEmail("email")
		if !errors.Is(err, identity.ErrInvalidEmail) {
			t.Errorf("ValidateEmail('email') want %v got %v", identity.ErrInvalidEmail, err)
		}
	})
	t.Run("Succeeds if email is valid", func(t *testing.T) {
		err := identitysvc.ValidateEmail("team@balancewise.app")
		if err != nil {
			t.Errorf("ValidateEmail('team@balancewise.app') want nil got %v", err)
		}
	})
}

func TestValidateIP(t *testing.T) {
	t.Run("Fails if IP is empty", func(t *testing.T) {
		err := identitysvc.ValidateIP("")
		if !errors.Is(err, identity.ErrIPCannotBeEmpty) {
			t.Errorf("ValidateIP('') want %v got %v", identity.ErrIPCannotBeEmpty, err)
		}
	})
	t.Run("Fails if IP is invalid", func(t *testing.T) {
		err := identitysvc.ValidateIP("1234")
		if !errors.Is(err, identity.ErrInvalidIP) {
			t.Errorf("ValidateIP('1234') want %v got %v", identity.ErrInvalidIP, err)
		}
	})
	t.Run("Succeeds if IP is valid", func(t *testing.T) {
		err := identitysvc.ValidateIP("1.1.1.1")
		if err != nil {
			t.Errorf("ValidateIP('1.1.1.1') want nil got %v", err)
		}
	})
}

func TestValidateUserAgent(t *testing.T) {
	t.Run("Fails if UserAgent is empty", func(t *testing.T) {
		err := identitysvc.ValidateUserAgent("")
		if !errors.Is(err, identity.ErrUserAgentCannotBeEmpty) {
			t.Errorf("ValidateUserAgent('') want %v got %v", identity.ErrUserAgentCannotBeEmpty, err)
		}
	})
	t.Run("Fails if UserAgent is invalid", func(t *testing.T) {
		err := identitysvc.ValidateUserAgent("invalid")
		if !errors.Is(err, identity.ErrInvalidUserAgent) {
			t.Errorf("ValidateUserAgent('invalid') want %v got %v", identity.ErrInvalidUserAgent, err)
		}
	})
	t.Run("Succeeds if UserAgent is valid", func(t *testing.T) {
		err := identitysvc.ValidateUserAgent("Mozilla/5.0")
		if err != nil {
			t.Errorf("ValidateUserAgent('Mozilla/5.0') want nil got %v", err)
		}
	})
}

func TestValidateOS(t *testing.T) {
	t.Run("Fails if OS is invalid", func(t *testing.T) {
		err := identitysvc.ValidateOS("invalid")
		if !errors.Is(err, identity.ErrInvalidOS) {
			t.Errorf("ValidateOS('invalid') want %v got %v", identity.ErrInvalidOS, err)
		}
	})
	t.Run("Succeeds if OS is valid", func(t *testing.T) {
		err := identitysvc.ValidateOS("windows")
		if err != nil {
			t.Errorf("ValidateOS('windows') want nil got %v", err)
		}
	})
}

func TestValidateDeviceFingerprint(t *testing.T) {
	t.Run("Fails if DeviceFingerprint is empty", func(t *testing.T) {
		err := identitysvc.ValidateDeviceFingerprint("")
		if !errors.Is(err, identity.ErrDeviceFingerprintCannotBeEmpty) {
			t.Errorf("ValidateDeviceFingerprint('') want %v got %v", identity.ErrDeviceFingerprintCannotBeEmpty, err)
		}
	})
	t.Run("Succeeds if DeviceFingerprint is valid", func(t *testing.T) {
		err := identitysvc.ValidateDeviceFingerprint("fingerprint")
		if err != nil {
			t.Errorf("ValidateDeviceFingerprint('fingerprint') want nil got %v", err)
		}
	})
}

func TestValidateTimezone(t *testing.T) {
	t.Run("Fails if Timezone is empty", func(t *testing.T) {
		err := identitysvc.ValidateTimezone(&time.Location{})
		if !errors.Is(err, identity.ErrTimezoneCannotBeEmpty) {
			t.Errorf("ValidateTimezone('') want %v got %v", identity.ErrTimezoneCannotBeEmpty, err)
		}
	})
	t.Run("Succeeds if Timezone is valid", func(t *testing.T) {
		err := identitysvc.ValidateTimezone(time.UTC)
		if err != nil {
			t.Errorf("ValidateTimezone('UTC') want nil got %v", err)
		}
	})
}
