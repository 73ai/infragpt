package postgres

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DBName   string `mapstructure:"db_name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

func (c Config) connStr() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Password, c.DBName)
}

func (c Config) New(ctx context.Context) (*InfraGPTDB, error) {
	db, err := sql.Open("pgx", c.connStr())
	if err != nil {
		return nil, err
	}

	return &InfraGPTDB{
		db:      db,
		Querier: New(db),
	}, nil
}
