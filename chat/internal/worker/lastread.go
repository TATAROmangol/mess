package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/TATAROmangol/mess/chat/internal/loglables"
	"github.com/TATAROmangol/mess/chat/internal/storage"
	mqdto "github.com/TATAROmangol/mess/shared/dto/mq"
	"github.com/TATAROmangol/mess/shared/kafkav2"
	"github.com/TATAROmangol/mess/shared/logger"
)

type LastReadConfig struct {
	Kafka kafkav2.ProducerConfig `yaml:"kafka_producer"`
	Delay time.Duration          `yaml:"delay"`
	Limit int                    `yaml:"limit"`
}

type LastReadWorker struct {
	Producer *kafkav2.Producer
	Storage  storage.Service
	lg       logger.Logger
	cfg      *LastReadConfig
}

func NewLastReadWorker(storage storage.Service, lg logger.Logger, cfg *LastReadConfig) (*LastReadWorker, error) {
	producer, err := kafkav2.NewProducer(cfg.Kafka)
	if err != nil {
		return nil, fmt.Errorf("new producer: %w", err)
	}
	return &LastReadWorker{
		Producer: producer,
		Storage:  storage,
		lg:       lg,
		cfg:      cfg,
	}, nil
}

var (
	NoLastReadsError = fmt.Errorf("no more last reads")
)

func (lrw *LastReadWorker) Send(ctx context.Context) ([]int, error) {
	tx, err := lrw.Storage.WithTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("with transaction: %w", err)
	}
	defer tx.Rollback()

	lastReadOutbox, err := tx.LastReadOutbox().GetLastReadOutbox(ctx, lrw.cfg.Limit)
	if err != nil {
		return nil, fmt.Errorf("outbox get keys: %w", err)
	}
	if len(lastReadOutbox) == 0 {
		return nil, NoLastReadsError
	}

	pairs := make([]*kafkav2.KeyValPair, 0, len(lastReadOutbox))
	ids := make([]int, 0)
	for _, out := range lastReadOutbox {
		ids = append(ids, out.ID)

		sendMessage := mqdto.LastRead{
			ChatID:      out.ChatID,
			RecipientID: out.RecipientID,
			SubjectID:   out.SubjectID,
			MessageID:   out.MessageID,
		}

		val, err := json.Marshal(sendMessage)
		if err != nil {
			return nil, fmt.Errorf("marshal: %w", err)
		}

		pair := kafkav2.KeyValPair{
			Key: []byte(out.RecipientID),
			Val: val,
		}

		pairs = append(pairs, &pair)
	}

	if err := lrw.Producer.Publish(pairs); err != nil {
		return nil, fmt.Errorf("batch publish: %w", err)
	}

	_, err = tx.LastReadOutbox().DeleteLastReadOutbox(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("delete message outbox: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return ids, nil
}

func (lrw *LastReadWorker) Run(ctx context.Context) {
	lrw.lg.Info("run last read worker")

	ticker := time.NewTicker(lrw.cfg.Delay)
	defer ticker.Stop()

	defer lrw.Producer.Close()

	for {
		select {
		case <-ctx.Done():
			lrw.lg.Info("context done - stop")
			return
		default:
			ids, err := lrw.Send(ctx)
			if err == nil {
				lg := lrw.lg.With(loglables.IDs, ids)
				lg.Info("send messages")
				continue
			}

			if errors.Is(err, NoMessagesError) {
				lrw.lg.Info("no last reads")
			} else {
				lrw.lg.Error(fmt.Errorf("send: %w", err))
			}

			select {
			case <-ctx.Done():
				lrw.lg.Info("context done - stop")
				return
			case <-ticker.C:
				lrw.lg.Info("wait delay")
				continue
			}
		}
	}
}
