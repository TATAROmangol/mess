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

func (s *Storage) doAndReturnProfile(ctx context.Context, query string, args []interface{}) (*model.Profile, error) {
	var entity ProfileEntity
	err := sqlx.GetContext(ctx, s.exec, &entity, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return entity.ToModel(), nil
}

func (s *Storage) doAndReturnProfiles(ctx context.Context, query string, args []interface{}) ([]*model.Profile, error) {
	var entities []*ProfileEntity
	err := sqlx.SelectContext(ctx, s.exec, &entities, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return ProfileEntitiesToModels(entities), nil
}

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
		Values(subjID, alias, 1, time.Now().UTC(), time.Now().UTC()).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnProfile(ctx, query, args)
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

	return s.doAndReturnProfile(ctx, query, args)
}

func (s *Storage) GetProfilesFromAlias(ctx context.Context, alias string, filter *postgres.PaginationFilter) ([]*model.Profile, error) {
	b := sq.
		Select(AllLabelsSelect).
		From(ProfileTable).
		Where(sq.Like{ProfileAliasLabel: alias + "%"}).
		Where(sq.Expr(deletedATIsNullProfileFilter))

	query, args, err := postgres.MakeQueryWithPagination(ctx, b, filter)
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnProfiles(ctx, query, args)
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

	return s.doAndReturnProfile(ctx, query, args)
}

func (s *Storage) UpdateAvatarKey(ctx context.Context, subjID string, avatarKey string) (*model.Profile, error) {
	query, args, err := sq.
		Update(ProfileTable).
		Set(ProfileAvatarKeyLabel, avatarKey).
		Set(ProfileUpdatedAtLabel, time.Now().UTC()).
		Where(sq.Eq{ProfileSubjectIDLabel: subjID}).
		Where(sq.Expr(deletedATIsNullProfileFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnProfile(ctx, query, args)
}

func (s *Storage) DeleteProfile(ctx context.Context, subjID string) (*model.Profile, error) {
	query, args, err := sq.
		Update(ProfileTable).
		Set(ProfileDeletedAtLabel, time.Now().UTC()).
		Set(ProfileUpdatedAtLabel, time.Now().UTC()).
		Where(sq.Eq{ProfileSubjectIDLabel: subjID}).
		Where(sq.Expr(deletedATIsNullProfileFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnProfile(ctx, query, args)
}

func (s *Storage) DeleteAvatarKey(ctx context.Context, subjID string) (*model.Profile, error) {
	query, args, err := sq.
		Update(ProfileTable).
		Set(ProfileAvatarKeyLabel, nil).
		Set(ProfileUpdatedAtLabel, time.Now().UTC()).
		Where(sq.Eq{ProfileSubjectIDLabel: subjID}).
		Where(sq.Expr(deletedATIsNullProfileFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnProfile(ctx, query, args)
}
