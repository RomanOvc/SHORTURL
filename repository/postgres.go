package repository

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

type PsqlConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Dbname   string
	Sslmode  string
}

func InitPostgresDb(cfg PsqlConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s  dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Dbname, cfg.Password, cfg.Sslmode))
	if err != nil {
		return nil, errors.Wrap(err, "repository/postgres  InitPostgresDb() not found parametrs database")
	}

	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "repository/postgres  InitPostgresDb() not include to database")
	}

	return db, nil
}
