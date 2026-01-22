package worker

// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	"github.com/TATAROmangol/mess/chat/internal/storage"
// 	"github.com/TATAROmangol/mess/shared/kafkav2"
// 	"github.com/TATAROmangol/mess/shared/logger"
// )

// type LastReadConfig struct {
// 	Kafka          kafkav2.GroupConsumerConfig `yaml:"kafka_group_consumer"`
// 	Delay          time.Duration               `yaml:"delay"`
// 	RetryReconnect int                         `yaml:"retry_reconnect"`
// }

// type LastReadWorker struct {
// 	Consumer *kafkav2.GroupConsumer
// 	Storage  storage.Service
// 	lg       logger.Logger
// 	cfg      *LastReadConfig
// }

// func NewLastReadWorker(storage storage.Service, lg logger.Logger, cfg *LastReadConfig) (*LastReadWorker, error) {
// 	consumer, err := kafkav2.NewGroupConsumer(cfg.Kafka)
// 	if err != nil {
// 		return nil, fmt.Errorf("new group consumer: %w", err)
// 	}
// 	return &LastReadWorker{
// 		Consumer: consumer,
// 		Storage:  storage,
// 		lg:       lg,
// 		cfg:      cfg,
// 	}, nil
// }

// func (lrw *LastReadWorker) Read(msgs chan *kafkav2.GroupConsumerMessage) ([]int, error) {
// 	for message := range msgs{

// 	}
// }

// func (lrw *LastReadWorker) Run(ctx context.Context) {
// 	go func() {
// 		defer cancel()
// 		err
// 		if err := lrw.Consumer.Run(ctx); err != nil {
// 			lrw.lg.Error(fmt.Errorf("consumer run: %w", err))
// 			return
// 		}
// 	}()

// 	ticker := time.NewTicker(mw.cfg.Delay)
// 	defer ticker.Stop()

// 	defer mw.Producer.Close()
// }
