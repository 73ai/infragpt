package postgres

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/generic/postgresconfig"
)

type Config struct {
	postgresconfig.Config
}

func (c Config) New() (*InfraGPTDB, error) {
	db, err := c.Init()
	if err != nil {
		return nil, err
	}

	return &InfraGPTDB{
		db:      db,
		Querier: New(db),
	}, nil
}
