package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/TATAROmangol/mess/profile/internal/adapter/avatar"
	"github.com/TATAROmangol/mess/profile/internal/transport"
	workers "github.com/TATAROmangol/mess/profile/internal/wokers"
	"github.com/TATAROmangol/mess/shared/auth/keycloak"
	"github.com/TATAROmangol/mess/shared/postgres"
	"github.com/goccy/go-yaml"
)

type Config struct {
	MigrationsPath string                       `yaml:"migrations_path"`
	Postgres       postgres.Config              `yaml:"postgres"`
	S3             avatar.Config                `yaml:"s3"`
	HTTP           transport.Config             `yaml:"http"`
	Keycloak       keycloak.Config              `yaml:"keycloak"`
	AvatarDeleter  workers.AvatarDeleterConfig  `yaml:"avatar_deleter"`
	AvatarUploader workers.AvatarUploaderConfig `yaml:"avatar_uploader"`
	ProfileDeleter workers.ProfileDeleterConfig `yaml:"profile_deleter"`
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
