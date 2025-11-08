package entities

import "time"

type Subject struct {
	ID           int
	Login        string
	PasswordHash string
	CreatedAt    time.Time
	DeletedAt    *time.Time
	Version      int
}
