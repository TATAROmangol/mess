package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/TATAROmangol/mess/profile/internal/model"
	"github.com/jmoiron/sqlx"
)

var (
	deletedATIsNullAvatarKeyFilter = fmt.Sprintf("%v %v", AvatarKeyOutboxDeletedAtLabel, IsNullLabel)
)

func (s *Storage) GetKeys(ctx context.Context, limit int) ([]*model.AvatarKeyOutbox, error) {
	query, args, err := sq.
		Select(AllLabelsSelect).
		From(AvatarKeyOutboxTable).
		Where(sq.Expr(deletedATIsNullAvatarKeyFilter)).
		Limit(uint64(limit)).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build select not deleted key sql: %w", err)
	}

	var entities []*AvatarKeyOutboxEntity
	err = sqlx.SelectContext(ctx, s.exec, &entities, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("db get: %w", err)
	}

	return AvatarKeyOutboxEntitiesToModels(entities), nil
}

func (s *Storage) AddKey(ctx context.Context, subjectID string, key string) (*model.AvatarKeyOutbox, error) {
	query, args, err := sq.
		Insert(AvatarKeyOutboxTable).
		Columns(
			AvatarKeyOutboxKeyLabel,
			AvatarKeyOutboxSubjectIDLabel,
			AvatarKeyOutboxCreatedAtLabel,
			AvatarKeyOutboxDeletedAtLabel,
		).
		Values(key, subjectID, time.Now().UTC(), nil).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build insert profile sql: %w", err)
	}

	var entity AvatarKeyOutboxEntity
	err = sqlx.GetContext(ctx, s.exec, &entity, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return entity.ToModel(), nil
}

func (s *Storage) DeleteKeys(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	query, args, err := sq.
		Update(AvatarKeyOutboxTable).
		Set(AvatarKeyOutboxDeletedAtLabel, time.Now().UTC()).
		Where(sq.Eq{AvatarKeyOutboxKeyLabel: keys}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("build update avatar keys sql: %w", err)
	}

	_, err = s.exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("db exec: %w", err)
	}

	return nil
}
