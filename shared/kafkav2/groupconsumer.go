package kafkav2

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
)

type GroupConsumerMessage struct {
	Value   []byte
	message *sarama.ConsumerMessage
	session sarama.ConsumerGroupSession
}

type consumerGroupHandler struct {
	msgCh chan<- *GroupConsumerMessage
}

func (h *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}
func (h consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.msgCh <- &GroupConsumerMessage{
			Value:   message.Value,
			message: message,
			session: session,
		}
	}

	return nil
}

type GroupConsumerConfig struct {
	Brokers []string `yaml:"brokers"`
	Topics  []string `yaml:"topics"`
	GroupID string   `yaml:"group_id"`
}

type GroupConsumer struct {
	cfg    GroupConsumerConfig
	client sarama.ConsumerGroup

	handler   *consumerGroupHandler
	messageCh chan *GroupConsumerMessage
	cancel    context.CancelFunc
	done      chan struct{}
}

func NewConsumer(cfg GroupConsumerConfig) (*GroupConsumer, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	client, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("new consumer group: %w", err)
	}

	msgCh := make(chan *GroupConsumerMessage)
	handler := &consumerGroupHandler{
		msgCh: msgCh,
	}

	return &GroupConsumer{
		cfg:       cfg,
		client:    client,
		handler:   handler,
		messageCh: msgCh,
	}, nil
}

func (c *GroupConsumer) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.done = make(chan struct{})

	for {
		if err := c.client.Consume(ctx, c.cfg.Topics, c.handler); err != nil {
			return fmt.Errorf("consume: %w", err)
		}

		if ctx.Err() != nil && errors.Is(ctx.Err(), context.Canceled) {
			return ctx.Err()
		}
	}
}

func (c *GroupConsumer) GetMessagesChan() chan *GroupConsumerMessage {
	return c.messageCh
}

func (c *GroupConsumer) Commit(msg *GroupConsumerMessage) {
	msg.session.MarkMessage(msg.message, "")
}

func (c *GroupConsumer) Close() error {
	c.cancel()
	close(c.messageCh)
	return c.client.Close()
}
