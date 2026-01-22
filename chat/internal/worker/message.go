package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/TATAROmangol/mess/chat/internal/loglables"
	"github.com/TATAROmangol/mess/chat/internal/model"
	"github.com/TATAROmangol/mess/chat/internal/storage"
	mqdto "github.com/TATAROmangol/mess/shared/dto/mq"
	"github.com/TATAROmangol/mess/shared/kafkav2"
	"github.com/TATAROmangol/mess/shared/logger"
)

type MessageWorkerConfig struct {
	Kafka         kafkav2.ProducerConfig `yaml:"kafka_producer"`
	Delay         time.Duration          `yaml:"delay"`
	UsersLimit    int                    `yaml:"users_limit"`
	MessagesLimit int                    `yaml:"messages_limit"`
}

type MessageWorker struct {
	Producer *kafkav2.Producer
	Storage  storage.Service
	lg       logger.Logger
	cfg      *MessageWorkerConfig
}

func NewMessageWorker(storage storage.Service, lg logger.Logger, cfg *MessageWorkerConfig) (*MessageWorker, error) {
	producer, err := kafkav2.NewProducer(cfg.Kafka)
	if err != nil {
		return nil, fmt.Errorf("new producer: %w", err)
	}
	return &MessageWorker{
		Producer: producer,
		Storage:  storage,
		lg:       lg,
		cfg:      cfg,
	}, nil
}

var (
	NoMessagesError = fmt.Errorf("no more messages")
)

func (mw *MessageWorker) Send(ctx context.Context) ([]int, error) {
	tx, err := mw.Storage.WithTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("with transaction: %w", err)
	}
	defer tx.Rollback()

	messagesOutbox, err := tx.MessageOutbox().GetMessageOutbox(ctx, mw.cfg.UsersLimit, mw.cfg.MessagesLimit)
	if err != nil {
		return nil, fmt.Errorf("outbox get keys: %w", err)
	}
	if len(messagesOutbox) == 0 {
		return nil, NoMessagesError
	}

	messages, err := tx.Message().GetMessagesByIDs(ctx, model.GetMessageIDsFromMessageOutboxes(messagesOutbox))
	if err != nil {
		return nil, fmt.Errorf("get messages by ids: %w", err)
	}

	messagesMap := make(map[int]*model.Message)
	for _, mess := range messages {
		messagesMap[mess.ID] = mess
	}

	pairs := make([]*kafkav2.KeyValPair, 0, len(messagesOutbox))
	ids := make([]int, 0)
	for _, out := range messagesOutbox {
		mess, ok := messagesMap[out.MessageID]
		if !ok {
			mw.lg.Error(fmt.Errorf("not found message: %v", *mess))
			continue
		}

		ids = append(ids, out.ID)

		sendMessage := mqdto.SendMessage{
			ChatID:      mess.ChatID,
			RecipientID: out.RecipientID,
			Message: &mqdto.Message{
				ID:        mess.ID,
				SenderID:  mess.SenderSubjectID,
				Version:   mess.Version,
				Content:   mess.Content,
				CreatedAt: mess.CreatedAt,
			},
		}

		if out.Operation == model.AddOperation {
			sendMessage.Operation = mqdto.AddOperation
		} else {
			sendMessage.Operation = mqdto.UpdateOperation
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

	if err := mw.Producer.Publish(pairs); err != nil {
		return nil, fmt.Errorf("batch publish: %w", err)
	}

	_, err = tx.MessageOutbox().DeleteMessageOutbox(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("delete message outbox: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return ids, nil
}

func (mw *MessageWorker) Run(ctx context.Context) {
	mw.lg.Info("run message worker")

	ticker := time.NewTicker(mw.cfg.Delay)
	defer ticker.Stop()

	defer mw.Producer.Close()

	for {
		select {
		case <-ctx.Done():
			mw.lg.Info("context done - stop")
			return
		default:
			ids, err := mw.Send(ctx)
			if err == nil {
				lg := mw.lg.With(loglables.IDs, ids)
				lg.Info("send messages")
				continue
			}

			if errors.Is(err, NoMessagesError) {
				mw.lg.Info("no messages")
			} else {
				mw.lg.Error(fmt.Errorf("send: %w", err))
			}

			select {
			case <-ctx.Done():
				mw.lg.Info("context done - stop")
				return
			case <-ticker.C:
				mw.lg.Info("wait delay")
				continue
			}
		}
	}
}
