package avatar

import "context"

type Service interface {
	GetUploadURL(ctx context.Context, key string, prevKey string) (string, error)
	GetAvatarURL(ctx context.Context, key string) (string, error)
	DeleteObjects(ctx context.Context, keys []string) ([]string, error)
}
