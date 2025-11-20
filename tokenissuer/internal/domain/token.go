package domain

import (
	"context"
	"tokenissuer/internal/adapter/identifier"
	"tokenissuer/internal/model"
)

type TokenService interface {
}

type TokenVerifier interface {
	VerifyToken(ctx context.Context, accessToken string) (*model.User, error)
}

type TokenDomain struct {
	iden identifier.Service
}

func NewTokenDomain(iden identifier.Service) *TokenDomain {
	return &TokenDomain{
		iden: iden,
	}
}

func (td *TokenDomain) VerifyToken(ctx context.Context, accessToken string) (*model.User, error) {
	return nil, nil
}
