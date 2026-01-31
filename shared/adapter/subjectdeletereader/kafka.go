package subjectdeletereader

import (
	"context"
	"fmt"

	mqdto "github.com/1ocknight/mess/shared/dto/mq"
	"github.com/IBM/sarama"
)

type MessageIMPL struct {
	ds      *mqdto.DeleteSubject
	message *sarama.ConsumerMessage
	session sarama.ConsumerGroupSession
}

func (m *MessageIMPL) GetSubjectID() string {
	return m.ds.GetSubjectID()
}

type handler struct {
	msgCh chan<- *MessageIMPL
}

func (h *handler) Setup(session sarama.ConsumerGroupSession) error { return nil }

func (h *handler) Cleanup(session sarama.ConsumerGroupSession) error { return nil }

func (h handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		ds, err := mqdto.UnmarshallDeleteSubject(message.Value)
		if err != nil {
			return fmt.Errorf("unmarshal delete subject message: %w", err)
		}

		h.msgCh <- &MessageIMPL{
			ds:      ds,
			message: message,
			session: session,
		}
	}

	return nil
}

type Config struct {
	Brokers []string `yaml:"brokers"`
	Topics  []string `yaml:"topics"`
	GroupID string   `yaml:"group_id"`
}

type Consumer struct {
	cfg    Config
	client sarama.ConsumerGroup

	handler   *handler
	messageCh chan *MessageIMPL
	cancel    context.CancelFunc
	done      chan struct{}
}

func New(cfg Config) (Service, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	client, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("new consumer group: %w", err)
	}

	msgCh := make(chan *MessageIMPL)
	handler := &handler{
		msgCh: msgCh,
	}

	return &Consumer{
		cfg:       cfg,
		client:    client,
		handler:   handler,
		messageCh: msgCh,
	}, nil
}

func (c *Consumer) FetchMessage(ctx context.Context) (Message, error) {
	if err := c.client.Consume(ctx, c.cfg.Topics, c.handler); err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}

	return <-c.messageCh, nil
}

func (c *Consumer) Commit(msg Message) error {
	if impl, ok := msg.(*MessageIMPL); ok {
		impl.session.MarkMessage(impl.message, "")
		return nil
	}

	return fmt.Errorf("invalid message type")
}

func (c *Consumer) Close() error {
	c.cancel()
	close(c.messageCh)
	return c.client.Close()
}
