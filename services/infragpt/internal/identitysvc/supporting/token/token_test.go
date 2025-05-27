package token_test

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/supporting/token"
	"testing"

	"github.com/google/uuid"
)

func TestNewManager(t *testing.T) {
	// NIST deems RSA 2048 sufficient until 2030
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	t.Run("NewRefreshToken", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			sid := someSessionID()
			m := token.NewManager(privateKey)
			_, err := m.NewRefreshToken(sid)
			if err != nil {
				t.Errorf("NewRefreshToken() error = %v, wantErr nil", err)
			}
		})
	})
	t.Run("NewAccessToken", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			sid := someSessionID()
			m := token.NewManager(privateKey)
			_, err := m.NewAccessToken(sid)
			if err != nil {
				t.Errorf("NewAccessToken() error = %v, wantErr nil", err)
			}
		})
	})
}

func someSessionID() string {
	return uuid.New().String()
}
