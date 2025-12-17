package postgres

import (
	"database/sql"
	"fmt"
	"profile/pkg/postgres"
)

type Storage struct {
	db *sql.DB
}

func New(cfg postgres.Config) (*Storage, error) {
	db, err := postgres.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	return &Storage{db: db}, nil
}
