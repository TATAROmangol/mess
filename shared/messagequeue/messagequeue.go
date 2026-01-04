package messagequeue

import (
	"context"
)

type Message interface {
	Value() []byte
}

type Consumer interface {
	ReadMessage(ctx context.Context) (Message, error)
	Commit(ctx context.Context, msg Message) error
	Close() error
}
