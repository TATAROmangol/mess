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

func (c *Chat) GetSecondSubject(subj string) string {
	var recipient string
	if c.FirstSubjectID != subj {
		recipient = c.FirstSubjectID
	} else {
		recipient = c.SecondSubjectID
	}

	return recipient
}
