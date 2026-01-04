package domain_test

import (
	"fmt"
	"testing"

	"github.com/TATAROmangol/mess/profile/internal/domain"
	"github.com/TATAROmangol/mess/profile/internal/model"
	"github.com/TATAROmangol/mess/shared/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestDomain_GetCurrentProfile_Success(t *testing.T) {
	const (
		subjID    = "subj-1"
		avatarKey = "avatar-key"
		avatarURL = "https://avatar.url/img.png"
	)

	profile := &model.Profile{
		SubjectID: subjID,
		AvatarKey: utils.StringPtr(avatarKey),
	}

	env := newTestEnv(t)
	defer env.Finish()

	env.subj.EXPECT().GetSubjectId().Return(subjID)
	env.profile.EXPECT().GetProfileFromSubjectID(env.ctx, subjID).Return(profile, nil)
	env.avatar.EXPECT().GetAvatarURL(env.ctx, avatarKey).Return(avatarURL, nil)

	p, url, err := env.domain.GetCurrentProfile(env.ctx)
	require.NoError(t, err)
	require.Equal(t, profile, p)
	require.Equal(t, avatarURL, url)
}

func TestDomain_GetProfileFromSubjectID_Success(t *testing.T) {
	const (
		subjID    = "subj-1"
		avatarKey = "avatar-key"
		avatarURL = "https://avatar.url/img.png"
	)

	profile := &model.Profile{
		SubjectID: subjID,
		AvatarKey: utils.StringPtr(avatarKey),
	}

	env := newTestEnv(t)
	defer env.Finish()

	env.profile.EXPECT().GetProfileFromSubjectID(env.ctx, subjID).Return(profile, nil)
	env.avatar.EXPECT().GetAvatarURL(env.ctx, avatarKey).Return(avatarURL, nil)

	p, url, err := env.domain.GetProfileFromSubjectID(env.ctx, subjID)
	require.NoError(t, err)
	require.Equal(t, profile, p)
	require.Equal(t, avatarURL, url)
}

func TestDomain_GetProfilesFromAlias_Success(t *testing.T) {
	const (
		alias   = "alias-1"
		size    = 2
		avatar1 = "avatar-1"
		avatar2 = "avatar-2"
	)

	profiles := []*model.Profile{
		{SubjectID: "subj-1", AvatarKey: utils.StringPtr(avatar1)},
		{SubjectID: "subj-2", AvatarKey: utils.StringPtr(avatar2)},
	}

	avatarsURLS := map[string]string{
		"subj-1": "https://avatar.url/1.png",
		"subj-2": "https://avatar.url/2.png",
	}

	env := newTestEnv(t)
	defer env.Finish()

	env.profile.EXPECT().
		GetProfilesFromAlias(env.ctx, size, domain.Asc, domain.SortLabel, alias).
		Return("next-token", profiles, nil)

	env.avatar.EXPECT().GetAvatarURL(env.ctx, avatar1).Return("https://avatar.url/1.png", nil)
	env.avatar.EXPECT().GetAvatarURL(env.ctx, avatar2).Return("https://avatar.url/2.png", nil)

	nextToken, profs, avatars, err := env.domain.GetProfilesFromAlias(env.ctx, alias, size, "")
	require.NoError(t, err)
	require.Equal(t, "next-token", nextToken)
	require.Equal(t, profiles, profs)
	require.Equal(t, avatarsURLS, avatars)
}

func TestDomain_GetProfilesFromAlias_Errors(t *testing.T) {
	const (
		alias   = "alias-1"
		size    = 2
		avatar1 = "avatar-1"
		avatar2 = "avatar-2"
	)

	profiles := []*model.Profile{
		{SubjectID: "subj-1", AvatarKey: utils.StringPtr(avatar1)},
		{SubjectID: "subj-2", AvatarKey: utils.StringPtr(avatar2)},
	}

	avatarsURLS := map[string]string{
		"subj-1": "https://avatar.url/1.png",
	}

	env := newTestEnv(t)
	defer env.Finish()

	env.profile.EXPECT().
		GetProfilesFromAlias(env.ctx, size, domain.Asc, domain.SortLabel, alias).
		Return("next-token", profiles, nil)

	env.avatar.EXPECT().GetAvatarURL(env.ctx, avatar1).Return("https://avatar.url/1.png", nil)
	env.avatar.EXPECT().GetAvatarURL(env.ctx, avatar2).Return("", fmt.Errorf("err"))

	env.lg.EXPECT().Errors(gomock.Any(), gomock.Any())

	nextToken, profs, avatars, err := env.domain.GetProfilesFromAlias(env.ctx, alias, size, "")
	require.NoError(t, err)
	require.Equal(t, "next-token", nextToken)
	require.Equal(t, profiles, profs)
	require.Equal(t, avatarsURLS, avatars)
}

func TestDomain_AddProfile_Success(t *testing.T) {
	const (
		subjID = "subj-1"
		alias  = "alias-1"
	)

	profile := &model.Profile{SubjectID: subjID}

	env := newTestEnv(t)
	defer env.Finish()

	env.subj.EXPECT().GetSubjectId().Return(subjID)

	env.profile.EXPECT().AddProfile(env.ctx, subjID, alias).Return(profile, nil)

	p, err := env.domain.AddProfile(env.ctx, alias)
	require.NoError(t, err)
	require.Equal(t, profile, p)
}

func TestDomain_UpdateProfileMetadata_Success(t *testing.T) {
	const (
		subjID       = "subj-1"
		prevVersion  = 1
		alias        = "alias-1"
		avatarKeyStr = "avatar-key"
		avatarURL    = "https://avatar.url/img.png"
	)

	profile := &model.Profile{
		SubjectID: subjID,
		AvatarKey: utils.StringPtr(avatarKeyStr),
	}

	env := newTestEnv(t)
	defer env.Finish()

	env.subj.EXPECT().GetSubjectId().Return(subjID)
	env.profile.EXPECT().UpdateProfileMetadata(env.ctx, subjID, prevVersion, alias).Return(profile, nil)
	env.avatar.EXPECT().GetAvatarURL(env.ctx, avatarKeyStr).Return(avatarURL, nil)

	p, url, err := env.domain.UpdateProfileMetadata(env.ctx, prevVersion, alias)
	require.NoError(t, err)
	require.Equal(t, profile, p)
	require.Equal(t, avatarURL, url)
}

func TestDomain_LoadAvatar_Success_EmptyKey(t *testing.T) {
	const SubjID = "subj-1"
	const URL = "url"

	profile := &model.Profile{
		SubjectID: SubjID,
	}

	env := newTestEnv(t)
	defer env.Finish()

	env.subj.EXPECT().GetSubjectId().Return(SubjID)
	env.profile.EXPECT().GetProfileFromSubjectID(env.ctx, SubjID).Return(profile, nil)
	env.avatar.EXPECT().GetUploadURL(env.ctx, gomock.Any()).Return(URL, nil)

	url, err := env.domain.UploadAvatar(env.ctx)
	require.NoError(t, err)
	require.Equal(t, URL, url)
}

func TestDomain_LoadAvatar_Success_EmptyPrevKey(t *testing.T) {
	const SubjID = "subj-1"
	const URL = "url"

	ind := domain.NewAvatarIdentifier(SubjID, nil)
	key, err := ind.Key()
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	profile := &model.Profile{
		SubjectID: SubjID,
		AvatarKey: &key,
	}

	env := newTestEnv(t)
	defer env.Finish()

	env.subj.EXPECT().GetSubjectId().Return(SubjID)
	env.profile.EXPECT().GetProfileFromSubjectID(env.ctx, SubjID).Return(profile, nil)
	env.avatar.EXPECT().GetUploadURL(env.ctx, gomock.Any()).Return(URL, nil)

	url, err := env.domain.UploadAvatar(env.ctx)
	require.NoError(t, err)
	require.Equal(t, URL, url)
}

func TestDomain_LoadAvatar_Success(t *testing.T) {
	const SubjID = "subj-1"
	const URL = "url"

	sKey := domain.NewAvatarIdentifier(SubjID, utils.StringPtr("prev-key"))
	src, err := sKey.Key()
	if err != nil {
		t.Fatalf("token: %v", err)
	}

	profile := &model.Profile{
		SubjectID: SubjID,
		AvatarKey: &src,
	}

	env := newTestEnv(t)
	defer env.Finish()

	env.subj.EXPECT().GetSubjectId().Return(SubjID)
	env.profile.EXPECT().GetProfileFromSubjectID(env.ctx, SubjID).Return(profile, nil)
	env.avatar.EXPECT().GetUploadURL(env.ctx, gomock.Any()).Return(URL, nil)

	url, err := env.domain.UploadAvatar(env.ctx)
	require.NoError(t, err)
	require.Equal(t, URL, url)
}

func TestDomain_DeleteAvatar_SuccessWithLogger(t *testing.T) {
	const subjID = "subj-1"

	profile := &model.Profile{
		SubjectID: subjID,
		AvatarKey: utils.StringPtr("avatar-key"),
	}

	outboxKey := &model.AvatarKeyOutbox{
		SubjectID: subjID,
		Key:       "test",
	}

	env := newTestEnv(t)
	defer env.Finish()

	env.subj.EXPECT().GetSubjectId().Return(subjID)

	env.profile.EXPECT().
		GetProfileFromSubjectID(env.ctx, subjID).
		Return(profile, nil)

	env.profile.EXPECT().DeleteAvatarKey(env.ctx, subjID).Return(nil)
	env.outbox.EXPECT().AddKey(env.ctx, subjID, *profile.AvatarKey).Return(outboxKey, nil)

	env.lg.EXPECT().With(domain.AvatarOutboxKeyLogLabel, *outboxKey).Return(env.lg)
	env.lg.EXPECT().Info(gomock.Any())

	err := env.domain.DeleteAvatar(env.ctx)
	require.NoError(t, err)
}
