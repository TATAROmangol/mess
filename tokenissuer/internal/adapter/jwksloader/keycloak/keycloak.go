package keycloak

import (
	"time"

	"github.com/go-resty/resty/v2"
)

type Config struct {
	JWKSEndpoint string        `yaml:"jwks_endpoint"`
	Timeout      time.Duration `yaml:"timeout"`
}

type Keycloak struct {
	cfg    Config
	client *resty.Client
}

func NewKeycloak(cfg Config) *Keycloak {
	client := resty.New()
	client.SetTimeout(cfg.Timeout)

	return &Keycloak{
		cfg:    cfg,
		client: client,
	}
}
