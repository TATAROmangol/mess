package kafka

import (
	"context"
	"time"

	"github.com/1ocknight/mess/shared/messagequeue"
	"github.com/segmentio/kafka-go"
)

type ProducerConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(cfg ProducerConfig) messagequeue.Producer {
	return &Producer{
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:  cfg.Brokers,
			Topic:    cfg.Topic,
			Balancer: &kafka.Hash{},
		}),
	}
}

func (p *Producer) Publish(ctx context.Context, pair *messagequeue.KeyValPair) error {
	kMsg := kafka.Message{
		Key:   pair.Key,
		Value: pair.Val,
		Time:  time.Now(),
	}

	return p.writer.WriteMessages(ctx, kMsg)
}

func (p *Producer) BatchPublish(ctx context.Context, pairs []*messagequeue.KeyValPair) error {
	msgs := make([]kafka.Message, 0, len(pairs))

	for _, pair := range pairs {
		msgs = append(msgs, kafka.Message{
			Key:   pair.Key,
			Value: pair.Val,
			Time:  time.Now(),
		})
	}

	return p.writer.WriteMessages(ctx, msgs...)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
