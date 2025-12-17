package avatar

import "context"

type Service interface {
	Upload(ctx context.Context, id string, data []byte, contentType string) (string, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, id string, data []byte, contentType string) (string, error)
}
