package domain

import (
	"context"
	"fmt"
	"tokenissuer/internal/adapter/identifier"
)

type TokenService interface {
	RefreshTokenPair(ctx context.Context, refreshToken string) (identifier.TokenPair, error)
	GetTokenPair(ctx context.Context, code string, redirectURL string) (identifier.TokenPair, error)
}

type TokenDomain struct {
	iden identifier.Service
}

func NewTokenDomain(iden identifier.Service) *TokenDomain {
	return &TokenDomain{
		iden: iden,
	}
}

func (td *TokenDomain) RefreshTokenPair(ctx context.Context, refreshToken string) (identifier.TokenPair, error) {
	pair, err := td.iden.Refresh(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh: %w", err)
	}

	return pair, nil
}

func (td *TokenDomain) GetTokenPair(ctx context.Context, code string, redirectURL string) (identifier.TokenPair, error) {
	pair, err := td.iden.ExchangeCode(ctx, code, redirectURL)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}

	return pair, nil
}
