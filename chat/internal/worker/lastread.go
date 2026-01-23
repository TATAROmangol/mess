package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/TATAROmangol/mess/chat/internal/storage"
	mqdto "github.com/TATAROmangol/mess/shared/dto/mq"
	"github.com/TATAROmangol/mess/shared/kafkav2"
	"github.com/TATAROmangol/mess/shared/logger"
)

type LastReadConfig struct {
	Kafka          kafkav2.GroupConsumerConfig `yaml:"kafka_group_consumer"`
	Delay          time.Duration               `yaml:"delay"`
	RetryReconnect int                         `yaml:"retry_reconnect"`
}

type LastReadWorker struct {
	Consumer *kafkav2.GroupConsumer
	Storage  storage.Service
	lg       logger.Logger
	cfg      *LastReadConfig
}

func NewLastReadWorker(storage storage.Service, lg logger.Logger, cfg *LastReadConfig) (*LastReadWorker, error) {
	consumer, err := kafkav2.NewGroupConsumer(cfg.Kafka)
	if err != nil {
		return nil, fmt.Errorf("new group consumer: %w", err)
	}
	return &LastReadWorker{
		Consumer: consumer,
		Storage:  storage,
		lg:       lg,
		cfg:      cfg,
	}, nil
}

func (lrw *LastReadWorker) Read(msgs chan *kafkav2.GroupConsumerMessage) {
	for message := range msgs {
		mqdto.
	}
}

func (lrw *LastReadWorker) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		if err := lrw.Consumer.Run(ctx); err != nil {
			lrw.lg.Error(fmt.Errorf("consumer run: %w", err))
			return
		}
	}()

	msgs := lrw.Consumer.GetMessagesChan()

	go lrw.Read(msgs)

	<-ctx.Done()

	defer lrw.Consumer.Close()
}
