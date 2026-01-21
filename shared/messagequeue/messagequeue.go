package messagequeue

import (
	"context"
)

type Message interface {
	Value() []byte
}

type KeyValPair struct {
	Key []byte
	Val []byte
}

type Consumer interface {
	ReadMessage(ctx context.Context) (Message, error)
	Commit(ctx context.Context, msg Message) error
	Close() error
}

type Producer interface {
	Publish(ctx context.Context, pair *KeyValPair) error
	BatchPublish(ctx context.Context, pairs []*KeyValPair) error
	Close() error
}
