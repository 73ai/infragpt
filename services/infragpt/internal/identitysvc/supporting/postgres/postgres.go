package postgres

import (
	"crypto/rsa"
	"database/sql"
)

type IdentityDB struct {
	db         *sql.DB
	querier    Querier
	queries    *Queries
	privateKey *rsa.PrivateKey
}
