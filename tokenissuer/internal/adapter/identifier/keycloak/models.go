package keycloak

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
)

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
}

func (t *TokenResponse) GetAccessToken() string {
	return t.AccessToken
}

func (t *TokenResponse) GetRefreshToken() string {
	return t.RefreshToken
}

func (t *TokenResponse) GetExpiresIn() int {
	return t.ExpiresIn
}

func (t *TokenResponse) GetRefreshExpiresIn() int {
	return t.RefreshExpiresIn
}

func (t *TokenResponse) GetTokenType() string {
	return t.TokenType
}

type JWKS struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (j JWKS) ToPublicKey() (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(j.N)
	if err != nil {
		return nil, fmt.Errorf("decode n: %w", err)
	}

	eb, err := base64.RawURLEncoding.DecodeString(j.E)
	if err != nil {
		return nil, fmt.Errorf("decode e: %w", err)
	}

	e := big.NewInt(0).SetBytes(eb).Int64()

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nb),
		E: int(e),
	}, nil
}

type User struct {
	ID    string
	Name  string
	Email string
}

func (s *User) GetID() string {
	return s.Name
}

func (s *User) GetName() string {
	return s.Name
}

func (s *User) GetEmail() string {
	return s.Email
}
