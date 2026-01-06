package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TATAROmangol/mess/profile/internal/ctxkey"
	"github.com/TATAROmangol/mess/profile/internal/loglables"
	"github.com/TATAROmangol/mess/profile/internal/model"
	"github.com/TATAROmangol/mess/profile/internal/storage"
	"github.com/TATAROmangol/mess/shared/messagequeue"
	"github.com/TATAROmangol/mess/shared/messagequeue/kafka"
	"github.com/TATAROmangol/mess/shared/utils"
)

type AvatarUploaderConfig struct {
	kafka kafka.ConsumerConfig `yaml:"kafka"`
	delay time.Duration        `yaml:"delay"`
}

type AvatarUploaderMessage struct {
	Key string `json:"Key"`
}

type AvatarUploader struct {
	cfg      AvatarUploaderConfig
	consumer messagequeue.Consumer
	storage  storage.Service
}

func NewAvatarUploader(cfg AvatarUploaderConfig, storage storage.Service) *AvatarUploader {
	consumer := kafka.NewConsumer(cfg.kafka)
	return &AvatarUploader{
		cfg:      cfg,
		consumer: consumer,
		storage:  storage,
	}
}

func (au *AvatarUploader) Upload(ctx context.Context) error {
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return fmt.Errorf("extract logger: %v", err)
	}

	mqMsg, err := au.consumer.ReadMessage(ctx)
	if err != nil {
		return fmt.Errorf("read message: %v", err)
	}

	var msg AvatarUploaderMessage
	if err := json.Unmarshal(mqMsg.Value(), &msg); err != nil {
		return fmt.Errorf("unmarshal: %v", err)
	}

	ind, err := model.ParseAvatarKey(msg.Key)
	if err != nil {
		return fmt.Errorf("parse avatar key token: %v", err)
	}

	profile, err := au.storage.Profile().GetProfileFromSubjectID(ctx, ind.SubjectID)
	if err != nil {
		return fmt.Errorf("profile get profile from subject id: %v", err)
	}

	if utils.StringPtrEqual(profile.AvatarKey, &msg.Key) {
		return nil
	}

	if !utils.StringPtrEqual(ind.PreviousKey, profile.AvatarKey) {
		outboxKey, err := au.storage.AvatarOutbox().AddKey(ctx, ind.SubjectID, msg.Key)
		if err != nil {
			return fmt.Errorf("avatar key outbox add key: %v", err)
		}
		lg.With(loglables.AvatarOutbox, *outboxKey)
		lg.Info("add avatar outbox")
		return nil
	}

	tx, err := au.storage.WithTransaction(ctx)
	if err != nil {
		return fmt.Errorf("with transaction: %v", err)
	}
	defer tx.Rollback()

	profile, err = tx.Profile().UpdateAvatarKey(ctx, ind.SubjectID, msg.Key)
	if err != nil {
		return fmt.Errorf("profile update avatar key: %v", err)
	}
	lg.With(loglables.Profile, *profile)

	if profile.AvatarKey == nil {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit: %v", err)
		}
		return nil
	}

	outbox, err := au.storage.AvatarOutbox().AddKey(ctx, ind.SubjectID, *profile.AvatarKey)
	if err != nil {
		return fmt.Errorf("avatar key outbox add key: %v", err)
	}
	lg.With(loglables.AvatarOutbox, *outbox)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %v", err)
	}

	if err = au.consumer.Commit(ctx, mqMsg); err != nil {
		return fmt.Errorf("commit: %v", err)
	}

	lg.Info("success update")

	return nil
}

func (au *AvatarUploader) Start(ctx context.Context) error {
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return fmt.Errorf("extract logger: %v", err)
	}

	go func() {
		for {
			err := au.Upload(ctx)
			if err == nil {
				continue
			}

			lg.Error(fmt.Errorf("upload: %v", err))

			select {
			case <-time.After(au.cfg.delay):
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
