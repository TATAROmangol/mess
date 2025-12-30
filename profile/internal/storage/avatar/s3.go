package avatar

import (
	"context"
	"fmt"

	"github.com/TATAROmangol/mess/shared/s3client"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	Client s3client.Config `yaml:"client"`
	Bucket string          `yaml:"public_bucket"`
}

type Storage struct {
	cfg Config
	c   *s3.Client
	p   *s3.PresignClient
}

func New(ctx context.Context, cfg Config) (*Storage, error) {
	client, err := s3client.New(ctx, cfg.Client)
	if err != nil {
		return nil, fmt.Errorf("create s3 client: %w", err)
	}

	p := s3.NewPresignClient(client)

	return &Storage{
		cfg: cfg,
		c:   client,
		p:   p,
	}, nil
}

func (s *Storage) GetUploadURL(ctx context.Context, key string) (string, error) {
	req, err := s.p.PresignPutObject(ctx,
		&s3.PutObjectInput{
			Bucket: &s.cfg.Bucket,
			Key:    &key,
		},
	)
	if err != nil {
		return "", fmt.Errorf("presign upload part: %v", err)
	}

	return req.URL, nil
}

func (s *Storage) GetAvatarURL(ctx context.Context, key string) (string, error) {
	req, err := s.p.PresignGetObject(ctx,
		&s3.GetObjectInput{
			Bucket: &s.cfg.Bucket,
			Key:    &key,
		},
	)

	if err != nil {
		return "", fmt.Errorf("presign get object: %v", err)
	}

	return req.URL, nil
}

// returns non deleted keys
func (s *Storage) DeleteObjects(ctx context.Context, keys []string) ([]string, error) {
	objects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = types.ObjectIdentifier{Key: aws.String(key)}
	}

	resp, err := s.c.DeleteObjects(ctx,
		&s3.DeleteObjectsInput{
			Bucket: &s.cfg.Bucket,
			Delete: &types.Delete{
				Objects: objects,
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("delete objects")
	}

	notDeleted := make([]string, 0, len(resp.Errors))
	for _, errObj := range resp.Errors {
		notDeleted = append(notDeleted, *errObj.Key)
	}

	return notDeleted, nil
}
