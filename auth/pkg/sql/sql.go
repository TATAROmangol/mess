package sql

import (
	"database/sql"
	"fmt"
)

const (
	PostgresDriver = "postgres"
)

type PostgresConfig struct {
	Host     string `json:"pg_host"`
	Port     int    `json:"pg_port"`
	User     string `json:"pg_user"`
	Password string `json:"pg_password"`
	DBName   string `json:"pg_dbname"`
	SSLMode  string `json:"pg_sslmode"`
}

func NewPostgresDB(cfg PostgresConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open(PostgresDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return db, nil
}
