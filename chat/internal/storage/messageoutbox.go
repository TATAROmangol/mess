package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/TATAROmangol/mess/chat/internal/model"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

var (
	deletedATIsNullMessageOutboxFilter = fmt.Sprintf("%v %v", MessageOutboxDeletedAtLabel, IsNullLabel)
)

func (s *Storage) doAndReturnMessageOutbox(ctx context.Context, query string, args []interface{}) (*model.MessageOutbox, error) {
	var entity MessageOutboxEntity
	err := sqlx.GetContext(ctx, s.exec, &entity, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return entity.ToModel(), nil
}

func (s *Storage) doAndReturnMessageOutboxes(ctx context.Context, query string, args []interface{}) ([]*model.MessageOutbox, error) {
	var entities []*MessageOutboxEntity
	err := sqlx.SelectContext(ctx, s.exec, &entities, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return MessageOutboxEntitiesToModels(entities), nil
}

func (s *Storage) AddMessageOutbox(ctx context.Context, recipientID string, messageID int, operation model.Operation) (*model.MessageOutbox, error) {
	query, args, err := sq.
		Insert(MessageOutboxTable).
		Columns(
			MessageOutboxRecipientIDLabel,
			MessageOutboxMessageIDLabel,
			MessageOutboxOperationLabel,
		).
		Values(recipientID, messageID, operation).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnMessageOutbox(ctx, query, args)
}

func (s *Storage) GetMessageOutbox(ctx context.Context, limitUsers int, limitMessages int) ([]*model.MessageOutbox, error) {
	query1, args1, err := sq.
		Select(AllLabelsSelect).
		From(MessageOutboxTable).
		Where(sq.Expr(deletedATIsNullMessageOutboxFilter)).
		OrderBy(MessageOutboxRecipientIDLabel).
		Limit(uint64(limitUsers)).
		Suffix(SkipLocked).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build lock query: %w", err)
	}

	lockedRows, err := s.doAndReturnMessageOutboxes(ctx, query1, args1)
	if err != nil {
		return nil, fmt.Errorf("lock rows: %w", err)
	}

	recipientIDsMap := make(map[string]struct{})
	for _, msg := range lockedRows {
		recipientIDsMap[msg.RecipientID] = struct{}{}
	}
	recipientIDs := make([]interface{}, 0, len(recipientIDsMap))
	for id := range recipientIDsMap {
		recipientIDs = append(recipientIDs, id)
	}

	query2, args2, err := sq.
		Select(AllLabelsSelect).
		From(MessageOutboxTable).
		Where(sq.Eq{MessageOutboxRecipientIDLabel: recipientIDs}).
		Where(sq.Expr(deletedATIsNullMessageOutboxFilter)).
		Limit(uint64(limitMessages)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build main query sql: %w", err)
	}

	return s.doAndReturnMessageOutboxes(ctx, query2, args2)
}

func (s *Storage) DeleteMessageOutbox(ctx context.Context, ids []int) ([]*model.MessageOutbox, error) {
	if len(ids) == 0 {
		return []*model.MessageOutbox{}, nil
	}

	query, args, err := sq.
		Update(MessageOutboxTable).
		Set(MessageOutboxDeletedAtLabel, time.Now().UTC()).
		Where(sq.Eq{MessageOutboxIDLabel: ids}).
		Where(sq.Expr(deletedATIsNullMessageOutboxFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnMessageOutboxes(ctx, query, args)
}
