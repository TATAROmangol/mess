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

	GetChatByID(ctx context.Context, chatID string) (*model.Chat, error)
	GetChatIDBySubjects(ctx context.Context, firstSubjectID, secondSubjectID string) (*model.Chat, error)
	GetChatsBySubjectID(ctx context.Context, subjectID string, filter *postgres.PaginationFilter) ([]*model.Chat, error)

	IncrementChatMessageNumber(ctx context.Context, chatID string) (*model.Chat, error)

	DeleteChat(ctx context.Context, chatID string) (*model.Chat, error)
}

type LastRead interface {
	CreateLastRead(ctx context.Context, subjectID string, chatID string) (*model.LastRead, error)

	GetLastReadByChatIDs(ctx context.Context, subjectID string, chatIDs []string) ([]*model.LastRead, error)

	UpdateLastRead(ctx context.Context, subjectID string, chatID string, messageNumber int) (*model.LastRead, error)

	DeleteLastRead(ctx context.Context, subjectID string, chatID string) (*model.LastRead, error)
}

type Message interface {
	CreateMessage(ctx context.Context, chatID string, senderSubjectID string, content string, number int) (*model.Message, error)

	GetLastMessagesByChatsID(ctx context.Context, chatsID []string) ([]*model.Message, error)
	GetMessagesByChatID(ctx context.Context, chatID string, filter *postgres.PaginationFilter) ([]*model.Message, error)

	UpdateMessageContent(ctx context.Context, messageID string, content string, version int) (*model.Message, error)

	DeleteMessagesChatID(ctx context.Context, chatID string) (*model.Message, error)
}

type Service interface {
	WithTransaction(ctx context.Context) (ServiceTransaction, error)
	Chat() Chat
	LastRead() LastRead
	Message() Message
}

type ServiceTransaction interface {
	Chat() Chat
	LastRead() LastRead
	Message() Message
	Commit() error
	Rollback() error
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
