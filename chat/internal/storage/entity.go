package storage

import (
	"time"

	"github.com/TATAROmangol/mess/chat/internal/model"
)

type ChatEntity struct {
	ID              int        `db:"id"`
	FirstSubjectID  string     `db:"first_subject_id"`
	SecondSubjectID string     `db:"second_subject_id"`
	MessagesCount   int        `db:"messages_count"`
	UpdatedAt       time.Time  `db:"updated_at"`
	CreatedAt       time.Time  `db:"created_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
}

func (e *ChatEntity) ToModel() *model.Chat {
	return &model.Chat{
		ID:              e.ID,
		FirstSubjectID:  e.FirstSubjectID,
		SecondSubjectID: e.SecondSubjectID,
		MessagesCount:   e.MessagesCount,
		UpdatedAt:       e.UpdatedAt,
		CreatedAt:       e.CreatedAt,
		DeletedAt:       e.DeletedAt,
	}
}

func ChatEntitiesToModels(entities []*ChatEntity) []*model.Chat {
	models := make([]*model.Chat, 0, len(entities))
	for _, entity := range entities {
		models = append(models, entity.ToModel())
	}
	return models
}

type LastReadEntity struct {
	SubjectID     string     `db:"subject_id"`
	ChatID        int        `db:"chat_id"`
	MessageID     int        `db:"message_id"`
	MessageNumber int        `db:"message_number"`
	UpdatedAt     time.Time  `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at"`
}

func (e *LastReadEntity) ToModel() *model.LastRead {
	return &model.LastRead{
		SubjectID:     e.SubjectID,
		ChatID:        e.ChatID,
		MessageID:     e.MessageID,
		MessageNumber: e.MessageNumber,
		UpdatedAt:     e.UpdatedAt,
		DeletedAt:     e.DeletedAt,
	}
}

func LastReadEntitiesToModels(entities []*LastReadEntity) []*model.LastRead {
	models := make([]*model.LastRead, 0, len(entities))
	for _, entity := range entities {
		models = append(models, entity.ToModel())
	}
	return models
}

type MessageEntity struct {
	ID              int        `db:"id"`
	ChatID          int        `db:"chat_id"`
	SenderSubjectID string     `db:"sender_subject_id"`
	Content         string     `db:"content"`
	Number          int        `db:"number"`
	Version         int        `db:"version"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
}

func (e *MessageEntity) ToModel() *model.Message {
	return &model.Message{
		ID:              e.ID,
		ChatID:          e.ChatID,
		SenderSubjectID: e.SenderSubjectID,
		Content:         e.Content,
		Number:          e.Number,
		Version:         e.Version,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
		DeletedAt:       e.DeletedAt,
	}
}

func MessageEntitiesToModels(entities []*MessageEntity) []*model.Message {
	models := make([]*model.Message, 0, len(entities))
	for _, entity := range entities {
		models = append(models, entity.ToModel())
	}
	return models
}

type MessageOutboxEntity struct {
	ID        int        `db:"id"`
	ChatID    int        `db:"chat_id"`
	MessageID int        `db:"message_id"`
	Operation int        `db:"operation"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (e *MessageOutboxEntity) ToModel() *model.MessageOutbox {
	return &model.MessageOutbox{
		ID:        e.ID,
		ChatID:    e.ChatID,
		MessageID: e.MessageID,
		Operation: model.Operation(e.Operation),
		DeletedAt: e.DeletedAt,
	}
}

func MessageOutboxEntitiesToModels(entities []*MessageOutboxEntity) []*model.MessageOutbox {
	models := make([]*model.MessageOutbox, 0, len(entities))
	for _, entity := range entities {
		models = append(models, entity.ToModel())
	}
	return models
}
