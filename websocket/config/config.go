package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/TATAROmangol/mess/shared/auth/keycloak"
	"github.com/TATAROmangol/mess/websocket/internal/transport"
	"github.com/TATAROmangol/mess/websocket/internal/worker"
	"github.com/goccy/go-yaml"
)

type Config struct {
	Keycloak      keycloak.Config            `yaml:"keycloak"`
	MessageWorker worker.MessageWorkerConfig `yaml:"message_worker"`
	HTTP          transport.HTTPConfig       `yaml:"http"`
	WSConfig      transport.WSHandlerConfig  `yaml:"ws_config"`
}

func LoadConfig() (*Config, error) {
	var configPath = flag.String("config", "", "path to config")
	flag.Parse()

	path := *configPath
	if path == "" {
		path = os.Getenv("CONFIG")
	}

	if path == "" {
		panic("Config path is not set. Pass -config or set CONFIG")
	}

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
