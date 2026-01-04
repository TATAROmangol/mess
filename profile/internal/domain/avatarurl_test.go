package domain_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	avatarmocks "github.com/TATAROmangol/mess/profile/internal/adapter/avatar/mocks"
	"github.com/TATAROmangol/mess/profile/internal/ctxkey"
	"github.com/TATAROmangol/mess/profile/internal/domain"
	"github.com/TATAROmangol/mess/profile/internal/model"
	logmocks "github.com/TATAROmangol/mess/shared/logger/mocks"
	subjmocks "github.com/TATAROmangol/mess/shared/model/mocks"
	"github.com/TATAROmangol/mess/shared/utils"
	"github.com/golang/mock/gomock"
)

func TestDomain_GetAvatarsURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subj := subjmocks.NewMockSubject(ctrl)
	ctx := ctxkey.WithSubject(t.Context(), subj)

	lg := logmocks.NewMockLogger(ctrl)
	ctx = ctxkey.WithLogger(ctx, lg)

	avatar := avatarmocks.NewMockService(ctrl)

	d := domain.Domain{
		Avatar: avatar,
	}

	avatar.EXPECT().GetAvatarURL(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, key string) (string, error) {
		val, _ := strconv.Atoi(key)
		if val%2 == 0 {
			return "", fmt.Errorf("random error for %s", key)
		}
		return "url_for_" + key, nil
	}).Times(100)

	profiles := []*model.Profile{}
	for i := 0; i < 100; i++ {
		profiles = append(profiles, &model.Profile{
			SubjectID: fmt.Sprintf("%v", i),
			AvatarKey: utils.StringPtr(fmt.Sprintf("%v", i)),
		})
	}

	res, errs := d.GetAvatarsURL(ctx, profiles)

	if len(res)+len(errs) != 100 {
		t.Errorf("total results (%d urls + %d errs) != 100", len(res), len(errs))
	}
}
