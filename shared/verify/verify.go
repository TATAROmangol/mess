package verify

import (
	"fmt"
	"strings"
	"time"

	"github.com/1ocknight/mess/shared/logger"
	"github.com/1ocknight/mess/shared/model"
	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
)

const (
	SubClaim      = "sub"
	EmailClaim    = "email"
)

type Config struct {
	JWKSEndpoint string `json:"jwks_endpoint"`
}

type Verify struct {
	jwks *keyfunc.JWKS
}

func New(cfg Config, lg logger.Logger) (*Verify, error) {
	jwks, err := keyfunc.Get(cfg.JWKSEndpoint, keyfunc.Options{
		RefreshInterval: time.Minute * 10,
		RefreshTimeout:  time.Second * 10,
		RefreshErrorHandler: func(err error) {
			lg.Error(fmt.Errorf("refresh: %w", err))
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	return &Verify{
		jwks: jwks,
	}, nil
}

func (k *Verify) Verify(src string) (model.Subject, error) {
	parts := strings.Split(src, " ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token len: %s", src)
	}
	tokenStr := parts[1]

	token, err := jwt.Parse(tokenStr, k.jwks.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	subj := &model.SubjectIMPL{
		SubjectID: claims[SubClaim].(string),
		Email:     claims[EmailClaim].(string),
	}

	return subj, nil
}
