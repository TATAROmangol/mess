package tokenissuer

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

type AdapterConfig struct {
	AuthURL  string `yaml:"url"`
	ClientID string `yaml:"client_id"`
	Login    string `yaml:"login"`
	Password string `yaml:"password"`
}

type TokenIssuerConfig struct {
	VerifyCodeEndpoint string `yaml:"verify_code_endpoint"`
	RefreshEndpoint    string `yaml:"refresh_endpoint"`
	VerifyGrpcAddress  string `yaml:"verify_grpc_address"`
	HTTPPort           int    `yaml:"http_port"`
}

type Config struct {
	Adapter AdapterConfig     `yaml:"adapter"`
	Issuer  TokenIssuerConfig `yaml:"token_issuer"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config file: %w", err)
	}

	return &cfg, nil
}
