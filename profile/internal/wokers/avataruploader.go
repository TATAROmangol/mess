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
	Kafka kafka.ConsumerConfig `yaml:"kafka"`
	Delay time.Duration        `yaml:"delay"`
}

type AvatarUploaderMessage struct {
	Records []struct {
		S3 struct {
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

func (aum *AvatarUploaderMessage) Key() string {
	return aum.Records[0].S3.Object.Key
}

type AvatarUploader struct {
	CFG      AvatarUploaderConfig
	Consumer messagequeue.Consumer
	Storage  storage.Service
}

func NewAvatarUploader(cfg AvatarUploaderConfig, storage storage.Service) *AvatarUploader {
	consumer := kafka.NewConsumer(cfg.Kafka)
	return &AvatarUploader{
		CFG:      cfg,
		Consumer: consumer,
		Storage:  storage,
	}
}

func (au *AvatarUploader) Upload(ctx context.Context) error {
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return fmt.Errorf("extract logger: %v", err)
	}

	mqMsg, err := au.Consumer.ReadMessage(ctx)
	if err != nil {
		return fmt.Errorf("read message: %v", err)
	}

	var msg AvatarUploaderMessage
	if err := json.Unmarshal(mqMsg.Value(), &msg); err != nil {
		return fmt.Errorf("unmarshal: %v", err)
	}

	ind, err := model.ParseAvatarKey(msg.Key())
	if err != nil {
		return fmt.Errorf("parse avatar key token: %v", err)
	}

	prevProfile, err := au.Storage.Profile().GetProfileFromSubjectID(ctx, ind.SubjectID)
	if err != nil {
		return fmt.Errorf("profile get profile from subject id: %v", err)
	}

	if utils.StringPtrEqual(prevProfile.AvatarKey, utils.StringPtr(msg.Key())) {
		lg = lg.With(loglables.AvatarKey, msg.Key)
		lg.Info("skip duplicate")
		return nil
	}

	if !utils.StringPtrEqual(ind.PreviousKey, prevProfile.AvatarKey) {
		outboxKey, err := au.Storage.AvatarOutbox().AddKey(ctx, ind.SubjectID, msg.Key())
		if err != nil {
			return fmt.Errorf("avatar key outbox add key: %v", err)
		}
		lg = lg.With(loglables.AvatarOutbox, *outboxKey)
		if err = au.Consumer.Commit(ctx, mqMsg); err != nil {
			return fmt.Errorf("commit: %v", err)
		}
		lg.Info("add avatar outbox, skip old message")
		return nil
	}

	if ind.PreviousKey == nil {
		profile, err := au.Storage.Profile().UpdateAvatarKey(ctx, ind.SubjectID, msg.Key())
		if err != nil {
			return fmt.Errorf("profile update avatar key: %v", err)
		}
		lg = lg.With(loglables.Profile, *profile)
		if err = au.Consumer.Commit(ctx, mqMsg); err != nil {
			return fmt.Errorf("commit: %v", err)
		}
		lg.Info("success update")
		return nil
	}

	tx, err := au.Storage.WithTransaction(ctx)
	if err != nil {
		return fmt.Errorf("with transaction: %v", err)
	}
	defer tx.Rollback()

	profile, err := tx.Profile().UpdateAvatarKey(ctx, ind.SubjectID, msg.Key())
	if err != nil {
		return fmt.Errorf("profile update avatar key: %v", err)
	}
	lg = lg.With(loglables.Profile, *profile)

	outbox, err := au.Storage.AvatarOutbox().AddKey(ctx, ind.SubjectID, *prevProfile.AvatarKey)
	if err != nil {
		return fmt.Errorf("avatar key outbox add key: %v", err)
	}
	lg = lg.With(loglables.AvatarOutbox, *outbox)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %v", err)
	}

	if err = au.Consumer.Commit(ctx, mqMsg); err != nil {
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
			case <-time.After(au.CFG.Delay):
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
