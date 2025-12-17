package storage

import (
	"profile/internal/storage/avatar"
	"profile/internal/storage/profile"
)

type Service interface {
	Profile() profile.Service
	Avatar() avatar.Service
}

type IMPL struct {
	Profile profile.Service
	Avatar  avatar.Service
}

func New(profileService profile.Service, avatarService avatar.Service) Service {
	return &IMPL{
		Profile: profileService,
		Avatar:  avatarService,
	}
}
