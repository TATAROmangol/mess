package model

import "time"

type LastMessage struct {
	MessageID int
	Content  string
	SenderID string
}

type ChatMetadata struct {
	ChatID            int
	UpdatedAt         time.Time

	LastMessage       *LastMessage

	UnreadCount       int
	IsLastMessageRead bool
}
