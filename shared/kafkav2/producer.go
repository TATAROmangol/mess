package kafkav2

import (
	"context"
	"time"

	"github.com/IBM/sarama"
)

type KeyValPair struct {
	Key []byte
	Val []byte
}

type ProducerConfig struct {
	Brokers []string      `yaml:"brokers"`
	Topic   string        `yaml:"topic"`
	Retry   int           `yaml:"retry"`
	Timeout time.Duration `yaml:"timeout"`
}

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewProducer(cfg ProducerConfig) (*Producer, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.RequiredAcks = sarama.NoResponse
	saramaCfg.Producer.Retry.Max = cfg.Retry
	saramaCfg.Producer.Timeout = cfg.Timeout

	prod, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: prod,
		topic:    cfg.Topic,
	}, nil
}

func (p *Producer) Publish(ctx context.Context, pairs []*KeyValPair) error {
	var saramaMsgs []*sarama.ProducerMessage
	for _, pair := range pairs {
		saramaMsgs = append(saramaMsgs, &sarama.ProducerMessage{
			Topic:     p.topic,
			Key:       sarama.ByteEncoder(pair.Key),
			Value:     sarama.ByteEncoder(pair.Val),
			Timestamp: time.Now(),
		})
	}

	return p.producer.SendMessages(saramaMsgs)
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
