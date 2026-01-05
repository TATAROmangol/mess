package domain

import (
	"context"

	"github.com/TATAROmangol/mess/profile/internal/adapter/avatar"
	"github.com/TATAROmangol/mess/profile/internal/model"
	"github.com/TATAROmangol/mess/profile/internal/storage"
)

type Service interface {
	GetCurrentProfile(ctx context.Context) (*model.Profile, string, error)

	GetProfileFromSubjectID(ctx context.Context, subjID string) (*model.Profile, string, error)
	GetProfilesFromAlias(ctx context.Context, alias string, size int, token string) (string, []*model.Profile, map[string]string, error)

	AddProfile(ctx context.Context, alias string) (*model.Profile, error)

	UpdateProfileMetadata(ctx context.Context, prevVersion int, alias string) (*model.Profile, string, error)
	UploadAvatar(ctx context.Context) (string, error)

	DeleteAvatar(ctx context.Context) (*model.Profile, string, error)
}

const (
	DefaultPageSize = 100
	Asc             = true
	SortLabel       = storage.ProfileAliasLabel
)

type Domain struct {
	Storage storage.Service
	Avatar  avatar.Service
}

func New(storage storage.Service, avatar avatar.Service) Service {
	return &Domain{
		Storage: storage,
		Avatar:  avatar,
	}
}
