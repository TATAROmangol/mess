package model

import "time"

type LastReadOutbox struct {
	ID          int
	RecipientID string
	ChatID      int
	SubjectID   string
	MessageID   int
	DeletedAt   *time.Time
}

func GetIDsFromLastReadOutbox(outboxes []*LastReadOutbox) []int {
	res := make([]int, 0, len(outboxes))
	for _, out := range outboxes {
		res = append(res, out.ID)
	}
	return res
}
