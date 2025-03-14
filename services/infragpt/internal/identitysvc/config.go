package identitysvc

import (
	"github.com/priyanshujain/infragpt/services/infragpt/identity"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/domain"
)

type Config struct {
	domain.SessionRepository
	domain.UserRepository
	EmailService
	TokenManager
}

func (c Config) New() identity.Service {
	if c.SessionRepository == nil {
		panic("SessionRepository is nil")
	}
	if c.UserRepository == nil {
		panic("UserRepository is nil")
	}
	if c.EmailService == nil {
		panic("EmailService is nil")
	}

	s := &service{
		sessionRepo:  c.SessionRepository,
		userRepo:     c.UserRepository,
		emailSVC:     c.EmailService,
		tokenManager: c.TokenManager,
	}
	return s
}
