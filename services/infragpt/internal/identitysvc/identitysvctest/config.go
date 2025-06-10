package identitysvctest

import (
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/domaintest"
)

func NewConfig() Config {
	return Config{
		Config: identitysvc.Config{},
	}
}

type Config struct {
	identitysvc.Config
}

type fixture struct {
	svc infragpt.IdentityService
}

func (f *fixture) Service() infragpt.IdentityService {
	return f.svc
}

func (c Config) Fixture() *fixture {
	return &fixture{
		svc: c.New(),
	}
}

func (c Config) New() infragpt.IdentityService {
	userRepo := domaintest.NewUserRepository()
	organizationRepo := domaintest.NewOrganizationRepository()
	memberRepo := domaintest.NewMemberRepository()

	return identitysvc.New(userRepo, organizationRepo, memberRepo)
}
