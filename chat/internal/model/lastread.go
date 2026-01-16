package model

import "time"

type LastRead struct {
	SubjectID     string
	ChatID        string
	MessageNumber int
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}
