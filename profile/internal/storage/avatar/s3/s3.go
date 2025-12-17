package s3

import (
	"context"
	"fmt"
	"profile/pkg/s3client"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Client       s3client.Config `yaml:"client"`
	PublicBucket string          `yaml:"public_bucket"`
}

type Storage struct {
	cfg Config
	c   *s3.Client
}

func New(ctx context.Context, cfg Config) (*Storage, error) {
	client, err := s3client.New(ctx, cfg.Client)
	if err != nil {
		return nil, fmt.Errorf("create s3 client: %w", err)
	}

	return &Storage{
		cfg: cfg,
		c:   client,
	}, nil
}
