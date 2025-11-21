package domain

import (
	"context"
	"fmt"
	"tokenissuer/internal/adapter/identifier"
	"tokenissuer/internal/model"
)

type TokenService interface {
	RefreshTokenPair(ctx context.Context, refreshToken string) (*model.TokenPair, error)
	GetTokenPair(ctx context.Context, code string, redirectURL string) (*model.TokenPair, error)
}

type TokenDomain struct {
	iden identifier.Service
}

func NewTokenDomain(iden identifier.Service) *TokenDomain {
	return &TokenDomain{
		iden: iden,
	}
}

func (td *TokenDomain) RefreshTokenPair(ctx context.Context, refreshToken string) (*model.TokenPair, error) {
	//TODO...
	return nil, nil
}

func (td *TokenDomain) GetTokenPair(ctx context.Context, code string, redirectURL string) (*model.TokenPair, error) {
	_, err := td.iden.ExchangeCode(ctx, code, redirectURL)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}

	//TODO...
	return nil, nil
}
