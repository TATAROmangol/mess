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
	ProfileSVC profile.Service
	AvatarSVC  avatar.Service
}

func New(profileService profile.Service, avatarService avatar.Service) Service {
	return &IMPL{
		ProfileSVC: profileService,
		AvatarSVC:  avatarService,
	}
}

func (s *IMPL) Profile() profile.Service {
	return s.ProfileSVC
}

func (s *IMPL) Avatar() avatar.Service {
	return s.AvatarSVC
}
