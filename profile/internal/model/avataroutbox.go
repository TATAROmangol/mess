package model

import "time"

type AvatarOutbox struct {
	SubjectID string
	Key       string
	CreatedAt time.Time
	DeletedAt *time.Time
}

func GetOutboxKeys(arr []*AvatarOutbox) []string {
	res := make([]string, len(arr))
	for i, k := range arr {
		res[i] = k.Key
	}

	return res
}
