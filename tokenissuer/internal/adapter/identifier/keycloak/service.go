package keycloak

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

	SubClaim      = "sub"
	EmailClaim    = "email"
	UsernameClaim = "preferred_username"
)

func (k *Keycloak) ExchangeCode(code string, redirectURL string) (*TokenResponse, error) {
	resp, err := k.client.R().
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

	return resp.Result().(*TokenResponse), nil
}

func (k *Keycloak) Refresh(refreshToken string) (*TokenResponse, error) {
	resp, err := k.client.R().
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

	return resp.Result().(*TokenResponse), nil
}

func (k *Keycloak) loadJWKS() error {
	resp, err := k.client.R().
		SetResult(&struct {
			Keys []JWKS `json:"keys"`
		}{}).
		Get(k.cfg.JWKSEndpoint)

	if err != nil {
		return fmt.Errorf("request: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("response: %w", err)
	}

	result := resp.Result().(*struct {
		Keys []JWKS `json:"keys"`
	})

	k.jwks = make(map[string]JWKS)
	for _, jwks := range result.Keys {
		k.jwks[jwks.Kid] = jwks
	}

	k.jwksUpdated = time.Now()

	return nil
}

func (k *Keycloak) findKeyByKid(kid string) (*rsa.PublicKey, error) {
	k.mu.RLock()
	jwk, ok := k.jwks[kid]
	if ok && time.Since(k.jwksUpdated) < k.jwksTTL {
		k.mu.RUnlock()
		return jwk.ToPublicKey()
	}
	k.mu.RUnlock()

	k.mu.Lock()
	if jwk, ok := k.jwks[kid]; ok && time.Since(k.jwksUpdated) < k.jwksTTL {
		return jwk.ToPublicKey()
	}

	if err := k.loadJWKS(); err != nil {
		return nil, fmt.Errorf("load jwks: %w", err)
	}
	k.mu.Unlock()

	jwk, ok = k.jwks[kid]
	if !ok {
		return nil, fmt.Errorf("kid=%s not found", kid)
	}

	return jwk.ToPublicKey()
}

func (k *Keycloak) VerifyAccessToken(accessToken string) (*User, error) {
	token, _, err := k.parser.ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("parse unverified: %w", err)
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("no kid in token header")
	}

	pubKey, err := k.findKeyByKid(kid)
	if err != nil {
		return nil, fmt.Errorf("find key by kid: %w", err)
	}

	claims := jwt.MapClaims{}

	_, err = jwt.ParseWithClaims(accessToken, &claims, func(t *jwt.Token) (interface{}, error) {
		return pubKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse with claims: %w", err)
	}

	user := &User{
		ID:    claims[SubClaim].(string),
		Email: claims[EmailClaim].(string),
		Name:  claims[UsernameClaim].(string),
	}

	return user, nil
}
