package workers_test

import(
	"context"
	"github.com/TATAROmangol/mess/profile/internal/adapter/avatar"
	"github.com/TATAROmangol/mess/profile/internal/storage"
	"github.com/TATAROmangol/mess/profile/internal/wokers"
	"testing"
)

func TestAvatarDeleter_Delete(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		cfg     workers.AvatarDeleterConfig
		avatar  avatar.Service
		outbox  storage.AvatarOutbox
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad := workers.NewAvatarDeleter(tt.cfg, tt.avatar, tt.outbox)
			gotErr := ad.Delete(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Delete() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Delete() succeeded unexpectedly")
			}
		})
	}
}
