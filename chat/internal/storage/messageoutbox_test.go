package storage_test

import (
	"github.com/TATAROmangol/mess/chat/internal/storage"
	"testing"
)

func TestStorage_GetMessageOutbox(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	outbox, err := s.MessageOutbox().GetMessageOutbox(t.Context(), 2)
	if err != nil {
		t.Fatalf("get message outbox: %v", err)
	}

	if len(outbox) != 2 {
		t.Fatalf("wait len 2, have: %v", len(outbox))
	}
}

func TestStorage_DeleteMessageOutbox(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	del, err := s.MessageOutbox().DeleteMessageOutbox(t.Context(), []int{1, 2})
	if err != nil {
		t.Fatalf("delete message outbox: %v", err)
	}
	if len(del) != 2 {
		t.Fatalf("wait len 2, have: %v", len(del))
	}

	for _, dl := range del {
		if dl.DeletedAt == nil {
			t.Fatalf("not delete: %v", *dl)
		}
	}
}
