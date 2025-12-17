package postgres

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Migrator struct {
	mig *migrate.Migrate
}

func NewMigrator(cfg Config, migrationsPath string) (*Migrator, error) {
	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	mig, err := migrate.New(
		migrationsPath,
		databaseURL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{
		mig: mig,
	}, nil
}

func (m *Migrator) Up() error {
	errUp := m.mig.Up()
	if errUp == nil || errUp == migrate.ErrNoChange {
		return nil
	}

	version, dirty, err := m.mig.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("version: %w", err)
	}

	if !dirty {
		return fmt.Errorf("database is not dirty, migrator up: %w", errUp)
	}

	if err := m.mig.Force(int(version - 1)); err != nil {
		return fmt.Errorf("force: %w", err)
	}

	return fmt.Errorf("migrator up: %w", errUp)
}

func (m *Migrator) Close() error {
	sErr, dbErr := m.mig.Close()
	if sErr != nil {
		return fmt.Errorf("migrator close, source: %w", sErr)
	}

	if dbErr != nil {
		return fmt.Errorf("migrator close, database: %w", dbErr)
	}

	return nil
}
