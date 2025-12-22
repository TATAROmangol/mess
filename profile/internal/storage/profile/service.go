package profile

import (
	"context"

	"github.com/TATAROmangol/mess/profile/internal/model"
)

type Service interface {
	AddProfile(ctx context.Context, profile *model.Profile) error
	GetProfileFromSubjectID(ctx context.Context, subjID string) (*model.Profile, error)
	UpdateProfile(ctx context.Context, profile *model.Profile) error
	GetProfilesFromAliasWithInfo(ctx context.Context, size int, asc bool, sortLabel Label, alias string)
	GetProfilesFromAliasWithToken(ctx context.Context, token, alias string) (string, []*model.Profile, error)
	DeleteProfileFromSubjectID(ctx context.Context, subjID string) error
}
