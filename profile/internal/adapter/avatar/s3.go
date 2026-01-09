package avatar

import (
	"context"
	"fmt"
	"time"

	"github.com/TATAROmangol/mess/shared/s3client"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	Client          s3client.Config `yaml:"client"`
	Bucket          string          `yaml:"bucket"`
	PresignDuration time.Duration   `yaml:"presign_duration"`
}

type S3 struct {
	cfg Config
	c   *s3.Client
	p   *s3.PresignClient
}

func New(ctx context.Context, cfg Config) (Service, error) {
	client, err := s3client.New(ctx, cfg.Client)
	if err != nil {
		return nil, fmt.Errorf("create s3 client: %w", err)
	}

	p := s3.NewPresignClient(client)

	return &S3{
		cfg: cfg,
		c:   client,
		p:   p,
	}, nil
}

func (s *S3) GetUploadURL(ctx context.Context, key string) (string, error) {
	req, err := s.p.PresignPutObject(ctx,
		&s3.PutObjectInput{
			Bucket: &s.cfg.Bucket,
			Key:    &key,
		},
		s3.WithPresignExpires(s.cfg.PresignDuration),
	)
	if err != nil {
		return "", fmt.Errorf("presign upload part: %w", err)
	}

	return req.URL, nil
}

func (s *S3) GetAvatarURL(ctx context.Context, key string) (string, error) {
	req, err := s.p.PresignGetObject(ctx,
		&s3.GetObjectInput{
			Bucket: &s.cfg.Bucket,
			Key:    &key,
		},
		s3.WithPresignExpires(s.cfg.PresignDuration),
	)

	if err != nil {
		return "", fmt.Errorf("presign get object: %w", err)
	}

	return req.URL, nil
}

func (s *S3) DeleteObjects(ctx context.Context, keys []string) error {
	objects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = types.ObjectIdentifier{Key: aws.String(key)}
	}

	_, err := s.c.DeleteObjects(ctx,
		&s3.DeleteObjectsInput{
			Bucket: &s.cfg.Bucket,
			Delete: &types.Delete{
				Objects: objects,
				Quiet:   aws.Bool(false),
			},
		},
	)

	if err != nil {
		return fmt.Errorf("delete objects")
	}

	return nil
}
