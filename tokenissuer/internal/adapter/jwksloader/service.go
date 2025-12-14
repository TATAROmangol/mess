package jwksloader

import (
	"context"
	"tokenissuer/pkg/jwks"
)

type Service interface {
	LoadJWKS(ctx context.Context) (map[string]jwks.JWKS, error)
}
