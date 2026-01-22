package kafkav2

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/IBM/sarama"
)

type ConsumerConfig struct {
	Brokers       []string `yaml:"brokers"`
	Topic         string   `yaml:"topic"`
	MessagesLimit int      `yaml:"messages_limit"`
}

type ConsumerMessage struct {
	Value []byte
}

type Consumer struct {
	cfg ConsumerConfig

	consumer sarama.Consumer

	errorsCh chan error
	msgCh    chan *ConsumerMessage
	wg       sync.WaitGroup
}

func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &Consumer{
		cfg:      cfg,
		consumer: consumer,
		errorsCh: make(chan error),
		msgCh:    make(chan *ConsumerMessage),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	partitions, err := c.consumer.Partitions(c.cfg.Topic)
	if err != nil {
		return fmt.Errorf("failed to get partitions: %w", err)
	}
	if len(partitions) == 0 {
		return fmt.Errorf("no partitions found for topic %s", c.cfg.Topic)
	}

	for _, partition := range partitions {
		slog.Info("start partioning")
		pc, err := c.consumer.ConsumePartition(c.cfg.Topic, partition, sarama.OffsetOldest)
		if err != nil {
			return fmt.Errorf("failed to consume partition %d: %w", partition, err)
		}

		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			for {
				select {
				case msg := <-pc.Messages():
					c.msgCh <- &ConsumerMessage{
						Value: msg.Value,
					}
				case <-ctx.Done():
					pc.Close()
					return
				}
			}
		}()

		go func() {
			for err := range pc.Errors() {
				c.errorsCh <- err
			}
		}()
	}

	go func() {
		c.wg.Wait()
		close(c.msgCh)
	}()

	return nil
}

func (c *Consumer) GetMessagesChan() chan *ConsumerMessage {
	return c.msgCh
}

func (c *Consumer) GetErrorsChan() chan error {
	return c.errorsCh
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
