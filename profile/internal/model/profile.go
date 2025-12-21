package model

import "time"

type Profile struct {
	SubjectID string
	Alias     string
	AvatarURL string
	Version   int
	UpdatedAt time.Time
	CreatedAt time.Time
}
