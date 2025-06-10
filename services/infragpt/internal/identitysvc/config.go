package identitysvc

import (
	"database/sql"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/supporting/clerk"

	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/supporting/postgres"
)

type Config struct {
	Database *sql.DB      `mapstructure:"-"`
	Clerk    clerk.Config `mapstructure:"clerk"`
}

func (c Config) New(db *sql.DB) *service {
	queries := postgres.New(db)
	userRepo := postgres.NewUserRepository(queries)
	organizationRepo := postgres.NewOrganizationRepository(queries)
	memberRepo := postgres.NewMemberRepository(queries)

	return &service{
		userRepo:         userRepo,
		organizationRepo: organizationRepo,
		memberRepo:       memberRepo,
		authService:      c.Clerk.NewAuthService(),
	}
}
