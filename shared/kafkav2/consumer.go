package kafkav2

import (
	"fmt"

	"github.com/IBM/sarama"
)

type consumerGroupHandler struct{}

func (consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		fmt.Printf("Message topic:%s partition:%d offset:%d value:%s\n",
			message.Topic, message.Partition, message.Offset, string(message.Value))

		session.MarkMessage(message, "")
	}
	return nil
}

type ConsumerConfig struct {
	Brokers       []string `yaml:"brokers"`
	Topic         string   `yaml:"topic"`
	GroupID       string   `yaml:"group_id"`
	MessagesLimit int      `yaml:"messages_limit"`
}

type Consumer struct {
	client sarama.ConsumerGroup
}

func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	client, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("new consumer group: %w", err)
	}

	

	return &Consumer{
		client: client,
	}, nil
}

func (c *Consumer) Close() error {
	return c.client.Close()
}
