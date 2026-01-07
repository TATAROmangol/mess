package keycloak

import (
	"fmt"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/TATAROmangol/mess/shared/logger"
	"github.com/TATAROmangol/mess/shared/model"
	"github.com/golang-jwt/jwt/v4"
)

const (
	SubClaim      = "sub"
	EmailClaim    = "email"
	UsernameClaim = "preferred_username"
)

type Config struct {
	JWKSEndpoint string `json:"jwks_endpoint"`
}

type Keycloak struct {
	jwks *keyfunc.JWKS
}

func New(cfg Config, lg logger.Logger) (*Keycloak, error) {
	jwks, err := keyfunc.Get(cfg.JWKSEndpoint, keyfunc.Options{
		RefreshInterval: time.Minute * 10,
		RefreshTimeout:  time.Second * 10,
		RefreshErrorHandler: func(err error) {
			lg.Error(fmt.Errorf("refresh: %v", err))
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get: %v", err)
	}

	return &Keycloak{
		jwks: jwks,
	}, nil
}

func (k *Keycloak) Verify(src string) (model.Subject, error) {
	parts := strings.Split(src, " ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token len: %v", src)
	}
	tokenStr := parts[1]

	token, err := jwt.Parse(tokenStr, k.jwks.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("parse: %v", err)
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
