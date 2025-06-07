package identitysvc

import (
	"database/sql"

	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/supporting/postgres"
)

type Config struct {
	Clerk ClerkConfig `mapstructure:"clerk"`
}

type ClerkConfig struct {
	WebhookSecret  string `mapstructure:"webhook_secret"`
	PublishableKey string `mapstructure:"publishable_key"`
}

func (c Config) New(db *sql.DB) infragpt.IdentityService {
	queries := postgres.New(db)
	userRepo := postgres.NewUserRepository(queries)
	organizationRepo := postgres.NewOrganizationRepository(queries)
	memberRepo := postgres.NewMemberRepository(queries)

	return New(userRepo, organizationRepo, memberRepo)
}
