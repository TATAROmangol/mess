package postgres

import (
	"fmt"
	"profile/pkg/postgres"

	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db *sqlx.DB
}

func New(cfg postgres.Config) (*Storage, error) {
	db, err := postgres.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	return &Storage{db: db}, nil
}
