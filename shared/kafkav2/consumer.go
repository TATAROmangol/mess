package kafkav2

import (
	"context"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
)

type ConsumerConfig struct {
	Brokers       []string `yaml:"brokers"`
	Topics        []string `yaml:"topics"`
	MessagesLimit int      `yaml:"messages_limit"`
}

type PartitionMessage struct {
	Value     []byte
	partition int32
	offset    int64
}

type PartitionConsumer struct {
	brokers []string
	topic   string

	client   sarama.Client
	consumer sarama.Consumer

	msgCh chan *PartitionMessage
	wg    sync.WaitGroup
}

func NewPartitionConsumer(brokers []string, topic string) (*PartitionConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &PartitionConsumer{
		brokers:  brokers,
		topic:    topic,
		client:   client,
		consumer: consumer,
		msgCh:    make(chan *PartitionMessage),
	}, nil
}

func (c *PartitionConsumer) Start(ctx context.Context) error {
	partitions, err := c.consumer.Partitions(c.topic)
	if err != nil {
		return fmt.Errorf("failed to get partitions: %w", err)
	}

	for _, partition := range partitions {
		pc, err := c.consumer.ConsumePartition(c.topic, partition, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("failed to consume partition %d: %w", partition, err)
		}

		c.wg.Add(1)
		go func(pc sarama.PartitionConsumer, partition int32) {
			defer c.wg.Done()
			for {
				select {
				case msg := <-pc.Messages():
					c.msgCh <- &PartitionMessage{
						Value:     msg.Value,
						partition: msg.Partition,
						offset:    msg.Offset,
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

func (c *PartitionConsumer) GetMessagesChan() chan *PartitionMessage {
	return c.msgCh
}

func (c *PartitionConsumer) Close() error {
	if err := c.consumer.Close(); err != nil {
		return err
	}
	return c.client.Close()
}
