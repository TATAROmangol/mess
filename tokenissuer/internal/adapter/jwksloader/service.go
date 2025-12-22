package jwksloader

import (
	"context"

	"github.com/TATAROmangol/mess/shared/jwks"
)

type Service interface {
	LoadJWKS(ctx context.Context) (map[string]jwks.JWKS, error)
}
