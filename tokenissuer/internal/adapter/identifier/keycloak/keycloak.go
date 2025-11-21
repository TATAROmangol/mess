package keycloak

import (
	"time"

	"github.com/go-resty/resty/v2"
)

type Config struct {
	RefreshEndpoint      string        `json:"refresh_endpoint"`
	ExchangeCodeEndpoint string        `json:"exchange_code_endpoint"`
	JWKSEndpoint         string        `json:"jwks_endpoint"`
	ClientID             string        `json:"client_id"`
	ClientSecret         string        `json:"client_secret"`
	Timeout              time.Duration `json:"timeout"`
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
