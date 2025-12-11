wpackage postgres

import (
	"github.com/73ai/infragpt/services/backend/internal/generic/postgresconfig"
	_ "github.com/lib/pq"
)

type Config struct {
	postgresconfig.Config
}

func (c Config) New() (*BackendDB, error) {
	db, err := c.Init()
	if err != nil {
		return nil, err
	}

	return &BackendDB{
		db:      db,
		Querier: New(db),
	}, nil
}
