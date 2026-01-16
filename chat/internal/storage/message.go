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
	deletedATIsNullMessageFilter = fmt.Sprintf("%v %v", MessageDeletedAtLabel, IsNullLabel)
)

func (s *Storage) doAndReturnMessage(ctx context.Context, query string, args []interface{}) (*model.Message, error) {
	var entity MessageEntity
	err := sqlx.GetContext(ctx, s.exec, &entity, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return entity.ToModel(), nil
}

func (s *Storage) doAndReturnMessages(ctx context.Context, query string, args []interface{}) ([]*model.Message, error) {
	var entities []*MessageEntity
	err := sqlx.SelectContext(ctx, s.exec, &entities, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db get: %w", err)
	}

	return MessageEntitiesToModels(entities), nil
}

func (s *Storage) CreateMessage(ctx context.Context, chatID string, senderSubjectID string, content string, number int) (*model.Message, error) {
	query, args, err := sq.
		Insert(MessageTable).
		Columns(
			MessageChatIDLabel,
			MessageSenderSubjectIDLabel,
			MessageContentLabel,
			MessageNumberLabel,
		).
		Values(chatID, senderSubjectID, content, number).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnMessage(ctx, query, args)
}

func (s *Storage) GetLastMessagesByChatsID(ctx context.Context, chatsID []string) ([]*model.Message, error) {
	query, args, err := sq.
		Select(AllLabelsSelect).
		From(MessageTable).
		Where(sq.Eq{MessageChatIDLabel: chatsID}).
		Where(sq.Expr(deletedATIsNullMessageFilter)).
		OrderBy(fmt.Sprintf("%s %s", MessageCreatedAtLabel, AscSortLabel)).
		Distinct().
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnMessages(ctx, query, args)
}

func (s *Storage) GetMessagesByChatID(ctx context.Context, chatID string, filter *postgres.PaginationFilter) ([]*model.Message, error) {
	b := sq.
		Select(AllLabelsSelect).
		From(MessageTable).
		Where(sq.Eq{MessageChatIDLabel: chatID}).
		Where(sq.Expr(deletedATIsNullMessageFilter))

	query, args, err := postgres.MakeQueryWithPagination(ctx, b, filter)
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnMessages(ctx, query, args)
}

func (s *Storage) UpdateMessageContent(ctx context.Context, messageID string, content string, version int) (*model.Message, error) {
	query, args, err := sq.
		Update(ChatTable).
		Set(MessageContentLabel, content).
		Set(MessageUpdatedAtLabel, time.Now().UTC()).
		Set(MessageVersionLabel, version+1).
		Where(sq.Eq{MessageIDLabel: messageID}).
		Where(sq.Eq{MessageVersionLabel: version}).
		Where(sq.Expr(deletedATIsNullChatFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnMessage(ctx, query, args)
}

func (s *Storage) DeleteMessagesChatID(ctx context.Context, chatID string) (*model.Message, error) {
	query, args, err := sq.
		Update(MessageTable).
		Set(MessageDeletedAtLabel, time.Now().UTC()).
		Set(MessageUpdatedAtLabel, time.Now().UTC()).
		Where(sq.Eq{ChatCreatedAtLabel: chatID}).
		Where(sq.Expr(deletedATIsNullMessageFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnMessage(ctx, query, args)
}
