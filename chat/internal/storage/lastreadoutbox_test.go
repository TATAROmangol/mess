package storage_test

import (
	"testing"

	"github.com/TATAROmangol/mess/chat/internal/storage"
)

func TestStorage_GetLastReadOutbox(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	outbox, err := s.LastReadOutbox().GetLastReadOutbox(t.Context(), 2)
	if err != nil {
		t.Fatalf("get message outbox: %v", err)
	}

	if len(outbox) != 2 {
		t.Fatalf("wait len 2, have: %v", len(outbox))
	}

	if outbox[0].RecipientID != InitLastReadOutboxes[0].RecipientID ||
		outbox[0].ChatID != InitLastReadOutboxes[0].ChatID ||
		outbox[0].SubjectID != InitLastReadOutboxes[0].SubjectID ||
		outbox[0].MessageID != InitLastReadOutboxes[0].MessageID {
		t.Fatalf("not equal")
	}
}

func TestStorage_DeleteLastReadOutbox(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	del, err := s.LastReadOutbox().DeleteLastReadOutbox(t.Context(), []int{1, 2})
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
