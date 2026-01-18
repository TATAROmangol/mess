package storage

import (
	"chat/internal/model"
	"context"
	"database/sql"
	"fmt"

	"github.com/TATAROmangol/mess/shared/postgres"
	"github.com/jmoiron/sqlx"
)

type Chat interface {
	CreateChat(ctx context.Context, firstSubjectID, secondSubjectID string) (*model.Chat, error)

	GetChatByID(ctx context.Context, chatID int) (*model.Chat, error)
	GetChatIDBySubjects(ctx context.Context, firstSubjectID, secondSubjectID string) (*model.Chat, error)
	GetChatsBySubjectID(ctx context.Context, subjectID string, filter *PaginationFilterIntLastID) ([]*model.Chat, error)

	IncrementChatMessageNumber(ctx context.Context, chatID int) (*model.Chat, error)

	DeleteChat(ctx context.Context, chatID int) (*model.Chat, error)
}

type LastRead interface {
	CreateLastRead(ctx context.Context, subjectID string, chatID int) (*model.LastRead, error)

	GetLastReadByChatIDs(ctx context.Context, subjectID string, chatIDs []int) ([]*model.LastRead, error)

	UpdateLastRead(ctx context.Context, subjectID string, chatID int, messageID int, messageNumber int) (*model.LastRead, error)

	DeleteLastRead(ctx context.Context, subjectID string, chatID int) (*model.LastRead, error)
}

type Message interface {
	CreateMessage(ctx context.Context, chatID int, senderSubjectID string, content string, number int) (*model.Message, error)

	GetLastMessagesByChatsID(ctx context.Context, chatsID []int) ([]*model.Message, error)
	GetMessagesByChatID(ctx context.Context, chatID int, filter *PaginationFilterIntLastID) ([]*model.Message, error)

	UpdateMessageContent(ctx context.Context, messageID int, content string, version int) (*model.Message, error)

	DeleteMessagesChatID(ctx context.Context, chatID int) ([]*model.Message, error)
}

type MessageOutbox interface {
	AddMessageOutbox(ctx context.Context, chatID int, messageID int, operation model.Operation) (*model.MessageOutbox, error)
	GetMessageOutbox(ctx context.Context, limit int) ([]*model.MessageOutbox, error)
	DeleteMessageOutbox(ctx context.Context, ids []int) ([]*model.MessageOutbox, error)
}

type Service interface {
	WithTransaction(ctx context.Context) (ServiceTransaction, error)
	Chat() Chat
	LastRead() LastRead
	Message() Message
	MessageOutbox() MessageOutbox
}

type ServiceTransaction interface {
	Chat() Chat
	LastRead() LastRead
	Message() Message
	MessageOutbox() MessageOutbox
	Commit() error
	Rollback() error
}

type PaginationFilterIntLastID struct {
	LastID    *int
	Limit     int
	Asc       bool
	SortLabel string
}

var (
	ErrNoRows = sql.ErrNoRows
)

type Storage struct {
	db   *sqlx.DB
	exec sqlx.ExtContext
}

func New(cfg postgres.Config) (Service, error) {
	db, err := postgres.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	return &Storage{
		db:   db,
		exec: db,
	}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) WithTransaction(ctx context.Context) (ServiceTransaction, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin txx: %w", err)
	}

	return &Storage{
		db:   s.db,
		exec: tx,
	}, nil
}

func (s *Storage) Chat() Chat {
	return &Storage{
		db:   s.db,
		exec: s.exec,
	}
}

func (s *Storage) LastRead() LastRead {
	return &Storage{
		db:   s.db,
		exec: s.exec,
	}
}

func (s *Storage) Message() Message {
	return &Storage{
		db:   s.db,
		exec: s.exec,
	}
}

func (s *Storage) MessageOutbox() MessageOutbox {
	return &Storage{
		db:   s.db,
		exec: s.exec,
	}
}

func (s *Storage) Commit() error {
	tx, ok := s.exec.(*sqlx.Tx)
	if !ok {
		return fmt.Errorf("commit called without transaction")
	}
	return tx.Commit()
}

func (s *Storage) Rollback() error {
	tx, ok := s.exec.(*sqlx.Tx)
	if !ok {
		return fmt.Errorf("rollback called without transaction")
	}
	return tx.Rollback()
}
