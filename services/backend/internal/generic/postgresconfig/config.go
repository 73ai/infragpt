package postgresconfig

import (
	"database/sql"
	"fmt"
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

func (c Config) Init() (*sql.DB, error) {
	db, err := sql.Open("postgres", c.connStr())
	if err != nil {
		return nil, err
	}

	return db, nil
}
