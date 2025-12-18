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

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) AddProfile(profile *model.Profile) error {
	query, args, err := sq.
		Insert(ProfileTable).
		Columns(SubjectIDLabel, AliasLabel, AvatarURLLabel, VersionLabel, UpdatedAtLabel, CreatedAtLabel).
		Values(profile.SubjectID, profile.Alias, profile.AvatarURL, profile.Version, profile.UpdatedAt, profile.CreatedAt).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert profile sql: %w", err)
	}

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("insert profile: %w", err)
	}

	return nil
}

func (s *Storage) GetProfileFromSubjectID(subjID string) (*model.Profile, error) {
	query, args, err := sq.
		Select(AllLabelsSelect).
		From(ProfileTable).
		Where(sq.Eq{SubjectIDLabel: subjID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select profile by subject id sql: %w", err)
	}

	var entity ProfileEntity
	err = s.db.Get(&entity, query, args...)

	if err != nil {
		return nil, fmt.Errorf("select profile by subject id: %w", err)
	}

	return entity.ToModel(), nil
}

func (s *Storage) UpdateProfile(profile *model.Profile) error {
	query, args, err := sq.
		Update(ProfileTable).
		Set(AliasLabel, profile.Alias).
		Set(AvatarURLLabel, profile.AvatarURL).
		Set(VersionLabel, profile.Version).
		Set(UpdatedAtLabel, profile.UpdatedAt).
		Where(sq.Eq{SubjectIDLabel: profile.SubjectID}).
		Where(sq.Eq{VersionLabel: profile.Version - 1}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update profile sql: %w", err)
	}

	res, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows updated, possible version conflict")
	}

	return nil
}
