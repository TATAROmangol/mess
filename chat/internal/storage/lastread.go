package storage

import (
	"github.com/TATAROmangol/mess/chat/internal/model"
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

var (
	deletedATIsNullLastReadFilter = fmt.Sprintf("%v %v", LastReadDeletedAtLabel, IsNullLabel)
)

func (s *Storage) doAndReturnLastRead(ctx context.Context, query string, args []interface{}) (*model.LastRead, error) {
	var entity LastReadEntity
	err := sqlx.GetContext(ctx, s.exec, &entity, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return entity.ToModel(), nil
}

func (s *Storage) doAndReturnLastReads(ctx context.Context, query string, args []interface{}) ([]*model.LastRead, error) {
	var entities []*LastReadEntity
	err := sqlx.SelectContext(ctx, s.exec, &entities, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return LastReadEntitiesToModels(entities), nil
}

func (s *Storage) CreateLastRead(ctx context.Context, subjectID string, chatID int) (*model.LastRead, error) {
	query, args, err := sq.
		Insert(LastReadTable).
		Columns(
			LastReadSubjectIDLabel,
			LastReadChatIDLabel,
		).
		Values(subjectID, chatID).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnLastRead(ctx, query, args)
}

func (s *Storage) GetLastReadsByChatIDs(ctx context.Context, chatIDs []int) ([]*model.LastRead, error) {
	query, args, err := sq.
		Select(AllLabelsSelect).
		From(LastReadTable).
		Where(sq.Eq{LastReadChatIDLabel: chatIDs}).
		Where(sq.Expr(deletedATIsNullLastReadFilter)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnLastReads(ctx, query, args)
}

func (s *Storage) GetLastReadsByChatID(ctx context.Context, chatID int) ([]*model.LastRead, error) {
	query, args, err := sq.
		Select(AllLabelsSelect).
		From(LastReadTable).
		Where(sq.Eq{LastReadChatIDLabel: chatID}).
		Where(sq.Expr(deletedATIsNullLastReadFilter)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnLastReads(ctx, query, args)
}

func (s *Storage) GetLastReadBySubjectID(ctx context.Context, subjectID string, chatID int) (*model.LastRead, error) {
	query, args, err := sq.
		Select(AllLabelsSelect).
		From(LastReadTable).
		Where(sq.Eq{LastReadSubjectIDLabel: subjectID}).
		Where(sq.Eq{LastReadChatIDLabel: chatID}).
		Where(sq.Expr(deletedATIsNullLastReadFilter)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnLastRead(ctx, query, args)
}

func (s *Storage) UpdateLastRead(ctx context.Context, subjectID string, chatID int, messageID int, messageNumber int) (*model.LastRead, error) {
	query, args, err := sq.
		Update(LastReadTable).
		Set(LastReadMessageIDLabel, messageID).
		Set(LastReadMessageNumberLabel, messageNumber).
		Set(LastReadUpdatedAtLabel, time.Now().UTC()).
		Where(sq.Lt{LastReadMessageNumberLabel: messageNumber}).
		Where(sq.Eq{LastReadSubjectIDLabel: subjectID}).
		Where(sq.Eq{LastReadChatIDLabel: chatID}).
		Where(sq.Expr(deletedATIsNullLastReadFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnLastRead(ctx, query, args)
}

func (s *Storage) DeleteLastRead(ctx context.Context, subjectID string, chatID int) (*model.LastRead, error) {
	query, args, err := sq.
		Update(LastReadTable).
		Set(LastReadDeletedAtLabel, time.Now().UTC()).
		Set(LastReadUpdatedAtLabel, time.Now().UTC()).
		Where(sq.Eq{LastReadChatIDLabel: chatID}).
		Where(sq.Eq{LastReadSubjectIDLabel: subjectID}).
		Where(sq.Expr(deletedATIsNullLastReadFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnLastRead(ctx, query, args)
}
