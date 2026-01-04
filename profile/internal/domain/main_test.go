package domain_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	avatarmocks "github.com/TATAROmangol/mess/profile/internal/adapter/avatar/mocks"
	"github.com/TATAROmangol/mess/profile/internal/ctxkey"
	"github.com/TATAROmangol/mess/profile/internal/domain"
	storagemocks "github.com/TATAROmangol/mess/profile/internal/storage/mocks"
	logmocks "github.com/TATAROmangol/mess/shared/logger/mocks"
	subjmocks "github.com/TATAROmangol/mess/shared/model/mocks"
)

type TestEnv struct {
	ctrl    *gomock.Controller
	domain  domain.Service
	ctx     context.Context
	storage *storagemocks.MockService
	profile *storagemocks.MockProfile
	outbox  *storagemocks.MockAvatarKeyOutbox
	avatar  *avatarmocks.MockService
	tx      *storagemocks.MockServiceTransaction
	subj    *subjmocks.MockSubject
	lg      *logmocks.MockLogger
}

func newTestEnv(t *testing.T) *TestEnv {
	ctrl := gomock.NewController(t)

	storage := storagemocks.NewMockService(ctrl)
	profile := storagemocks.NewMockProfile(ctrl)
	outbox := storagemocks.NewMockAvatarKeyOutbox(ctrl)
	tx := storagemocks.NewMockServiceTransaction(ctrl)

	storage.EXPECT().Profile().Return(profile).AnyTimes()
	storage.EXPECT().AvatarKeyOutbox().Return(outbox).AnyTimes()
	storage.EXPECT().WithTransaction(gomock.Any()).Return(tx, nil).AnyTimes()
	tx.EXPECT().Profile().Return(profile).AnyTimes()
	tx.EXPECT().AvatarKeyOutbox().Return(outbox).AnyTimes()
	tx.EXPECT().Commit().Return(nil).AnyTimes()
	tx.EXPECT().Rollback().Return(fmt.Errorf("test")).AnyTimes()

	avatar := avatarmocks.NewMockService(ctrl)

	subj := subjmocks.NewMockSubject(ctrl)
	ctx := ctxkey.WithSubject(t.Context(), subj)

	lg := logmocks.NewMockLogger(ctrl)
	ctx = ctxkey.WithLogger(ctx, lg)

	d := domain.New(storage, avatar)

	return &TestEnv{
		ctrl:    ctrl,
		domain:  d,
		ctx:     ctx,
		storage: storage,
		profile: profile,
		outbox:  outbox,
		avatar:  avatar,
		tx:      tx,
		subj:    subj,
		lg:      lg,
	}
}

func (te *TestEnv) Finish() {
	te.ctrl.Finish()
}
