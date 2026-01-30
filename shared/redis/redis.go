package redis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr     string        `yaml:"addr"`
	DB       int           `yaml:"db"`
	Password string        `yaml:"password"`
	Timeout  time.Duration `yaml:"timeout"`
}

func NewClient(cfg Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:        cfg.Addr,
		DB:          cfg.DB,
		Password:    cfg.Password,
		DialTimeout: cfg.Timeout,
	})
}
