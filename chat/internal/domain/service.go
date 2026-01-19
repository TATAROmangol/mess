package domain

import (
	"context"

	"github.com/TATAROmangol/mess/chat/internal/model"
	"github.com/TATAROmangol/mess/chat/internal/storage"
)

type Direction int

const (
	DirectionUnknown Direction = iota
	DirectionAfter
	DirectionBefore
)

type MessagePaginationFilter struct {
	Limit         int
	LastMessageID *int
	Direction     Direction
}

var DefaultPaginationMessage = storage.PaginationFilterIntLastID{
	Limit:     30,
	Asc:       false,
	SortLabel: storage.MessageCreatedAtLabel,
}

type ChatPaginationFilter struct {
	Limit      int
	LastChatID *int
	Direction  Direction
}

var DefaultPaginationChat = storage.PaginationFilterIntLastID{
	Limit:     30,
	Asc:       false,
	SortLabel: storage.ChatUpdatedAtLabel,
}

type Service interface {
	AddChat(ctx context.Context, secondSubjectID string) (*model.Chat, error)
	GetChatsMetadata(ctx context.Context, filter *ChatPaginationFilter) ([]*model.ChatMetadata, error)
	GetChatBySubjectID(ctx context.Context, secondSubjectID string) (*model.Chat, error)

	GetLastReads(ctx context.Context, chatID int) ([]*model.LastRead, error)
	UpdateLastRead(ctx context.Context, chatID int, messageID int) (*model.LastRead, error)

	GetMessages(ctx context.Context, chatID int, filter *MessagePaginationFilter) ([]*model.Message, error)
	GetMessagesToLastRead(ctx context.Context, chatID int, limit int) ([]*model.Message, error)
	SendMessage(ctx context.Context, chatID int, content string) (*model.Message, error)
	UpdateMessage(ctx context.Context, messageID int, content string, version int) (*model.Message, error)
}

type Domain struct {
	Storage storage.Service
}

func New(s storage.Service) Service {
	return &Domain{
		Storage: s,
	}
}
