package keycloak

import (
	"context"
	"fmt"
	"tokenissuer/pkg/jwks"
)

func (k *Keycloak) LoadJWKS(ctx context.Context) (map[string]jwks.JWKS, error) {
	res := make(map[string]jwks.JWKS)
	resp, err := k.client.R().
		SetContext(ctx).
		SetResult(&struct {
			Keys []jwks.JWKSImpl `json:"keys"`
		}{}).
		Get(k.cfg.JWKSEndpoint)

	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("response: %w", err)
	}

	result := resp.Result().(*struct {
		Keys []jwks.JWKSImpl `json:"keys"`
	})

	for _, jwks := range result.Keys {
		res[jwks.Kid] = jwks
	}

	return res, nil
}
