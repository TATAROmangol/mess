package profile

import (
	"context"

	"github.com/TATAROmangol/mess/profile/internal/model"
)

type Service interface {
	WithTransaction(ctx context.Context) (ServiceTransaction, error)
	AddProfile(ctx context.Context, subjID string, alias string, avatarURL string) (*model.Profile, error)
	GetProfileFromSubjectID(ctx context.Context, subjID string) (*model.Profile, error)
	UpdateProfile(ctx context.Context, profile *model.Profile) (*model.Profile, error)
	GetProfilesFromAlias(ctx context.Context, size int, asc bool, sortLabel Label, alias string) (string, []*model.Profile, error)
	GetProfilesFromAliasWithToken(ctx context.Context, token, alias string) (string, []*model.Profile, error)
	DeleteProfileFromSubjectID(ctx context.Context, subjID string) error
}

type ServiceTransaction interface {
	Service
	Commit() error
	Rollback() error
}
