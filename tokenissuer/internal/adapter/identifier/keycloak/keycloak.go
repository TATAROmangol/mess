package keycloak

import (
	"time"

	"github.com/go-resty/resty/v2"
)

type Config struct {
	RefreshEndpoint      string        `yaml:"refresh_endpoint"`
	ExchangeCodeEndpoint string        `yaml:"exchange_code_endpoint"`
	JWKSEndpoint         string        `yaml:"jwks_endpoint"`
	ClientID             string        `yaml:"client_id"`
	ClientSecret         string        `yaml:"client_secret"`
	Timeout              time.Duration `yaml:"timeout"`
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
