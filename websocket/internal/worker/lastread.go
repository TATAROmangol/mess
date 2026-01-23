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

type LastReadConfig struct {
	Kafka kafkav2.ConsumerConfig `yaml:"kafka_consumer"`
}

type LastReadWorker struct {
	Consumer    *kafkav2.Consumer
	hubMessages chan *model.Message
	lg          logger.Logger
}

func NewLastReadWorker(cfg LastReadConfig, hubMessages chan *model.Message, lg logger.Logger) (*LastReadWorker, error) {
	consumer, err := kafkav2.NewConsumer(cfg.Kafka)
	if err != nil {
		return nil, fmt.Errorf("new consumer: %w", err)
	}

	return &LastReadWorker{
		Consumer:    consumer,
		hubMessages: hubMessages,
		lg:          lg,
	}, nil
}

func (lrw *LastReadWorker) Send(kafkamessages chan *kafkav2.ConsumerMessage) {
	for kfMsg := range kafkamessages {
		var mqdtoMsg mqdto.LastRead
		err := json.Unmarshal(kfMsg.Value, &mqdtoMsg)
		if err != nil {
			lrw.lg.Error(fmt.Errorf("unmarshal: %w", err))
			continue
		}

		wsdtoMsg := wsdto.LastRead{
			ChatID:    mqdtoMsg.ChatID,
			SubjectID: mqdtoMsg.SubjectID,
			MessageID: mqdtoMsg.MessageID,
		}

		data, err := wsdtoMsg.GetData()
		if err != nil {
			lrw.lg.Error(fmt.Errorf("get data: %w", err))
			continue
		}

		wsdtoWSMsg := wsdto.WSMessage{
			Data: data,
			Type: wsdto.UpdateLastRead,
		}

		res := model.Message{
			SubjectID: mqdtoMsg.RecipientID,
			WSMessage: &wsdtoWSMsg,
		}
		lrw.hubMessages <- &res

		lrw.lg.With("lastread", res).Info("ok")
	}
}

func (lrw *LastReadWorker) Run(ctx context.Context) {
	err := lrw.Consumer.Start(ctx)
	if err != nil {
		lrw.lg.Error(fmt.Errorf("start: %w", err))
		return
	}

	msgs := lrw.Consumer.GetMessagesChan()
	go lrw.Send(msgs)

	errorsCh := lrw.Consumer.GetErrorsChan()
	go func() {
		for err := range errorsCh {
			lrw.lg.Error(err)
		}
	}()

	lrw.lg.Info("start message worker")

	<-ctx.Done()
	lrw.Consumer.Close()
}
