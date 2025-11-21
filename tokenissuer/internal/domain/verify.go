package domain

import (
	"context"
	"crypto/rsa"
	"fmt"
	"sync"
	"time"
	"tokenissuer/internal/adapter/identifier"
	"tokenissuer/internal/model"
	"tokenissuer/pkg/jwks"

	"github.com/golang-jwt/jwt/v5"
)

const (
	KidHeader     = "kid"
	SubClaim      = "sub"
	EmailClaim    = "email"
	UsernameClaim = "preferred_username"
)

type VerifyService interface {
	VerifyToken(ctx context.Context, accessToken string) (*model.User, error)
}

type Verify struct {
	iden identifier.JWKSLoader

	jwks        map[string]jwks.JWKS
	jwksUpdated time.Time
	jwksTTL     time.Duration

	parser *jwt.Parser
	mu     sync.RWMutex
}

func NewVerify(iden identifier.Service) *Verify {
	return &Verify{
		iden: iden,
	}
}

func (v *Verify) findKeyByKid(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	jwk, ok := v.jwks[kid]
	if ok && time.Since(v.jwksUpdated) < v.jwksTTL {
		v.mu.RUnlock()
		return jwk.GetPublicKey()
	}
	v.mu.RUnlock()

	v.mu.Lock()
	if jwk, ok := v.jwks[kid]; ok && time.Since(v.jwksUpdated) < v.jwksTTL {
		return jwk.GetPublicKey()
	}

	res, err := v.iden.LoadJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("load jwks: %w", err)
	}

	v.jwks = res
	v.mu.Unlock()

	jwk, ok = v.jwks[kid]
	if !ok {
		return nil, fmt.Errorf("kid=%s not found", kid)
	}

	return jwk.GetPublicKey()
}

func (v *Verify) VerifyToken(ctx context.Context, accessToken string) (*model.User, error) {
	token, _, err := v.parser.ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("parse unverified: %w", err)
	}

	kid, ok := token.Header[KidHeader].(string)
	if !ok {
		return nil, fmt.Errorf("no kid in token header")
	}

	pubKey, err := v.findKeyByKid(ctx, kid)
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

	user := &model.User{
		ID:    claims[SubClaim].(string),
		Name:  claims[UsernameClaim].(string),
		Email: claims[EmailClaim].(string),
	}

	return user, nil
}
