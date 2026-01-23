package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/TATAROmangol/mess/chat/internal/transport"
	"github.com/TATAROmangol/mess/chat/internal/worker"
	"github.com/TATAROmangol/mess/shared/auth/keycloak"
	"github.com/TATAROmangol/mess/shared/postgres"
	"github.com/goccy/go-yaml"
)

type Config struct {
	MigrationsPath string           `yaml:"migrations_path"`
	Postgres       postgres.Config  `yaml:"postgres"`
	HTTP           transport.Config `yaml:"http"`

	MessageWorker  worker.MessageWorkerConfig `json:"message_worker"`
	LastReadWorker worker.LastReadConfig      `json:"last_read_worker"`

	Keycloak keycloak.Config `yaml:"keycloak"`
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
