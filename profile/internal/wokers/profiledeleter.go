package workers

// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	"github.com/TATAROmangol/mess/profile/internal/adapter/avatar"
// 	"github.com/TATAROmangol/mess/profile/internal/ctxkey"
// 	"github.com/TATAROmangol/mess/profile/internal/storage"
// )

// type ProfileDeleterConfig struct {
// 	runAt        time.Duration `yaml:"run_at"`
// 	intervalDays int           `yaml:"interval_days"`
// }

// type ProfileDeleter struct {
// 	cfg     AvatarDeleterConfig
// 	avatar  avatar.Service
// 	profile storage.Profile
// }

// func NewProfileDeleter(cfg AvatarDeleterConfig, avatar avatar.Service, outbox storage.AvatarOutbox) *AvatarDeleter {
// 	return &AvatarDeleter{
// 		cfg:    cfg,
// 		avatar: avatar,
// 		outbox: outbox,
// 	}
// }

// func (pd *ProfileDeleter) Delete(ctx context.Context) error {
// 	prof, err := d.Storage.Profile().DeleteProfile(ctx, subjID)
// 	if err != nil {
// 		return fmt.Errorf("profile delete profile: %v", err)
// 	}

// 	if prof.AvatarKey == nil {
// 		return prof, nil
// 	}

// 	key, err := d.Storage.AvatarKeyOutbox().AddKey(ctx, prof.SubjectID, *prof.AvatarKey)
// 	if err != nil {
// 		return fmt.Errorf("add key: %v", err)
// 	}

// 	return nil
// }

// func (pd *ProfileDeleter) Start(ctx context.Context) error {
// 	lg, err := ctxkey.ExtractLogger(ctx)
// 	if err != nil {
// 		return fmt.Errorf("extract logger: %v", err)
// 	}

// 	go func() {
// 		delay := pd.delayUntilRunAt()

// 		timer := time.NewTimer(delay)
// 		defer timer.Stop()

// 		select {
// 		case <-timer.C:
// 			if err := pd.Delete(ctx); err != nil {
// 				lg.Error(fmt.Errorf("delete old avatars: %v", err))
// 			}
// 		case <-ctx.Done():
// 			return
// 		}

// 		ticker := time.NewTicker(time.Duration(pd.cfg.intervalDays) * 24 * time.Hour)
// 		defer ticker.Stop()

// 		for {
// 			select {
// 			case <-ticker.C:
// 				if err := pd.Delete(ctx); err != nil {
// 					lg.Error(fmt.Errorf("delete old avatars: %v", err))
// 				}
// 			case <-ctx.Done():
// 				return
// 			}
// 		}
// 	}()

// 	return nil
// }
