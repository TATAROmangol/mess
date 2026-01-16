package model

import "time"

type Chat struct {
	ID              string
	FirstSubjectID  string
	SecondSubjectID string
	MessagesCount   int
	UpdatedAt       time.Time
	CreatedAt       time.Time
	DeletedAt       *time.Time
}
