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
	RunHourUTC int           `yaml:"run_hour_utc"`
	Interval   time.Duration `yaml:"interval"`
}

type AvatarDeleter struct {
	CFG    AvatarDeleterConfig
	Avatar avatar.Service
	Outbox storage.AvatarOutbox
}

func NewAvatarDeleter(cfg AvatarDeleterConfig, avatar avatar.Service, outbox storage.AvatarOutbox) *AvatarDeleter {
	return &AvatarDeleter{
		CFG:    cfg,
		Avatar: avatar,
		Outbox: outbox,
	}
}

func (ad *AvatarDeleter) Delete(ctx context.Context) error {
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return fmt.Errorf("extract logger: %w", err)
	}

	for {
		keys, err := ad.Outbox.GetKeys(ctx, DeleteAvatarsLimit)
		if err != nil {
			return fmt.Errorf("outbox get keys: %w", err)
		}
		if len(keys) == 0 {
			lg.Info("no more avatars to delete")
			break
		}

		if err = ad.Avatar.DeleteObjects(ctx, model.GetOutboxKeys(keys)); err != nil {
			return fmt.Errorf("avatar delete objects: %w", err)
		}

		outboxes, err := ad.Outbox.DeleteKeys(ctx, model.GetOutboxKeys(keys))
		if err != nil {
			return fmt.Errorf("outbox delete keys: %w", err)
		}
		lg = lg.With(loglables.DeletedAvatarKeys, model.GetOutboxKeys(outboxes))
		lg.Info("success delete")
	}

	return nil
}

func (ad *AvatarDeleter) delayUntilRunAt() time.Duration {
	if ad.CFG.Interval == -1 {
		return time.Duration(0)
	}

	now := time.Now().UTC()

	runAt := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		ad.CFG.RunHourUTC,
		0,
		0,
		0,
		time.UTC,
	)

	if !runAt.After(now) {
		runAt = runAt.Add(24 * time.Hour)
	}

	return time.Until(runAt)
}

func (ad *AvatarDeleter) Start(ctx context.Context) error {
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return fmt.Errorf("extract logger: %w", err)
	}

	go func() {
		delay := ad.delayUntilRunAt()

		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-timer.C:
			if err := ad.Delete(ctx); err != nil {
				lg.Error(fmt.Errorf("delete old avatars: %w", err))
			}
		case <-ctx.Done():
			return
		}

		ticker := time.NewTicker(ad.CFG.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := ad.Delete(ctx); err != nil {
					lg.Error(fmt.Errorf("delete old avatars: %w", err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
