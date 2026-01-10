package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/TATAROmangol/mess/profile/internal/ctxkey"
	"github.com/TATAROmangol/mess/profile/internal/loglables"
	"github.com/TATAROmangol/mess/profile/internal/storage"
	"github.com/TATAROmangol/mess/shared/messagequeue"
	"github.com/TATAROmangol/mess/shared/messagequeue/kafka"
)

type ProfileDeleterConfig struct {
	ClientKafka kafka.ConsumerConfig `yaml:"client_kafka"`
	AdminKafka  kafka.ConsumerConfig `yaml:"admin_kafka"`
	Delay       time.Duration        `yaml:"delay"`
}

type ProfileDeleteMessage interface {
	GetSubjectID() string
}

type ClientProfileDeleteMessage struct {
	SubjectID string `json:"userId"`
}

func (cpdm *ClientProfileDeleteMessage) GetSubjectID() string {
	return cpdm.SubjectID
}

type AdminProfileDeleteMessage struct {
	SubjectID string `json:"resourceId"`
}

func (apdm *AdminProfileDeleteMessage) GetSubjectID() string {
	return apdm.SubjectID
}

type ProfileDeleter struct {
	CFG            ProfileDeleterConfig
	ClientConsumer messagequeue.Consumer
	AdminConsumer  messagequeue.Consumer
	Profile        storage.Profile
}

func NewProfileDeleter(cfg ProfileDeleterConfig, profile storage.Profile) *ProfileDeleter {
	clientConsumer := kafka.NewConsumer(cfg.ClientKafka)
	adminConsumer := kafka.NewConsumer(cfg.AdminKafka)

	return &ProfileDeleter{
		CFG:            cfg,
		ClientConsumer: clientConsumer,
		AdminConsumer:  adminConsumer,
		Profile:        profile,
	}
}

func ProfileDelete[T ProfileDeleteMessage](ctx context.Context, cons messagequeue.Consumer, store storage.Profile) error {
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return fmt.Errorf("extract logger: %w", err)
	}

	mqMsg, err := cons.ReadMessage(ctx)
	if err != nil {
		return fmt.Errorf("read message: %w", err)
	}

	var msg T
	if err := json.Unmarshal(mqMsg.Value(), &msg); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	prof, err := store.DeleteProfile(ctx, msg.GetSubjectID())
	if err != nil && errors.Is(err, storage.ErrNoRows) {
		if err := cons.Commit(ctx, mqMsg); err != nil {
			return fmt.Errorf("commit message: %w", err)
		}
		lg.Info("profile not found, nothing to delete")
		return nil
	}
	if err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}

	if err := cons.Commit(ctx, mqMsg); err != nil {
		return fmt.Errorf("commit message: %w", err)
	}

	lg = lg.With(loglables.Profile, *prof)
	lg.Info("success deleted")

	return nil
}

func (pd *ProfileDeleter) ClientDelete(ctx context.Context) error {
	return ProfileDelete[*ClientProfileDeleteMessage](ctx, pd.ClientConsumer, pd.Profile)
}

func (pd *ProfileDeleter) AdminDelete(ctx context.Context) error {
	return ProfileDelete[*AdminProfileDeleteMessage](ctx, pd.AdminConsumer, pd.Profile)
}

func (pd *ProfileDeleter) Start(ctx context.Context) error {
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return fmt.Errorf("extract logger: %w", err)
	}

	go func() {
		for {
			err := pd.ClientDelete(ctx)
			if err == nil {
				continue
			}

			lg.Error(fmt.Errorf("client delete: %w", err))

			select {
			case <-time.After(pd.CFG.Delay):
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			err := pd.AdminDelete(ctx)
			if err == nil {
				continue
			}

			lg.Error(fmt.Errorf("admin delete: %w", err))

			select {
			case <-time.After(pd.CFG.Delay):
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
