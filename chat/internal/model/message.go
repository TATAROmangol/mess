package model

import "time"

type Message struct {
	ID              string
	ChatID          string
	SenderSubjectID string
	Content         string
	Number          int
	Version         int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}
