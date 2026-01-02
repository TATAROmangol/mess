package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/TATAROmangol/mess/profile/internal/model"
	"github.com/TATAROmangol/mess/shared/postgres"

	sq "github.com/Masterminds/squirrel"

	"github.com/jmoiron/sqlx"
)

var (
	deletedATIsNullProfileFilter = fmt.Sprintf("%v %v", ProfileDeletedAtLabel, IsNullLabel)
)

func (s *Storage) AddProfile(ctx context.Context, subjID string, alias string) (*model.Profile, error) {
	query, args, err := sq.
		Insert(ProfileTable).
		Columns(
			ProfileSubjectIDLabel,
			ProfileAliasLabel,
			ProfileVersionLabel,
			ProfileUpdatedAtLabel,
			ProfileCreatedAtLabel,
		).
		Values(subjID, alias, 1, time.Now().UTC(), time.Now().UTC(), nil).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var entity ProfileEntity
	err = sqlx.GetContext(ctx, s.exec, &entity, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return entity.ToModel(), nil
}

func (s *Storage) GetProfileFromSubjectID(ctx context.Context, subjID string) (*model.Profile, error) {
	query, args, err := sq.
		Select(AllLabelsSelect).
		From(ProfileTable).
		Where(sq.Eq{ProfileSubjectIDLabel: subjID}).
		Where(sq.Expr(deletedATIsNullProfileFilter)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var entity ProfileEntity
	err = sqlx.GetContext(ctx, s.exec, &entity, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return entity.ToModel(), nil
}

func (s *Storage) GetProfilesFromAlias(ctx context.Context, size int, asc bool, sortLabel Label, alias string) (string, []*model.Profile, error) {
	last := postgres.NewLast(ProfileSubjectIDLabel, nil)
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
		From(ProfileTable).
		Where(sq.Like{ProfileAliasLabel: alias + "%"}).
		Where(sq.Expr(deletedATIsNullProfileFilter))

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

func (s *Storage) UpdateProfileMetadata(ctx context.Context, subjectID string, prevVersion int, alias string) (*model.Profile, error) {
	query, args, err := sq.
		Update(ProfileTable).
		Set(ProfileAliasLabel, alias).
		Set(ProfileVersionLabel, prevVersion+1).
		Set(ProfileUpdatedAtLabel, time.Now().UTC()).
		Where(sq.Eq{ProfileSubjectIDLabel: subjectID}).
		Where(sq.Eq{ProfileVersionLabel: prevVersion}).
		Where(sq.Expr(deletedATIsNullProfileFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var entity ProfileEntity
	err = sqlx.GetContext(ctx, s.exec, &entity, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return entity.ToModel(), nil
}

func (s *Storage) UpdateAvatarKey(ctx context.Context, subjID string, avatarKey string) error {
	query, args, err := sq.
		Update(ProfileTable).
		Set(ProfileAvatarKeyLabel, avatarKey).
		Set(ProfileUpdatedAtLabel, time.Now().UTC()).
		Where(sq.Eq{ProfileSubjectIDLabel: subjID}).
		Where(sq.Expr(deletedATIsNullProfileFilter)).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = s.exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("db exec: %w", err)
	}

	return nil
}

func (s *Storage) DeleteProfile(ctx context.Context, subjID string) error {
	query, args, err := sq.
		Update(ProfileTable).
		Set(ProfileDeletedAtLabel, time.Now().UTC()).
		Set(ProfileUpdatedAtLabel, time.Now().UTC()).
		Where(sq.Eq{ProfileSubjectIDLabel: subjID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = s.exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("db exec: %w", err)
	}

	return nil
}

func (s *Storage) DeleteAvatarKey(ctx context.Context, subjID string) error {
	query, args, err := sq.
		Update(ProfileTable).
		Set(ProfileAvatarKeyLabel, nil).
		Set(ProfileUpdatedAtLabel, time.Now().UTC()).
		Where(sq.Eq{ProfileSubjectIDLabel: subjID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = s.exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("db exec: %w", err)
	}

	return nil
}
