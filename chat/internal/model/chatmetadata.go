package model

import "time"

type LastMessage struct {
	MessageID int
	Content   string
	SenderID  string
}

type ChatMetadata struct {
	ChatID          int
	SecondSubjectID string
	UpdatedAt       time.Time

	LastMessage *LastMessage

	UnreadCount       int
	IsLastMessageRead bool
}
