package keycloak

import (
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	RefreshEndpoint      string        `json:"refresh_endpoint"`
	ExchangeCodeEndpoint string        `json:"exchange_code_endpoint"`
	JWKSEndpoint         string        `json:"jwks_endpoint"`
	ClientID             string        `json:"client_id"`
	ClientSecret         string        `json:"client_secret"`
	Timeout              time.Duration `json:"timeout"`
	JWKSTTL              time.Duration `json:"jwks_ttl"`
}

type Keycloak struct {
	cfg    Config
	client *resty.Client

	jwks        map[string]JWKS
	jwksUpdated time.Time
	jwksTTL     time.Duration

	parser *jwt.Parser
	mu     sync.RWMutex
}

func NewKeycloak(cfg Config) *Keycloak {
	client := resty.New()
	client.SetTimeout(cfg.Timeout)

	return &Keycloak{
		cfg:     cfg,
		client:  client,
		jwks:    make(map[string]JWKS),
		parser:  jwt.NewParser(jwt.WithoutClaimsValidation()),
		mu:      sync.RWMutex{},
		jwksTTL: cfg.JWKSTTL,
	}
}
