package postgres

import (
	"context"
	"fmt"
	"profile/internal/model"
	p "profile/internal/storage/profile"
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

func (s *Storage) AddProfile(ctx context.Context, prof *model.Profile) error {
	query, args, err := sq.
		Insert(p.ProfileTable).
		Columns(
			p.SubjectIDLabel,
			p.AliasLabel,
			p.AvatarURLLabel,
			p.VersionLabel,
			p.UpdatedAtLabel,
			p.CreatedAtLabel).
		Values(prof.SubjectID, prof.Alias, prof.AvatarURL, prof.Version, prof.UpdatedAt, prof.CreatedAt).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert profile sql: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert profile: %w", err)
	}

	return nil
}

func (s *Storage) GetProfileFromSubjectID(ctx context.Context, subjID string) (*model.Profile, error) {
	query, args, err := sq.
		Select(AllLabelsSelect).
		From(p.ProfileTable).
		Where(sq.Eq{p.SubjectIDLabel: subjID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select profile by subject id sql: %w", err)
	}

	var entity ProfileEntity
	err = s.db.GetContext(ctx, &entity, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return entity.ToModel(), nil
}

func (s *Storage) UpdateProfile(ctx context.Context, prof *model.Profile) error {
	query, args, err := sq.
		Update(p.ProfileTable).
		Set(p.AliasLabel, prof.Alias).
		Set(p.AvatarURLLabel, prof.AvatarURL).
		Set(p.VersionLabel, prof.Version).
		Set(p.UpdatedAtLabel, prof.UpdatedAt).
		Where(sq.Eq{p.SubjectIDLabel: prof.SubjectID}).
		Where(sq.Eq{p.VersionLabel: prof.Version - 1}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update profile sql: %w", err)
	}

	res, err := s.db.ExecContext(ctx, query, args...)
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

func (s *Storage) GetProfilesFromAlias(ctx context.Context, size int, asc bool, sortLabel p.Label, alias string) (string, []*model.Profile, error) {
	last := postgres.NewLast(p.SubjectIDLabel, nil)
	sort := postgres.NewSort(sortLabel, asc)

	pag := postgres.NewPagination(
		size,
		sort,
		last,
	)

	return s.getProfilesWithPagination(ctx, pag, alias)
}

func (s *Storage) GetProfilesFromAliasWithToken(ctx context.Context, token, alias string) (string, []*model.Profile, error) {
	pag, err := postgres.ParsePaginationToken(token)
	if err != nil {
		return "", nil, fmt.Errorf("parse pagination token: %w", err)
	}

	return s.getProfilesWithPagination(ctx, pag, alias)
}

func (s *Storage) getProfilesWithPagination(ctx context.Context, pag *postgres.Pagination, alias string) (string, []*model.Profile, error) {
	builder := sq.
		Select(AllLabelsSelect).
		From(p.ProfileTable).
		Where(sq.Like{p.AliasLabel: alias + "%"})

	newP, entities, err := postgres.MakeQueryWithPagination[*ProfileEntity](
		ctx,
		s.db,
		builder,
		pag,
	)
	if err != nil {
		return "", nil, fmt.Errorf("make query with pagination: %w", err)
	}

	return newP.Token(), ProfileEntitiesToModels(entities), nil
}
