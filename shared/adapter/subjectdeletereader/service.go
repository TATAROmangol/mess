package subjectdeletereader

import "context"

type Message interface {
	GetSubjectID() string
}

type Service interface {
	FetchMessage(ctx context.Context) (Message, error)
	Commit(msg Message) error
	Close() error
}
