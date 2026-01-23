package storage

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/TATAROmangol/mess/chat/internal/model"
	"github.com/jmoiron/sqlx"
)

var (
	deletedATIsNullLastReadOutboxFilter = fmt.Sprintf("%v %v", LastReadOutboxDeletedAtLabel, IsNullLabel)
)

func (s *Storage) doAndReturnLastReadOutbox(ctx context.Context, query string, args []interface{}) (*model.LastReadOutbox, error) {
	var entity LastReadOutboxEntity
	err := sqlx.GetContext(ctx, s.exec, &entity, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return entity.ToModel(), nil
}

func (s *Storage) doAndReturnLastReadOutboxes(ctx context.Context, query string, args []interface{}) ([]*model.LastReadOutbox, error) {
	var entities []*LastReadOutboxEntity
	err := sqlx.SelectContext(ctx, s.exec, &entities, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return LastReadOutboxEntityToModels(entities), nil
}

func (s *Storage) AddLastReadOutbox(ctx context.Context, recipientID string, subjectID string, chatID int, messageID int) (*model.LastReadOutbox, error) {
	query, args, err := sq.
		Insert(LastReadOutboxTable).
		Columns(
			LastReadOutboxRecipientIDLabel,
			LastReadOutboxSubjectIDLabel,
			LastReadOutboxChatIDLabel,
			LastReadOutboxMessageIDLabel,
		).
		Values(recipientID, subjectID, chatID, messageID).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnLastReadOutbox(ctx, query, args)
}

func (s *Storage) GetLastReadOutbox(ctx context.Context, limit int) ([]*model.LastReadOutbox, error) {
	query, args, err := sq.
		Select(AllLabelsSelect).
		From(LastReadOutboxTable).
		Where(sq.Expr(deletedATIsNullLastReadOutboxFilter)).
		Limit(uint64(limit)).
		Suffix(SkipLocked).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnLastReadOutboxes(ctx, query, args)
}

func (s *Storage) DeleteLastReadOutbox(ctx context.Context, ids []int) ([]*model.LastReadOutbox, error) {
	if len(ids) == 0 {
		return []*model.LastReadOutbox{}, nil
	}

	query, args, err := sq.
		Update(LastReadOutboxTable).
		Set(LastReadOutboxDeletedAtLabel, time.Now().UTC()).
		Where(sq.Eq{LastReadOutboxIDLabel: ids}).
		Where(sq.Expr(deletedATIsNullLastReadOutboxFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnLastReadOutboxes(ctx, query, args)
}
