package s3

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *Storage) Upload(ctx context.Context, id string, data []byte, contentType string) (string, error) {
	key := id

	_, err := s.c.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.cfg.PublicBucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload object to s3: %w", err)
	}

	return fmt.Sprintf("%s/%s/%s", s.cfg.Client.Endpoint, s.cfg.PublicBucket, id), nil
}

func (s *Storage) Delete(ctx context.Context, id string) error {
	key := id

	_, err := s.c.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.cfg.PublicBucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("delete object from s3: %w", err)
	}
	return nil
}

func (s *Storage) Update(ctx context.Context, id string, data []byte, contentType string) (string, error) {
	url, err := s.Upload(ctx, id, data, contentType)
	if err != nil {
		return "", fmt.Errorf("upload object in s3: %w", err)
	}

	return url, nil
}
