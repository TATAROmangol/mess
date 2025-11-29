package config

import (
	"fmt"
	"os"
	"tokenissuer/internal/adapter/identifier/keycloak"
	"tokenissuer/internal/service"
	"tokenissuer/internal/transport/grpc"
	"tokenissuer/internal/transport/rest"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Keycloak      keycloak.Config      `yaml:"keycloak"`
	HTTP          rest.Config          `yaml:"rest"`
	GRPC          grpc.Config          `yaml:"grpc"`
	VerifyService service.VerifyConfig `yaml:"verify_config"`
}

func Load(path string) (*Config, error) {
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
