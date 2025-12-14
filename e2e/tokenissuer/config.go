package tokenissuer

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type AdapterConfig struct {
	AuthURL       string        `yaml:"url"`
	ClientID      string        `yaml:"client_id"`
	ClientSecret  string        `yaml:"client_secret"`
	Login         string        `yaml:"login"`
	Password      string        `yaml:"password"`
	SubjectID     string        `yaml:"subject_id"`
	TokenDuration time.Duration `yaml:"token_duration"`
}

type TokenIssuerConfig struct {
	VerifyGrpcAddress string `yaml:"verify_grpc_address"`
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
