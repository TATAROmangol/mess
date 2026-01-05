package storage_test

import (
	"sort"
	"testing"

	"github.com/TATAROmangol/mess/profile/internal/model"
	p "github.com/TATAROmangol/mess/profile/internal/storage"
)

func TestStorage_GetKeys(t *testing.T) {
	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}
	initData(t)
	defer cleanupDB(t)

	keys, err := s.AvatarOutbox().GetKeys(t.Context(), len(InitAvatarKeys))
	if err != nil {
		t.Fatalf("get keys: %v", err)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].SubjectID > keys[j].SubjectID
	})

	sort.Slice(InitAvatarKeys, func(i, j int) bool {
		return InitAvatarKeys[i].SubjectID > InitAvatarKeys[j].SubjectID
	})

	for i, k := range keys {
		if k.Key != InitAvatarKeys[i].Key ||
			k.SubjectID != InitAvatarKeys[i].SubjectID ||
			k.DeletedAt != nil {
			t.Fatalf("not currently add, wait: %v, have: %v", InitAvatarKeys[i], k)
		}
	}
}

func TestStorage_AddKey(t *testing.T) {
	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}
	defer cleanupDB(t)

	key, err := s.AvatarOutbox().AddKey(t.Context(), InitAvatarKeys[0].SubjectID, InitAvatarKeys[0].Key)
	if err != nil {
		t.Fatalf("add keyL %v", err)
	}
	if key.SubjectID != InitAvatarKeys[0].SubjectID ||
		key.Key != InitAvatarKeys[0].Key ||
		key.DeletedAt != nil {
		t.Fatalf("not currently add, wait: %v, have: %v", InitAvatarKeys[0], key)
	}
}

func TestStorage_DeleteKeys(t *testing.T) {
	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	keys := model.GetOutboxKeys(InitAvatarKeys)

	modelKeys, err := s.AvatarOutbox().DeleteKeys(t.Context(), keys)
	if err != nil {
		t.Fatalf("delete keys: %v", err)
	}

	if len(modelKeys) != len(keys) {
		t.Fatalf("unexpected deleted keys count: got %d, want %d", len(modelKeys), len(keys))
	}

	want := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		want[k] = struct{}{}
	}

	for _, k := range modelKeys {
		if _, ok := want[k.Key]; !ok {
			t.Fatalf("unexpected deleted key: %v", k)
		}
	}
}
