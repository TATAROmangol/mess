package storage_test

import (
	"github.com/TATAROmangol/mess/chat/internal/storage"
	"testing"

	"github.com/TATAROmangol/mess/shared/utils"
)

func TestStorage_GetChatByID(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	chat, err := s.Chat().GetChatByID(t.Context(), InitChats[0].ID)
	if err != nil {
		t.Fatalf("get chat by id: %v", err)
	}

	if chat.ID != InitChats[0].ID {
		t.Fatalf("not equal, want: %v, have %v", *InitChats[0], *chat)
	}
}

func TestStorage_GetChatIDBySubjects(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	chat, err := s.Chat().GetChatIDBySubjects(t.Context(), InitChats[0].FirstSubjectID, InitChats[0].SecondSubjectID)
	if err != nil {
		t.Fatalf("get chat by id: %v", err)
	}

	if chat.ID != InitChats[0].ID {
		t.Fatalf("not equal, want: %v, have %v", *InitChats[0], *chat)
	}
}

func TestStorage_GetChatsBySubjectID(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	filter := storage.PaginationFilterIntLastID{
		Limit:     10,
		Asc:       true,
		SortLabel: storage.ChatCreatedAtLabel,
		LastID:    utils.IntPtr(1),
	}

	chats, err := s.Chat().GetChatsBySubjectID(t.Context(), InitChats[0].FirstSubjectID, &filter)
	if err != nil {
		t.Fatalf("get chat by id: %v", err)
	}

	if len(chats) != 1 {
		t.Fatalf("wait len: %v, have %v", 1, len(chats))
	}

	if chats[0].ID != InitChats[1].ID {
		t.Fatalf("not equal, want: %v, have %v", *InitChats[0], *chats[0])
	}
}

func TestStorage_IncrementChatMessageNumber(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	startChat := InitChats[0]

	chat, err := s.Chat().IncrementChatMessageNumber(t.Context(), startChat.ID)
	if err != nil {
		t.Fatalf("increment chat message number: %v", err)
	}

	if chat.MessagesCount-1 != startChat.MessagesCount {
		t.Fatalf("not increment, want 1, have: %v", *chat)
	}
}

func TestStorage_DeleteChat(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	chat, err := s.Chat().DeleteChat(t.Context(), InitChats[0].ID)
	if err != nil {
		t.Fatalf("delete chat: %v", err)
	}

	if chat.DeletedAt == nil {
		t.Fatalf("not delete: %v", *chat)
	}
}
