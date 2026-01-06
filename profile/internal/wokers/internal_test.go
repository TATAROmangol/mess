package workers_test

// import (
// 	"testing"

// 	"github.com/TATAROmangol/mess/profile/internal/domain"
// 	"github.com/TATAROmangol/mess/profile/internal/model"
// 	"github.com/TATAROmangol/mess/shared/utils"
// 	"github.com/golang/mock/gomock"
// )

// func TestDomain_UpdateAvatar_EqualKeys(t *testing.T) {
// 	const SubjID = "subj-1"

// 	ind := domain.NewAvatarIdentifier(SubjID, utils.StringPtr("test"))
// 	key, err := ind.Key()
// 	if err != nil {
// 		t.Fatalf("key: %v", err)
// 	}

// 	profile := &model.Profile{
// 		SubjectID: SubjID,
// 		AvatarKey: &key,
// 	}

// 	env := newTestEnv(t)
// 	defer env.Finish()

// 	env.profile.EXPECT().GetProfileFromSubjectID(env.ctx, SubjID).Return(profile, nil)

// 	err = env.domain.UpdateAvatar(env.ctx, key)
// 	if err != nil {
// 		t.Fatalf("update avatar: %v", err)
// 	}
// }

// func TestDomain_UpdateAvatar_IncorrectPreviousKey(t *testing.T) {
// 	const SubjID = "subj-1"

// 	ind := domain.NewAvatarIdentifier(SubjID, utils.StringPtr("test"))
// 	key, err := ind.Key()
// 	if err != nil {
// 		t.Fatalf("key: %v", err)
// 	}

// 	outboxKey := model.AvatarOutbox{}

// 	profile := &model.Profile{
// 		SubjectID: SubjID,
// 	}

// 	env := newTestEnv(t)
// 	defer env.Finish()

// 	env.profile.EXPECT().GetProfileFromSubjectID(env.ctx, SubjID).Return(profile, nil)
// 	env.outbox.EXPECT().AddKey(env.ctx, SubjID, key).Return(&outboxKey, nil)
// 	env.lg.EXPECT().With(domain.AvatarOutboxKeyLogLabel, outboxKey)
// 	env.lg.EXPECT().Info(gomock.Any())

// 	err = env.domain.UpdateAvatar(env.ctx, SubjID, key)
// 	if err != nil {
// 		t.Fatalf("update avatar: %v", err)
// 	}
// }

// func TestDomain_UpdateAvatar_PreviousKeyNul(t *testing.T) {
// 	const SubjID = "subj-1"

// 	ind := domain.NewAvatarIdentifier(SubjID, nil)
// 	key, err := ind.Key()
// 	if err != nil {
// 		t.Fatalf("key: %v", err)
// 	}

// 	profile := &model.Profile{
// 		SubjectID: SubjID,
// 	}

// 	env := newTestEnv(t)
// 	defer env.Finish()

// 	env.profile.EXPECT().GetProfileFromSubjectID(env.ctx, SubjID).Return(profile, nil)

// 	env.profile.EXPECT().UpdateAvatarKey(env.ctx, SubjID, key).Return(nil)

// 	err = env.domain.UpdateAvatar(env.ctx, key)
// 	if err != nil {
// 		t.Fatalf("update avatar: %v", err)
// 	}
// }

// func TestDomain_UpdateAvatar_AllSteps(t *testing.T) {
// 	const SubjID = "subj-1"

// 	prevInd := domain.NewAvatarIdentifier(SubjID, nil)
// 	prevKey, err := prevInd.Key()
// 	if err != nil {
// 		t.Fatalf("key: %v", err)
// 	}

// 	curInd := domain.NewAvatarIdentifier(SubjID, utils.StringPtr(prevKey))
// 	curKey, err := curInd.Key()
// 	if err != nil {
// 		t.Fatalf("key: %v", err)
// 	}

// 	outboxKey := model.AvatarOutbox{}

// 	profile := &model.Profile{
// 		SubjectID: SubjID,
// 		AvatarKey: &prevKey,
// 	}

// 	env := newTestEnv(t)
// 	defer env.Finish()

// 	env.profile.EXPECT().GetProfileFromSubjectID(env.ctx, SubjID).Return(profile, nil)
// 	env.profile.EXPECT().UpdateAvatarKey(env.ctx, SubjID, curKey).Return(nil)
// 	env.outbox.EXPECT().AddKey(env.ctx, SubjID, prevKey).Return(&outboxKey, nil)
// 	env.lg.EXPECT().With(domain.AvatarOutboxKeyLogLabel, outboxKey)
// 	env.lg.EXPECT().Info(gomock.Any())

// 	err = env.domain.UpdateAvatar(env.ctx, SubjID, curKey)
// 	if err != nil {
// 		t.Fatalf("update avatar: %v", err)
// 	}
// }

// func TestDomain_DeleteProfile(t *testing.T) {
// 	const SubjID = "subj-1"

// 	env := newTestEnv(t)
// 	defer env.Finish()

// 	env.profile.EXPECT().DeleteProfile(env.ctx, SubjID).Return(nil)

// 	if err := env.domain.DeleteProfile(env.ctx, SubjID); err != nil {
// 		t.Fatalf("delete profile: %v", err)
// 	}
// }

// func TestDomain_DeleteOldAvatars(t *testing.T) {
// 	env := newTestEnv(t)
// 	defer env.Finish()

// 	k1 := model.AvatarOutbox{Key: "k1"}
// 	k2 := model.AvatarOutbox{Key: "k2"}
// 	keys := []*model.AvatarOutbox{&k1, &k2}
// 	srcKeys := model.GetAvatarKeys(keys)
// 	env.outbox.EXPECT().GetKeys(env.ctx, domain.DeleteAvatarsLimit).Return(keys, nil)
// 	env.avatar.EXPECT().DeleteObjects(env.ctx, srcKeys).Return(nil)
// 	env.outbox.EXPECT().DeleteKeys(env.ctx, srcKeys).Return(nil)

// 	if err := env.domain.DeleteOldAvatars(env.ctx); err != nil {
// 		t.Errorf("delete old avatars: %v", err)
// 	}
// }
