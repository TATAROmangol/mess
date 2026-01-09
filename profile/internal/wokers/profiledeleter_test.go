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
	mq "github.com/TATAROmangol/mess/shared/messagequeue/mocks"
	"github.com/golang/mock/gomock"
)

func TestProfileDeleter_ClientDelete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := mq.NewMockConsumer(ctrl)
	store := storagemocks.NewMockProfile(ctrl)

	lg := loggermocks.NewMockLogger(ctrl)
	ctx := ctxkey.WithLogger(context.Background(), lg)

	msg := &workers.ClientProfileDeleteMessage{SubjectID: "user123"}
	raw, _ := json.Marshal(msg)

	mqMsg := mq.NewMockMessage(ctrl)
	mqMsg.EXPECT().Value().Return(raw)

	consumer.EXPECT().ReadMessage(ctx).Return(mqMsg, nil)
	store.EXPECT().DeleteProfile(ctx, "user123").Return(&model.Profile{SubjectID: "user123"}, nil)
	lg.EXPECT().With(loglables.Profile, gomock.Any()).Return(lg)
	lg.EXPECT().Info("success deleted")

	pd := &workers.ProfileDeleter{
		ClientConsumer: consumer,
		Profile:        store,
	}

	err := pd.ClientDelete(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProfileDeleter_AdminDelete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := mq.NewMockConsumer(ctrl)
	store := storagemocks.NewMockProfile(ctrl)

	lg := loggermocks.NewMockLogger(ctrl)
	ctx := ctxkey.WithLogger(context.Background(), lg)

	msg := &workers.AdminProfileDeleteMessage{SubjectID: "user123"}
	raw, _ := json.Marshal(msg)

	mqMsg := mq.NewMockMessage(ctrl)
	mqMsg.EXPECT().Value().Return(raw)

	consumer.EXPECT().ReadMessage(ctx).Return(mqMsg, nil)
	store.EXPECT().DeleteProfile(ctx, "user123").Return(&model.Profile{SubjectID: "user123"}, nil)
	lg.EXPECT().With(loglables.Profile, gomock.Any()).Return(lg)
	lg.EXPECT().Info("success deleted")

	pd := &workers.ProfileDeleter{
		AdminConsumer: consumer,
		Profile:       store,
	}

	err := pd.AdminDelete(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
