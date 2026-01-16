package storage

import (
	"chat/internal/model"
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/TATAROmangol/mess/shared/postgres"
	"github.com/jmoiron/sqlx"
)

var (
	deletedATIsNullChatFilter = fmt.Sprintf("%v %v", ChatDeletedAtLabel, IsNullLabel)
)

func (s *Storage) doAndReturnChat(ctx context.Context, query string, args []interface{}) (*model.Chat, error) {
	var entity ChatEntity
	err := sqlx.GetContext(ctx, s.exec, &entity, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return entity.ToModel(), nil
}

func (s *Storage) doAndReturnChats(ctx context.Context, query string, args []interface{}) ([]*model.Chat, error) {
	var entities []*ChatEntity
	err := sqlx.SelectContext(ctx, s.exec, &entities, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return ChatEntitiesToModels(entities), nil
}

func (s *Storage) CreateChat(ctx context.Context, firstSubjectID, secondSubjectID string) (*model.Chat, error) {
	query, args, err := sq.
		Insert(ChatTable).
		Columns(
			ChatFirstSubjectIDLabel,
			ChatSecondSubjectIDLabel,
		).
		Values(firstSubjectID, secondSubjectID).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnChat(ctx, query, args)
}

func (s *Storage) GetChatByID(ctx context.Context, chatID string) (*model.Chat, error) {
	query, args, err := sq.
		Select(AllLabelsSelect).
		From(ChatTable).
		Where(sq.Eq{ChatIDLabel: chatID}).
		Where(sq.Expr(deletedATIsNullChatFilter)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnChat(ctx, query, args)
}

func (s *Storage) GetChatIDBySubjects(ctx context.Context, firstSubjectID string, secondSubjectID string) (*model.Chat, error) {
	subjects := []string{firstSubjectID, secondSubjectID}
	query, args, err := sq.
		Select(AllLabelsSelect).
		From(ChatTable).
		Where(sq.Eq{ChatFirstSubjectIDLabel: subjects}).
		Where(sq.Eq{ChatSecondSubjectIDLabel: subjects}).
		Where(sq.Expr(deletedATIsNullChatFilter)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnChat(ctx, query, args)
}

func (s *Storage) GetChatsBySubjectID(ctx context.Context, subjectID string, filter *postgres.PaginationFilter) ([]*model.Chat, error) {
	b := sq.
		Select(AllLabelsSelect).
		From(ChatTable).
		Where(sq.Or{
			sq.Eq{ChatFirstSubjectIDLabel: subjectID},
			sq.Eq{ChatSecondSubjectIDLabel: subjectID},
		}).
		Where(sq.Expr(deletedATIsNullChatFilter))

	query, args, err := postgres.MakeQueryWithPagination(ctx, b, filter)
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnChats(ctx, query, args)
}

func (s *Storage) IncrementChatMessageNumber(ctx context.Context, chatID string) (*model.Chat, error) {
	query, args, err := sq.
		Update(ChatTable).
		Set(ChatMessagesCount, sq.Expr(fmt.Sprintf("%s + 1", ChatMessagesCount))).
		Set(ChatUpdatedAtLabel, time.Now().UTC()).
		Where(sq.Eq{ChatIDLabel: chatID}).
		Where(sq.Expr(deletedATIsNullChatFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnChat(ctx, query, args)
}

func (s *Storage) DeleteChat(ctx context.Context, chatID string) (*model.Chat, error) {
	query, args, err := sq.
		Update(ChatTable).
		Set(ChatDeletedAtLabel, time.Now().UTC()).
		Set(ChatUpdatedAtLabel, time.Now().UTC()).
		Where(sq.Eq{ChatIDLabel: chatID}).
		Where(sq.Expr(deletedATIsNullChatFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnChat(ctx, query, args)
}
