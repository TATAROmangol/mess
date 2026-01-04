package domain

import (
	"context"
	"fmt"

	"github.com/TATAROmangol/mess/profile/internal/ctxkey"
	"github.com/TATAROmangol/mess/profile/internal/model"
	"github.com/TATAROmangol/mess/shared/utils"
)

const (
	DeleteAvatarsLimit = 100
)

func (d *Domain) UpdateAvatar(ctx context.Context, subjID string, avatarKey string) error {
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return fmt.Errorf("extract logger: %v", err)
	}

	ind, err := ParseAvatarKey(avatarKey)
	if err != nil {
		return fmt.Errorf("parse avatar key token: %v", err)
	}

	profile, err := d.Storage.Profile().GetProfileFromSubjectID(ctx, subjID)
	if err != nil {
		return fmt.Errorf("profile get profile from subject id: %v", err)
	}

	if !utils.StringPtrEqual(ind.PreviousKey, profile.AvatarKey) {
		outboxKey, err := d.Storage.AvatarKeyOutbox().AddKey(ctx, subjID, avatarKey)
		if err != nil {
			return fmt.Errorf("avatar key outbox add key: %v", err)
		}
		lg.With(AvatarOutboxKeyLogLabel, *outboxKey)
		lg.Info("add avatar outbox key")
		return nil
	}

	tx, err := d.Storage.WithTransaction(ctx)
	if err != nil {
		return fmt.Errorf("with transaction: %v", err)
	}
	defer tx.Rollback()

	if err = tx.Profile().UpdateAvatarKey(ctx, subjID, avatarKey); err != nil {
		return fmt.Errorf("profile update avatar key: %v", err)
	}

	if profile.AvatarKey == nil {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit: %v", err)
		}
		return nil
	}

	outboxKey, err := d.Storage.AvatarKeyOutbox().AddKey(ctx, subjID, *profile.AvatarKey)
	if err != nil {
		return fmt.Errorf("avatar key outbox add key: %v", err)
	}
	lg.With(AvatarOutboxKeyLogLabel, *outboxKey)
	lg.Info("add avatar outbox key")

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %v", err)
	}

	return nil
}

func (d *Domain) DeleteProfile(ctx context.Context, subjID string) error {
	if err := d.Storage.Profile().DeleteProfile(ctx, subjID); err != nil {
		return fmt.Errorf("profile delete profile: %v", err)
	}

	return nil
}

func (d *Domain) DeleteOldAvatars(ctx context.Context) error {
	keys, err := d.Storage.AvatarKeyOutbox().GetKeys(ctx, DeleteAvatarsLimit)
	if err != nil {
		return fmt.Errorf("outbox get keys: %v", err)
	}

	srcKeys := model.GetAvatarKeys(keys)

	if err = d.Avatar.DeleteObjects(ctx, srcKeys); err != nil {
		return fmt.Errorf("avatar delete objects: %v", err)
	}

	if err = d.Storage.AvatarKeyOutbox().DeleteKeys(ctx, srcKeys); err != nil {
		return fmt.Errorf("outbox delete keys: %v", err)
	}

	return nil
}
