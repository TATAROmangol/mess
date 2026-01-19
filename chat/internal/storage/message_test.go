package storage_test

import (
	"github.com/TATAROmangol/mess/chat/internal/storage"
	"testing"
)

func TestStorage_GetLastMessagesByChatsID(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	messages, err := s.Message().GetLastMessagesByChatsID(t.Context(), []int{InitChats[0].ID, InitChats[1].ID})
	if err != nil {
		t.Fatalf("get last messages byt chats ids: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("wait len 2, have: %v", len(messages))
	}

	expected := map[int]string{
		1: "test-content-2",
		2: "test-content-1",
	}

	for _, m := range messages {
		want, ok := expected[m.ChatID]
		if !ok {
			t.Errorf("unexpected chat id: %v", m.ChatID)
			continue
		}
		if m.Content != want {
			t.Errorf("chat %v: expected last message content %q, got %q", m.ChatID, want, m.Content)
		}
	}
}

func TestStorage_GetMessagesByChatID(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	filter := &storage.PaginationFilterIntLastID{
		Limit:     10,
		Asc:       true,
		SortLabel: storage.MessageCreatedAtLabel,
	}

	messages, err := s.Message().GetMessagesByChatID(t.Context(), InitChats[0].ID, filter)
	if err != nil {
		t.Fatalf("get messages by chat id: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("wait len 2, have: %v", len(messages))
	}

	if messages[0].Content != InitMessages[0].Content || messages[1].Content != InitMessages[1].Content {
		t.Fatalf("not equal, wait: %v, %v, have: %v, %v", *InitMessages[0], *InitMessages[1], *messages[0], *messages[1])
	}
}

func TestStorage_UpdateMessageContent(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	message, err := s.Message().UpdateMessageContent(t.Context(), InitMessages[0].ID, "halo", InitMessages[0].Version)
	if err != nil {
		t.Fatalf("update message content: %v", err)
	}

	if message.Content != "halo" || message.Version-1 != InitMessages[0].Version {
		t.Fatalf("now wait message: %v", *message)
	}
}

func TestStorage_DeleteMessagesChatID(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	messages, err := s.Message().DeleteMessagesChatID(t.Context(), InitMessages[0].ID)
	if err != nil {
		t.Fatalf("delete messages: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("wait len 2, have: %v", len(messages))
	}

	for _, mes := range messages {
		if mes.DeletedAt == nil {
			t.Fatalf("not delete: %v", *mes)
		}
	}
}

func TestStorage_GetMessageByID(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	mess, err := s.Message().GetMessageByID(t.Context(), InitMessages[0].ID)
	if err != nil {
		t.Fatalf("get message by id: %v", err)
	}

	if mess.ID != InitMessages[0].ID || mess.Content != InitMessages[0].Content {
		t.Fatalf("not equal, want: %v. have: %v", *InitMessages[0], *mess)
	}
}
