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

func (s *Storage) CreateMessage(ctx context.Context, chatID int, senderSubjectID string, content string, number int) (*model.Message, error) {
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

func (s *Storage) GetMessageByID(ctx context.Context, messageID int) (*model.Message, error) {
	query, args, err := sq.
		Select(AllLabelsSelect).
		From(MessageTable).
		Where(sq.Eq{MessageIDLabel: messageID}).
		Where(sq.Expr(deletedATIsNullMessageFilter)).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnMessage(ctx, query, args)
}

func (s *Storage) GetLastMessagesByChatsID(ctx context.Context, chatsID []int) ([]*model.Message, error) {
	aliasRowNumber := "rn"
	subQuery := sq.
		Select(AllLabelsSelect, fmt.Sprintf(
			"ROW_NUMBER() OVER (PARTITION BY %v ORDER BY %v %v) AS %v",
			MessageChatIDLabel, MessageCreatedAtLabel, DescSortLabel, aliasRowNumber,
		)).
		From(MessageTable).
		Where(sq.Eq{MessageChatIDLabel: chatsID}).
		Where(sq.Expr(deletedATIsNullMessageFilter)).
		PlaceholderFormat(sq.Dollar)

	query, args, err := sq.
		Select(
			MessageIDLabel,
			MessageChatIDLabel,
			MessageSenderSubjectIDLabel,
			MessageContentLabel,
			MessageNumberLabel,
			MessageVersionLabel,
			MessageCreatedAtLabel,
			MessageUpdatedAtLabel,
			MessageDeletedAtLabel,
		).
		FromSelect(subQuery, "sub").
		Where(sq.Eq{aliasRowNumber: 1}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnMessages(ctx, query, args)
}

func (s *Storage) GetMessagesByChatID(ctx context.Context, chatID int, filter *PaginationFilterIntLastID) ([]*model.Message, error) {
	b := sq.
		Select(AllLabelsSelect).
		From(MessageTable).
		Where(sq.Eq{MessageChatIDLabel: chatID}).
		Where(sq.Expr(deletedATIsNullMessageFilter))

	storageFilter := &postgres.PaginationFilter[int]{
		Limit:     filter.Limit,
		Asc:       filter.Asc,
		SortLabel: filter.SortLabel,
		IDLabel:   MessageIDLabel,
		LastID:    filter.LastID,
	}
	query, args, err := postgres.MakeQueryWithPagination(ctx, b, storageFilter)
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnMessages(ctx, query, args)
}

func (s *Storage) UpdateMessageContent(ctx context.Context, messageID int, content string, version int) (*model.Message, error) {
	query, args, err := sq.
		Update(MessageTable).
		Set(MessageContentLabel, content).
		Set(MessageUpdatedAtLabel, time.Now().UTC()).
		Set(MessageVersionLabel, version+1).
		Where(sq.Eq{MessageIDLabel: messageID}).
		Where(sq.Eq{MessageVersionLabel: version}).
		Where(sq.Expr(deletedATIsNullMessageFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnMessage(ctx, query, args)
}

func (s *Storage) DeleteMessagesChatID(ctx context.Context, chatID int) ([]*model.Message, error) {
	query, args, err := sq.
		Update(MessageTable).
		Set(MessageDeletedAtLabel, time.Now().UTC()).
		Set(MessageUpdatedAtLabel, time.Now().UTC()).
		Where(sq.Eq{MessageChatIDLabel: chatID}).
		Where(sq.Expr(deletedATIsNullMessageFilter)).
		Suffix(ReturningSuffix).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	return s.doAndReturnMessages(ctx, query, args)
}
