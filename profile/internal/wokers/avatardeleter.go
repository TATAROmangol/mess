package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/TATAROmangol/mess/profile/internal/adapter/avatar"
	"github.com/TATAROmangol/mess/profile/internal/ctxkey"
	"github.com/TATAROmangol/mess/profile/internal/loglables"
	"github.com/TATAROmangol/mess/profile/internal/model"
	"github.com/TATAROmangol/mess/profile/internal/storage"
)

const (
	DeleteAvatarsLimit = 100
)

type AvatarDeleterConfig struct {
	runAt        time.Duration `yaml:"run_at"`
	intervalDays int           `yaml:"interval_days"`
}

type AvatarDeleter struct {
	cfg    AvatarDeleterConfig
	avatar avatar.Service
	outbox storage.AvatarOutbox
}

func NewAvatarDeleter(cfg AvatarDeleterConfig, avatar avatar.Service, outbox storage.AvatarOutbox) *AvatarDeleter {
	return &AvatarDeleter{
		cfg:    cfg,
		avatar: avatar,
		outbox: outbox,
	}
}

func (ad *AvatarDeleter) Delete(ctx context.Context) error {
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return fmt.Errorf("extract logger: %v", err)
	}

	keys, err := ad.outbox.GetKeys(ctx, DeleteAvatarsLimit)
	if err != nil {
		return fmt.Errorf("outbox get keys: %v", err)
	}

	if err = ad.avatar.DeleteObjects(ctx, model.GetOutboxKeys(keys)); err != nil {
		return fmt.Errorf("avatar delete objects: %v", err)
	}

	outboxes, err := ad.outbox.DeleteKeys(ctx, model.GetOutboxKeys(keys))
	if err != nil {
		return fmt.Errorf("outbox delete keys: %v", err)
	}
	lg.With(loglables.DeletedAvatarKeys, model.GetOutboxKeys(outboxes))
	lg.Info("success delete")

	return nil
}

func (ad *AvatarDeleter) delayUntilRunAt() time.Duration {
	now := time.Now().UTC()

	runAtToday := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		time.UTC,
	).Add(ad.cfg.runAt)

	if now.Before(runAtToday) {
		return runAtToday.Sub(now)
	}

	return runAtToday.Add(24 * time.Hour).Sub(now)
}

func (ad *AvatarDeleter) Start(ctx context.Context) error {
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return fmt.Errorf("extract logger: %v", err)
	}

	go func() {
		delay := ad.delayUntilRunAt()

		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-timer.C:
			if err := ad.Delete(ctx); err != nil {
				lg.Error(fmt.Errorf("delete old avatars: %v", err))
			}
		case <-ctx.Done():
			return
		}

		ticker := time.NewTicker(time.Duration(ad.cfg.intervalDays) * 24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := ad.Delete(ctx); err != nil {
					lg.Error(fmt.Errorf("delete old avatars: %v", err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
