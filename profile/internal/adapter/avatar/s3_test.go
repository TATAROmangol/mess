package avatar_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/TATAROmangol/mess/profile/internal/adapter/avatar"
	"github.com/TATAROmangol/mess/shared/s3client"
)

var CFG avatar.Config

var TestIDs = []string{"test-file-1", "test-file-1"}

var Content = []byte("test file content")

func TestMain(m *testing.M) {
	cfgClient := s3client.Config{
		Region:          "us-east-1",
		Endpoint:        "http://localhost:9000",
		AccessKeyID:     "profile",
		SecretAccessKey: "profile-secret",
		PathStyle:       true,
	}

	CFG = avatar.Config{
		Client: cfgClient,
		Bucket: "avatar",
	}

	os.Exit(m.Run())
}

func cleanUP(t *testing.T, s avatar.Service) {
	ctx := context.Background()
	err := s.DeleteObjects(ctx, TestIDs)
	if err != nil {
		t.Logf("cleanup failed: %v", err)
	}
}

func uploadDefaultFiles(ctx context.Context, t *testing.T, st avatar.Service) {
	for _, key := range TestIDs {
		uploadURL, err := st.GetUploadURL(ctx, key)
		if err != nil {
			t.Fatalf("GetUploadURL failed: %v", err)
		}
		if uploadURL == "" {
			t.Fatal("expected upload URL to be not empty")
		}

		file := bytes.NewReader(Content)

		req, err := http.NewRequest(http.MethodPut, uploadURL, file)
		if err != nil {
			t.Fatalf("failed to create upload request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("upload failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			t.Fatalf("unexpected upload status: %d", resp.StatusCode)
		}
	}
}

func TestGetUploadURLAndGetAvatarURL(t *testing.T) {
	st, err := avatar.New(t.Context(), CFG)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer cleanUP(t, st)

	uploadDefaultFiles(t.Context(), t, st)

	for _, key := range TestIDs {
		url, err := st.GetAvatarURL(t.Context(), key)
		if err != nil {
			t.Fatalf("failed to get avatar URL: %v", err)
		}
		if url == "" {
			t.Fatal("expected avatar URL to be not empty")
		}

		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("failed to GET avatar: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("avatar not accessible, status: %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if !bytes.Equal(body, Content) {
			t.Fatal("avatar content does not match uploaded content")
		}
	}
}

func TestDeleteObjects(t *testing.T) {
	st, err := avatar.New(t.Context(), CFG)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	uploadDefaultFiles(t.Context(), t, st)
	err = st.DeleteObjects(t.Context(), TestIDs)
	if err != nil {
		t.Fatalf("DeleteObjects failed: %v", err)
	}
}
