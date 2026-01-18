package model

import (
	"time"
)

type Chat struct {
	ID              int
	FirstSubjectID  string
	SecondSubjectID string
	MessagesCount   int
	UpdatedAt       time.Time
	CreatedAt       time.Time
	DeletedAt       *time.Time
}

func GetChatsID(chats []*Chat) []int {
	res := make([]int, 0, len(chats))

	for _, c := range chats {
		res = append(res, c.ID)
	}

	return res
}
