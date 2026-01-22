package worker

import (
	"context"
	"encoding/json"
	"fmt"

	mqdto "github.com/TATAROmangol/mess/shared/dto/mq"
	wsdto "github.com/TATAROmangol/mess/shared/dto/ws"
	"github.com/TATAROmangol/mess/shared/kafkav2"
	"github.com/TATAROmangol/mess/shared/logger"
	"github.com/TATAROmangol/mess/websocket/internal/model"
)

type MessageWorkerConfig struct {
	Kafka kafkav2.ConsumerConfig `yaml:"kafka_consumer"`
}

type MessageWorker struct {
	Consumer    *kafkav2.Consumer
	hubMessages chan *model.Message
	lg          logger.Logger
}

func NewMessageWorker(cfg MessageWorkerConfig, hubMessages chan *model.Message, lg logger.Logger) (*MessageWorker, error) {
	consumer, err := kafkav2.NewConsumer(cfg.Kafka)
	if err != nil {
		return nil, fmt.Errorf("new consumer: %w", err)
	}

	return &MessageWorker{
		Consumer:    consumer,
		hubMessages: hubMessages,
		lg:          lg,
	}, nil
}

func (mw *MessageWorker) Send(kafkamessages chan *kafkav2.ConsumerMessage) {
	for kfMsg := range kafkamessages {
		var mqdtoMsg mqdto.SendMessage
		err := json.Unmarshal(kfMsg.Value, &mqdtoMsg)
		if err != nil {
			mw.lg.Error(fmt.Errorf("unmarshal: %w", err))
			continue
		}

		wsdtoMsg := wsdto.Message{
			ChatID:    mqdtoMsg.ChatID,
			SenderID:  mqdtoMsg.Message.SenderID,
			Content:   mqdtoMsg.Message.Content,
			Version:   mqdtoMsg.Message.Version,
			CreatedAt: mqdtoMsg.Message.CreatedAt,
		}

		data, err := wsdtoMsg.GetData()
		if err != nil {
			mw.lg.Error(fmt.Errorf("get data: %w", err))
			continue
		}

		wsdtoWSMsg := wsdto.WSMessage{
			Data: data,
		}

		if mqdtoMsg.Operation == mqdto.AddOperation {
			wsdtoWSMsg.Type = wsdto.SendMessage
		} else {
			wsdtoWSMsg.Type = wsdto.UpdateMessage
		}

		res := model.Message{
			SubjectID: mqdtoMsg.RecipientID,
			WSMessage: &wsdtoWSMsg,
		}
		mw.hubMessages <- &res

		mw.lg.With("message", res).Info("ok")
	}
}

func (mw *MessageWorker) Run(ctx context.Context) {
	err := mw.Consumer.Start(ctx)
	if err != nil {
		mw.lg.Error(fmt.Errorf("start: %w", err))
		return
	}

	msgs := mw.Consumer.GetMessagesChan()
	go mw.Send(msgs)

	errorsCh := mw.Consumer.GetErrorsChan()
	go func() {
		for err := range errorsCh {
			mw.lg.Error(err)
		}
	}()

	mw.lg.Info("start message worker")

	<-ctx.Done()
	mw.Consumer.Close()
}
