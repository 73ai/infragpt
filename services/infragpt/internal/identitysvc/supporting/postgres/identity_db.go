package postgres

import (
	"database/sql"
)

type IdentityDB struct {
	db *sql.DB
	Querier
}