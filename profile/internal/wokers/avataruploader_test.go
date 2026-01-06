package workers_test

import(
	"context"
	"github.com/TATAROmangol/mess/profile/internal/storage"
	"github.com/TATAROmangol/mess/profile/internal/wokers"
	"testing"
)

func TestAvatarUploader_Upload(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		cfg     workers.AvatarUploaderConfig
		storage storage.Service
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			au := workers.NewAvatarUploader(tt.cfg, tt.storage)
			gotErr := au.Upload(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Upload() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Upload() succeeded unexpectedly")
			}
		})
	}
}
