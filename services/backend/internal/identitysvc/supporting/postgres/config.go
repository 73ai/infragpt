package postgres

import (
	_ "github.com/lib/pq"
	"github.com/priyanshujain/infragpt/services/backend/internal/generic/postgresconfig"
)

type Config struct {
	postgresconfig.Config
}

func (c Config) New() (*IdentityDB, error) {
	db, err := c.Init()
	if err != nil {
		return nil, err
	}

	return &IdentityDB{
		db:      db,
		Querier: New(db),
	}, nil
}
