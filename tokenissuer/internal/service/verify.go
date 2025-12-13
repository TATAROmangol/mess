package service

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
	"golang.org/x/sync/singleflight"
)

const (
	KidHeader     = "kid"
	SubClaim      = "sub"
	EmailClaim    = "email"
	UsernameClaim = "preferred_username"
	BearerType    = "Bearer"

	JWKSUpdateKey = "jwks_update"
)

type VerifyConfig struct {
	JwksRateLimit time.Duration `yaml:"jwks_rate_limit"`
}

type Verify interface {
	VerifyToken(ctx context.Context, typeToken, accessToken string) (*model.User, error)
}

type VerifyImpl struct {
	iden identifier.JWKSLoader

	jwks            map[string]jwks.JWKS
	jwksLastUpdated time.Time
	jwksRateLimit   time.Duration

	parser *jwt.Parser

	mu *sync.RWMutex
	sf singleflight.Group
}

func NewVerifyImpl(ctx context.Context, iden identifier.JWKSLoader, vCfg VerifyConfig) (*VerifyImpl, error) {
	jwks, err := iden.LoadJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("load jwks: %w", err)
	}

	return &VerifyImpl{
		iden: iden,

		jwks:            jwks,
		jwksLastUpdated: time.Now(),
		jwksRateLimit:   vCfg.JwksRateLimit,

		parser: jwt.NewParser(),
		mu:     &sync.RWMutex{},
		sf:     singleflight.Group{},
	}, nil
}

func (v *VerifyImpl) updateJWKSKeys(ctx context.Context) error {
	v.mu.RLock()
	if time.Since(v.jwksLastUpdated) < v.jwksRateLimit {
		v.mu.RUnlock()
		return nil
	}
	v.mu.RUnlock()

	_, err, _ := v.sf.Do(JWKSUpdateKey, func() (interface{}, error) {
		res, err := v.iden.LoadJWKS(ctx)
		if err != nil {
			return false, fmt.Errorf("load jwks: %w", err)
		}

		v.mu.Lock()
		v.jwks = res
		v.jwksLastUpdated = time.Now()
		v.mu.Unlock()

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("singleflight do: %w", err)
	}

	return nil
}

func (v *VerifyImpl) findKeyByKid(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	err := v.updateJWKSKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("update jwks keys: %w", err)
	}

	v.mu.RLock()
	jwk, ok := v.jwks[kid]
	v.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("kid=%s not found", kid)
	}

	return jwk.GetPublicKey()
}

func (v *VerifyImpl) VerifyToken(ctx context.Context, typeToken, accessToken string) (*model.User, error) {
	if typeToken != BearerType {
		return nil, fmt.Errorf("invalid token type")
	}

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
