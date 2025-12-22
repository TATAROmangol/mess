package avatar_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/TATAROmangol/mess/profile/internal/storage/avatar"
	"github.com/TATAROmangol/mess/shared/s3client"
)

var CFG avatar.Config

var TestID = "subject_id"

func TestMain(m *testing.M) {
	cfgClient := s3client.Config{
		Region:          "us-east-1",
		Endpoint:        "http://localhost:9000",
		AccessKeyID:     "avatar-backend",
		SecretAccessKey: "avatarback",
		PathStyle:       true,
	}

	CFG = avatar.Config{
		Client:       cfgClient,
		PublicBucket: "avatar",
	}

	os.Exit(m.Run())
}

func cleanUP(t *testing.T, s *avatar.Storage) {
	err := s.Delete(t.Context(), TestID)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
}

func checkData(t *testing.T, data []byte, contentType string) {
	url := fmt.Sprintf("%v/%v/%v", CFG.Client.Endpoint, CFG.PublicBucket, TestID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("http.Get(): %v", err)
	}
	defer resp.Body.Close()

	// 1. HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// 2. Content-Type
	gotCT := resp.Header.Get("Content-Type")
	if gotCT != contentType {
		t.Errorf("Content-Type = %q, want %q", gotCT, contentType)
	}

	// 3. Body
	gotBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if !bytes.Equal(gotBody, data) {
		t.Fatalf("body mismatch: got %q, want %q", gotBody, data)
	}
}

func TestStorage_Upload(t *testing.T) {
	ctx := context.Background()

	s, err := avatar.New(ctx, CFG)
	if err != nil {
		t.Fatalf("avatar.New(): %v", err)
	}
	defer cleanUP(t, s)

	data := []byte("fake image bytes")
	contentType := "image/png"

	url, err := s.Upload(ctx, TestID, data, contentType)
	if err != nil {
		t.Fatalf("Upload(): %v", err)
	}

	wantURL := fmt.Sprintf("%v/%v/%v", CFG.Client.Endpoint, CFG.PublicBucket, TestID)

	if url != wantURL {
		t.Fatalf("Upload() url = %s, want %s", url, wantURL)
	}

	checkData(t, data, contentType)
}

func TestStorage_Delete(t *testing.T) {
	ctx := context.Background()

	s, err := avatar.New(ctx, CFG)
	if err != nil {
		t.Fatalf("avatar.New(): %v", err)
	}

	data := []byte("to be deleted")
	contentType := "text/plain"

	_, err = s.Upload(ctx, TestID, data, contentType)
	if err != nil {
		t.Fatalf("Upload(): %v", err)
	}

	err = s.Delete(ctx, TestID)
	if err != nil {
		t.Fatalf("Delete(): %v", err)
	}

	url := fmt.Sprintf("%v/%v/%v", CFG.Client.Endpoint, CFG.PublicBucket, TestID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("http.Get(): %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected file to be deleted, but got status 200")
	}
}

func TestStorage_Update(t *testing.T) {
	ctx := context.Background()

	s, err := avatar.New(ctx, CFG)
	if err != nil {
		t.Fatalf("avatar.New(): %v", err)
	}
	defer cleanUP(t, s)

	dataV1 := []byte("old data")
	ctV1 := "text/plain"

	dataV2 := []byte("new data")
	ctV2 := "application/octet-stream"

	_, err = s.Upload(ctx, TestID, dataV1, ctV1)
	if err != nil {
		t.Fatalf("Upload(): %v", err)
	}

	url, err := s.Update(ctx, TestID, dataV2, ctV2)
	if err != nil {
		t.Fatalf("Update(): %v", err)
	}

	wantURL := fmt.Sprintf("%v/%v/%v", CFG.Client.Endpoint, CFG.PublicBucket, TestID)
	if url != wantURL {
		t.Fatalf("Update() url = %s, want %s", url, wantURL)
	}

	checkData(t, dataV2, ctV2)
}
