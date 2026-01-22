package httpdto

import "time"

type MessageResponse struct {
	ID        int       `json:"id"`
	Version   int       `json:"version"`
	Content   string    `json:"content"`
	SenderID  string    `json:"sender_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatResponse struct {
	ChatID          int    `json:"chat_id"`
	SecondSubjectID string `json:"second_subject_id"`

	// map subject_id -> message_id
	LastReads map[string]int `json:"last_reads"`

	Messages []*MessageResponse `json:"messages"`
}

type ChatsMetadataResponse struct {
	ChatID          int       `json:"chat_id"`
	SecondSubjectID string    `json:"second_subject_id"`
	UpdatedAt       time.Time `json:"updated_at"`

	LastMessage *MessageResponse `json:"last_message"`

	UnreadCount       int  `json:"unread_count"`
	IsLastMessageRead bool `json:"is_last_message_read"`
}

type AddMessageRequest struct {
	ChatID  int    `json:"chat_id"`
	Content string `json:"content"`
}

type UpdateMessageRequest struct {
	MessageID int    `json:"message_id"`
	Content   string `json:"content"`
	Version   int    `json:"version"`
}

type UpdateLastReadRequest struct {
	ChatID    int `json:"chat_id"`
	MessageID int `json:"message_id"`
}
