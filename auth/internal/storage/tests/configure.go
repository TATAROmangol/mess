package storage_test

import (
	"auth/internal/storage"
	"auth/pkg/sqlxdb"
	"fmt"

	sq "github.com/doug-martin/goqu/v9"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

var (
	CFG = sqlxdb.Config{
		Host:     "localhost",
		Port:     5430,
		User:     "test",
		Password: "test",
		DBName:   "test",
		SSLMode:  "disable",
	}

	MigrationsPath = "file://../../../migrations"
)

func setupTestDB() (*sqlx.DB, error) {
	db, err := sqlxdb.NewDB(CFG)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	truncateTables(db, storage.SubjectTableField)

	migr, err := sqlxdb.NewMigrator(CFG, MigrationsPath)
	if err != nil {
		return nil, fmt.Errorf("new migrator: %w", err)
	}
	defer migr.Close()

	err = migr.Up()
	if err != nil {
		return nil, fmt.Errorf("migrator up: %w", err)
	}

	return db, nil
}

func truncateTables(db *sqlx.DB, table string) error {
	query := sq.Truncate(table).Cascade()

	sqlStr, args, err := query.ToSQL()
	if err != nil {
		return fmt.Errorf("to sql: %w", err)
	}

	_, err = db.Exec(sqlStr, args...)
	if err != nil {
		return fmt.Errorf("truncate table %s: %w", table, err)
	}

	return nil
}
