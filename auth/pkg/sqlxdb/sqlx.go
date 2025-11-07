package sqlxdb

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	PostgresDriver = "postgres"
)

type Config struct {
	Host     string `json:"pg_host"`
	Port     int    `json:"pg_port"`
	User     string `json:"pg_user"`
	Password string `json:"pg_password"`
	DBName   string `json:"pg_dbname"`
	SSLMode  string `json:"pg_sslmode"`
}

func NewDB(cfg Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sqlx.Connect(PostgresDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return db, nil
}
