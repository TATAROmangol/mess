package storage_test

import (
	"chat/internal/storage"
	"errors"
	"testing"
)

func TestStorage_UpdateLastRead(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	lastRead, err := s.LastRead().UpdateLastRead(t.Context(), InitLastReads[0].SubjectID, InitLastReads[0].ChatID, 3, 3)
	if err != nil {
		t.Fatalf("update last read: %v", err)
	}

	if lastRead.MessageNumber != 3 && lastRead.MessageID != 3 {
		t.Fatalf("wait number 3 and id 3, have: %v", lastRead.MessageNumber)
	}

	lastRead, err = s.LastRead().UpdateLastRead(t.Context(), InitLastReads[0].SubjectID, InitLastReads[0].ChatID, 2, 2)
	if err == nil {
		t.Fatalf("want err no rows")
	}
	if !errors.Is(err, storage.ErrNoRows) {
		t.Fatalf("update last read: %v", err)
	}
}

func TestStorage_DeleteLastRead(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	lastRead, err := s.LastRead().DeleteLastRead(t.Context(), InitLastReads[0].SubjectID, InitLastReads[0].ChatID)
	if err != nil {
		t.Fatalf("delete last read: %v", err)
	}

	if lastRead.DeletedAt == nil {
		t.Fatalf("not delete")
	}
}

func TestStorage_GetLastReadByChatIDs(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	lastReads, err := s.LastRead().GetLastReadByChatIDs(t.Context(), InitLastReads[0].SubjectID, []int{InitLastReads[0].ChatID, InitLastReads[1].ChatID})
	if err != nil {
		t.Fatalf("get last read by chat ids: %v", err)
	}

	if len(lastReads) != 2 {
		t.Fatalf("wait len 2, have: %v", len(lastReads))
	}
}

func TestStorage_GetLastReadByChatID(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	lastRead, err := s.LastRead().GetLastReadByChatID(t.Context(), InitLastReads[0].SubjectID, InitLastReads[0].ChatID)
	if err != nil {
		t.Fatalf("get last read by chat id: %v", err)
	}

	if lastRead.ChatID != InitLastReads[0].ChatID || lastRead.SubjectID != InitLastReads[0].SubjectID {
		t.Fatalf("now equal, want %v, have %v", *InitLastReads[0], *lastRead)
	}
}
