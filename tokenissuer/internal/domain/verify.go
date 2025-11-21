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

func (v *Verify) VerifyToken(ctx context.Context, accessToken string) (*model.User, error) {
	//TODO...
	return nil, nil
}

func (k *Verify) findKeyByKid(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	k.mu.RLock()
	jwk, ok := k.jwks[kid]
	if ok && time.Since(k.jwksUpdated) < k.jwksTTL {
		k.mu.RUnlock()
		return jwk.GetPublicKey()
	}
	k.mu.RUnlock()

	k.mu.Lock()
	if jwk, ok := k.jwks[kid]; ok && time.Since(k.jwksUpdated) < k.jwksTTL {
		return jwk.GetPublicKey()
	}

	res, err := k.iden.LoadJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("load jwks: %w", err)
	}

	k.jwks = res
	k.mu.Unlock()

	jwk, ok = k.jwks[kid]
	if !ok {
		return nil, fmt.Errorf("kid=%s not found", kid)
	}

	return jwk.GetPublicKey()
}

func (k *Verify) VerifyAccessToken(ctx context.Context, accessToken string) (*model.User, error) {
	token, _, err := k.parser.ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("parse unverified: %w", err)
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("no kid in token header")
	}

	pubKey, err := k.findKeyByKid(ctx, kid)
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
