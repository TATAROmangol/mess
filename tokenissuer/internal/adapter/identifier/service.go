package identifier

import (
	"context"
	"tokenissuer/pkg/jwks"
)

type TokenPair interface {
	GetAccessToken() string
	GetRefreshToken() string
	GetExpiresIn() int
	GetRefreshExpiresIn() int
	GetTokenType() string
}

type CodeExchanger interface {
	ExchangeCode(ctx context.Context, code string, redirectURL string) (TokenPair, error)
}

type TokenRefresher interface {
	Refresh(ctx context.Context, refreshToken string) (TokenPair, error)
}

type JWKSLoader interface {
	LoadJWKS(ctx context.Context) (map[string]jwks.JWKS, error)
}

type Service interface {
	CodeExchanger
	TokenRefresher
	JWKSLoader
}
