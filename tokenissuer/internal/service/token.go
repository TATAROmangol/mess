package service

import (
	"context"
	"fmt"
	"tokenissuer/internal/adapter/identifier"
)

type Token interface {
	RefreshTokenPair(ctx context.Context, refreshToken string) (identifier.TokenPair, error)
	GetTokenPair(ctx context.Context, code string, redirectURL string) (identifier.TokenPair, error)
}

type TokenImpl struct {
	iden identifier.Service
}

func NewTokenImpl(iden identifier.Service) *TokenImpl {
	return &TokenImpl{
		iden: iden,
	}
}

func (td *TokenImpl) RefreshTokenPair(ctx context.Context, refreshToken string) (identifier.TokenPair, error) {
	pair, err := td.iden.Refresh(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh: %w", err)
	}

	return pair, nil
}

func (td *TokenImpl) GetTokenPair(ctx context.Context, code string, redirectURL string) (identifier.TokenPair, error) {
	pair, err := td.iden.ExchangeCode(ctx, code, redirectURL)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}

	return pair, nil
}
