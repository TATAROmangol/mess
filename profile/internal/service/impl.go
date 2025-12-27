package service

import (
	"context"

	"github.com/TATAROmangol/mess/profile/internal/model"
	"github.com/TATAROmangol/mess/profile/internal/storage"
)

type IMPL struct {
	s *storage.Service
}

func New(s *storage.Service) *IMPL {
	return &IMPL{
		s: s,
	}
}

func (i *IMPL) GetCurrentProfile(ctx context.Context) (*model.Profile, error) {
	return nil, nil
}
