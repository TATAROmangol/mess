package kafkav2

import (
	"context"
	"fmt"
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

	client   sarama.Client
	consumer sarama.Consumer

	msgCh chan *ConsumerMessage
	wg    sync.WaitGroup
}

func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	client, err := sarama.NewClient(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &Consumer{
		cfg:      cfg,
		client:   client,
		consumer: consumer,
		msgCh:    make(chan *ConsumerMessage),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	partitions, err := c.consumer.Partitions(c.cfg.Topic)
	if err != nil {
		return fmt.Errorf("failed to get partitions: %w", err)
	}

	for _, partition := range partitions {
		pc, err := c.consumer.ConsumePartition(c.cfg.Topic, partition, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("failed to consume partition %d: %w", partition, err)
		}

		c.wg.Add(1)
		go func(pc sarama.PartitionConsumer, partition int32) {
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
		}(pc, partition)
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

func (c *Consumer) Close() error {
	if err := c.consumer.Close(); err != nil {
		return err
	}
	return c.client.Close()
}
