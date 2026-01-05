package domain

import (
	"context"
	"fmt"

	"github.com/TATAROmangol/mess/profile/internal/ctxkey"
	"github.com/TATAROmangol/mess/profile/internal/loglables"
	"github.com/TATAROmangol/mess/profile/internal/model"
)

func (d *Domain) GetCurrentProfile(ctx context.Context) (*model.Profile, string, error) {
	subj, err := ctxkey.ExtractSubject(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("extract subject: %v", err)
	}

	profile, err := d.Storage.Profile().GetProfileFromSubjectID(ctx, subj.GetSubjectId())
	if err != nil {
		return nil, "", fmt.Errorf("profile get profile from subject id: %v", err)
	}

	avatarURL, err := d.GetAvatarURL(ctx, profile.AvatarKey)
	if err != nil {
		return nil, "", fmt.Errorf("avatar get avatar url: %v", err)
	}

	return profile, avatarURL, nil
}

func (d *Domain) GetProfileFromSubjectID(ctx context.Context, subjID string) (*model.Profile, string, error) {
	profile, err := d.Storage.Profile().GetProfileFromSubjectID(ctx, subjID)
	if err != nil {
		return nil, "", fmt.Errorf("profile get profile from subject id: %v", err)
	}

	avatarURL, err := d.GetAvatarURL(ctx, profile.AvatarKey)
	if err != nil {
		return nil, "", fmt.Errorf("get avatar url: %v", err)
	}

	return profile, avatarURL, nil
}

func (d *Domain) GetProfilesFromAlias(ctx context.Context, alias string, size int, token string) (string, []*model.Profile, map[string]string, error) {
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return "", nil, nil, fmt.Errorf("extract logger: %v", err)
	}

	if size == 0 {
		size = DefaultPageSize
	}

	var nextToken string
	var profiles []*model.Profile

	if token != "" {
		nextToken, profiles, err = d.Storage.Profile().GetProfilesFromAliasWithToken(ctx, alias, token)
		if err != nil {
			return "", nil, nil, fmt.Errorf("profile get profiles from alias with token: %v", err)
		}
	} else {
		nextToken, profiles, err = d.Storage.Profile().GetProfilesFromAlias(ctx, size, Asc, SortLabel, alias)
		if err != nil {
			return "", nil, nil, fmt.Errorf("profile first get profiles from alias: %v", err)
		}
	}

	avatarsURLS, errors := d.GetAvatarsURL(ctx, profiles)
	if len(errors) != 0 {
		lg.Errors("get avatars url", errors)
	}

	return nextToken, profiles, avatarsURLS, nil
}

func (d *Domain) AddProfile(ctx context.Context, alias string) (*model.Profile, error) {
	subj, err := ctxkey.ExtractSubject(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract subject: %v", err)
	}

	profile, err := d.Storage.Profile().AddProfile(ctx, subj.GetSubjectId(), alias)
	if err != nil {
		return nil, fmt.Errorf("profile add profile: %v", err)
	}

	return profile, nil
}

func (d *Domain) UpdateProfileMetadata(ctx context.Context, prevVersion int, alias string) (*model.Profile, string, error) {
	subj, err := ctxkey.ExtractSubject(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("extract subject: %v", err)
	}

	profile, err := d.Storage.Profile().UpdateProfileMetadata(ctx, subj.GetSubjectId(), prevVersion, alias)
	if err != nil {
		return nil, "", fmt.Errorf("profile update profile metadata: %v", err)
	}

	avatarURL, err := d.GetAvatarURL(ctx, profile.AvatarKey)
	if err != nil {
		return nil, "", fmt.Errorf("get avatar url: %v", err)
	}

	return profile, avatarURL, nil
}

func (d *Domain) UploadAvatar(ctx context.Context) (string, error) {
	subj, err := ctxkey.ExtractSubject(ctx)
	if err != nil {
		return "", fmt.Errorf("extract subject: %v", err)
	}

	prof, err := d.Storage.Profile().GetProfileFromSubjectID(ctx, subj.GetSubjectId())
	if err != nil {
		return "", fmt.Errorf("profile get profile from subject id: %v", err)
	}

	key, err := model.NewAvatarIdentifier(prof.SubjectID, prof.AvatarKey).Key()
	if err != nil {
		return "", fmt.Errorf("key: %v", err)
	}

	url, err := d.Avatar.GetUploadURL(ctx, key)
	if err != nil {
		return "", fmt.Errorf("get upload url: %v", err)
	}

	return url, nil
}

func (d *Domain) DeleteAvatar(ctx context.Context) (*model.Profile, string, error) {
	subj, err := ctxkey.ExtractSubject(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("extract subject: %v", err)
	}
	lg, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("extract logger: %v", err)
	}

	prof, err := d.Storage.Profile().GetProfileFromSubjectID(ctx, subj.GetSubjectId())
	if err != nil {
		return nil, "", fmt.Errorf("profile get profile from subject id: %v", err)
	}

	if prof.AvatarKey == nil {
		return prof, "", nil
	}

	s, err := d.Storage.WithTransaction(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("with transaction: %v", err)
	}
	defer s.Rollback()

	prof, err = s.Profile().DeleteAvatarKey(ctx, prof.SubjectID)
	if err != nil {
		return nil, "", fmt.Errorf("profile delete avatar key: %v", err)
	}

	outbox, err := s.AvatarOutbox().AddKey(ctx, prof.SubjectID, *prof.AvatarKey)
	if err != nil {
		return nil, "", fmt.Errorf("avatar key outbox add key: %v", err)
	}
	lg.With(loglables.AvatarOutbox, *outbox)

	if err := s.Commit(); err != nil {
		return nil, "", fmt.Errorf("commit: %v", err)
	}

	lg.Info("add avatar outbox")

	return prof, "", nil
}
