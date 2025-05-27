package postgres

import (
	"crypto/rsa"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/generic/postgresconfig"
)

type Config struct {
	postgresconfig.Config
	PrivateKey *rsa.PrivateKey
}

func (c Config) New() (*IdentityDB, error) {
	db, err := c.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to init postgres connection: %w", err)
	}

	return &IdentityDB{
		db:         db,
		querier:    New(db),
		queries:    New(db),
		privateKey: c.PrivateKey,
	}, nil
}
