package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/TATAROmangol/mess/shared/messagequeue"
	"github.com/segmentio/kafka-go"
)

type Message struct {
	Key       []byte
	Val       []byte
	Topic     string
	Partition int
	Offset    int64
	Time      time.Time
}

func (m *Message) Value() []byte {
	return m.Val
}

type ConsumerConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
	GroupID string   `yaml:"group_id"`
}

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(cfg ConsumerConfig) messagequeue.Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.Brokers,
			Topic:          cfg.Topic,
			GroupID:        cfg.GroupID,
			CommitInterval: 0,
		}),
	}
}

func (c *Consumer) ReadMessage(ctx context.Context) (messagequeue.Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return nil, err
	}

	return &Message{
		Key:       msg.Key,
		Val:       msg.Value,
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Time:      msg.Time,
	}, nil
}

func (c *Consumer) Commit(ctx context.Context, msg messagequeue.Message) error {
	m, ok := msg.(*Message)
	if !ok {
		return fmt.Errorf("incorrect msg type")
	}

	kMsg := kafka.Message{
		Key:       m.Key,
		Value:     m.Val,
		Topic:     m.Topic,
		Partition: m.Partition,
		Offset:    m.Offset,
		Time:      m.Time,
	}

	return c.reader.CommitMessages(ctx, kMsg)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
