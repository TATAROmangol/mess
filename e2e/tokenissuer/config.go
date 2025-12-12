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
	Scope    string `yaml:"scope"`
}

type Config struct {
	Adapter            AdapterConfig `yaml:"adapter"`
	VerifyCodeEndpoint string        `yaml:"verify_code_endpoint"`
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
