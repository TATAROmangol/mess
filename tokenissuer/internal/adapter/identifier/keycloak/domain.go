package keycloak

import (
	"context"
	"fmt"
	"tokenissuer/internal/adapter/identifier"
	"tokenissuer/pkg/jwks"
)

const (
	GrantType    = "authorization_code"
	RefreshToken = "refresh_token"

	RefreshTokenField = "refresh_token"
	GrantTypeField    = "grant_type"
	CodeField         = "code"
	RedirectURIField  = "redirect_uri"
	ClientIDField     = "client_id"
	ClientSecretField = "client_secret"
)

func (k *Keycloak) ExchangeCode(ctx context.Context, code string, redirectURL string) (identifier.TokenPair, error) {
	resp, err := k.client.R().
		SetContext(ctx).
		SetFormData(map[string]string{
			GrantTypeField:    GrantType,
			CodeField:         code,
			RedirectURIField:  redirectURL,
			ClientIDField:     k.cfg.ClientID,
			ClientSecretField: k.cfg.ClientSecret,
		}).
		SetResult(&TokenResponse{}).
		Post(k.cfg.ExchangeCodeEndpoint)

	if err != nil {
		return nil, fmt.Errorf("post: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("response: %s", resp.String())
	}

	return resp.Result().(identifier.TokenPair), nil
}

func (k *Keycloak) Refresh(ctx context.Context, refreshToken string) (identifier.TokenPair, error) {
	resp, err := k.client.R().
		SetContext(ctx).
		SetFormData(map[string]string{
			GrantTypeField:    RefreshToken,
			RefreshTokenField: refreshToken,
			ClientIDField:     k.cfg.ClientID,
			ClientSecretField: k.cfg.ClientSecret,
		}).
		SetResult(&TokenResponse{}).
		Post(k.cfg.RefreshEndpoint)

	if err != nil {
		return nil, fmt.Errorf("post: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("response: %s", resp.String())
	}

	return resp.Result().(identifier.TokenPair), nil
}

func (k *Keycloak) LoadJWKS(ctx context.Context) (map[string]jwks.JWKS, error) {
	res := make(map[string]jwks.JWKS)
	resp, err := k.client.R().
		SetContext(ctx).
		SetResult(&struct {
			Keys []jwks.JWKSToken `json:"keys"`
		}{}).
		Get(k.cfg.JWKSEndpoint)

	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("response: %w", err)
	}

	result := resp.Result().(*struct {
		Keys []jwks.JWKSToken `json:"keys"`
	})

	for _, jwks := range result.Keys {
		res[jwks.Kid] = jwks
	}

	return res, nil
}
