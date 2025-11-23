package jwks

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
)

type JWKS interface {
	GetPublicKey() (*rsa.PublicKey, error)
}

type JWKSImpl struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (j JWKSImpl) GetPublicKey() (*rsa.PublicKey, error) {
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
