package postgres

import (
	"fmt"
	"profile/internal/model"
	"profile/pkg/postgres"

	sq "github.com/Masterminds/squirrel"

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

func (s *Storage) GetProfileFromSubjectID(subjID string) (*model.Profile, error) {
	sql, args, err := sq.
		Select(AllLabelsSelect).
		From(ProfileTable).
		Where(sq.Eq{SubjectIDLabel: subjID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select profile by subject id sql: %w", err)
	}

	var entity ProfileEntity
	err = s.db.Get(&entity, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("select profile by subject id: %w", err)
	}

	return entity.ToModel(), nil
}
