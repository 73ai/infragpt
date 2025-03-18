package identitysvctest

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/supporting/postgres"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/supporting/postgres/postgrestest"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/supporting/token"
	"sync"
	"testing"

	"github.com/priyanshujain/infragpt/services/infragpt/identity"
	"github.com/priyanshujain/infragpt/services/infragpt/identity/identitytest"
)

func NewServiceWithExternalActions(t *testing.T) (identity.Service, identitytest.ExternalActions) {
	db := postgrestest.NewDB(t)
	// NIST deems RSA 2048 sufficient until 2030
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	ur := postgres.Config{}
	emailSVC := &emailService{
		emails:         sync.Map{},
		UserRepository: ur,
	}
	config := identitysvc.Config{
		UserRepository:    ur,
		SessionRepository: postgres.NewSessionRepository(db, privateKey),
		EmailService:      emailSVC,
		TokenManager:      token.NewManager(privateKey),
	}

	return config.New(), emailSVC
}
