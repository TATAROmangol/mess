package workers_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/TATAROmangol/mess/profile/internal/ctxkey"
	"github.com/TATAROmangol/mess/profile/internal/loglables"
	"github.com/TATAROmangol/mess/profile/internal/model"
	storagemocks "github.com/TATAROmangol/mess/profile/internal/storage/mocks"
	workers "github.com/TATAROmangol/mess/profile/internal/wokers"
	loggermocks "github.com/TATAROmangol/mess/shared/logger/mocks"
	messagequeuemocks "github.com/TATAROmangol/mess/shared/messagequeue/mocks"
	"github.com/golang/mock/gomock"
)

func TestAvatarUploader_Upload_SuccessUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := messagequeuemocks.NewMockConsumer(ctrl)
	storage := storagemocks.NewMockService(ctrl)
	profileRepo := storagemocks.NewMockProfile(ctrl)
	outboxRepo := storagemocks.NewMockAvatarOutbox(ctrl)
	tx := storagemocks.NewMockServiceTransaction(ctrl)
	lg := loggermocks.NewMockLogger(ctrl)

	ctx := ctxkey.WithLogger(context.Background(), lg)

	au := workers.AvatarUploader{
		Consumer: consumer,
		Storage:  storage,
	}

	subjectID := "123"
	ind := model.NewAvatarIdentifier(subjectID, nil)
	fKey, _ := ind.Key()

	sInd := model.NewAvatarIdentifier(subjectID, &fKey)
	sKey, _ := sInd.Key()

	// формируем сообщение с новой структурой
	msg := workers.AvatarUploaderMessage{
		Records: []struct {
			S3 struct {
				Object struct {
					Key string `json:"key"`
				} `json:"object"`
			} `json:"s3"`
		}{
			{
				S3: struct {
					Object struct {
						Key string `json:"key"`
					} `json:"object"`
				}{
					Object: struct {
						Key string `json:"key"`
					}{Key: sKey},
				},
			},
		},
	}

	msgBytes, _ := json.Marshal(msg)
	mqMsg := messagequeuemocks.NewMockMessage(ctrl)

	profileBefore := &model.Profile{
		SubjectID: subjectID,
		AvatarKey: &fKey,
	}
	profileAfter := &model.Profile{
		SubjectID: subjectID,
		AvatarKey: &sKey,
	}
	outbox := &model.AvatarOutbox{
		SubjectID: subjectID,
		Key:       fKey,
	}

	consumer.EXPECT().ReadMessage(ctx).Return(mqMsg, nil)
	mqMsg.EXPECT().Value().Return(msgBytes)

	storage.EXPECT().Profile().Return(profileRepo)
	profileRepo.EXPECT().GetProfileFromSubjectID(ctx, subjectID).Return(profileBefore, nil)

	storage.EXPECT().WithTransaction(ctx).Return(tx, nil)
	tx.EXPECT().Profile().Return(profileRepo)
	profileRepo.EXPECT().UpdateAvatarKey(ctx, subjectID, msg.Key()).Return(profileAfter, nil)
	lg.EXPECT().With(loglables.Profile, *profileAfter).Return(lg)

	storage.EXPECT().AvatarOutbox().Return(outboxRepo)
	outboxRepo.EXPECT().AddKey(ctx, subjectID, fKey).Return(outbox, nil)
	lg.EXPECT().With(loglables.AvatarOutbox, *outbox).Return(lg)

	tx.EXPECT().Commit().Return(nil)
	tx.EXPECT().Rollback().Return(nil)

	consumer.EXPECT().Commit(ctx, mqMsg).Return(nil)
	lg.EXPECT().Info(gomock.Any())

	if err := au.Upload(ctx); err != nil {
		t.Fatalf("upload: %v", err)
	}
}

func TestAvatarUploader_Upload_SuccessSkipOld(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := messagequeuemocks.NewMockConsumer(ctrl)
	storage := storagemocks.NewMockService(ctrl)
	profileRepo := storagemocks.NewMockProfile(ctrl)
	outboxRepo := storagemocks.NewMockAvatarOutbox(ctrl)
	lg := loggermocks.NewMockLogger(ctrl)

	ctx := ctxkey.WithLogger(context.Background(), lg)

	au := workers.AvatarUploader{
		Consumer: consumer,
		Storage:  storage,
	}

	subjectID := "123"
	ind := model.NewAvatarIdentifier(subjectID, &subjectID)
	fKey, _ := ind.Key()

	// создаём сообщение
	msg := workers.AvatarUploaderMessage{
		Records: []struct {
			S3 struct {
				Object struct {
					Key string `json:"key"`
				} `json:"object"`
			} `json:"s3"`
		}{
			{
				S3: struct {
					Object struct {
						Key string `json:"key"`
					} `json:"object"`
				}{
					Object: struct {
						Key string `json:"key"`
					}{Key: fKey},
				},
			},
		},
	}

	msgBytes, _ := json.Marshal(msg)
	mqMsg := messagequeuemocks.NewMockMessage(ctrl)

	profileBefore := &model.Profile{
		SubjectID: subjectID,
		AvatarKey: nil,
	}
	outbox := &model.AvatarOutbox{
		SubjectID: subjectID,
		Key:       msg.Key(),
	}

	consumer.EXPECT().ReadMessage(ctx).Return(mqMsg, nil)
	mqMsg.EXPECT().Value().Return(msgBytes)
	storage.EXPECT().Profile().Return(profileRepo)
	profileRepo.EXPECT().GetProfileFromSubjectID(ctx, subjectID).Return(profileBefore, nil)
	storage.EXPECT().AvatarOutbox().Return(outboxRepo)
	outboxRepo.EXPECT().AddKey(ctx, subjectID, msg.Key()).Return(outbox, nil)
	lg.EXPECT().With(loglables.AvatarOutbox, *outbox).Return(lg)
	consumer.EXPECT().Commit(ctx, mqMsg).Return(nil)
	lg.EXPECT().Info(gomock.Any())

	if err := au.Upload(ctx); err != nil {
		t.Fatalf("upload: %v", err)
	}
}
