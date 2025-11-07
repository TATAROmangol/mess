package storage_test

import (
	"context"
	"testing"

	"auth/internal/entities"
	"auth/internal/storage"

	_ "github.com/mattn/go-sqlite3"
)

func TestStorage_CreateSubject_Integration(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	defer db.Close()

	s := storage.New(db)
	ctx := context.Background()

	tests := []struct {
		name      string
		login     string
		password  string
		setup     func()
		expectErr bool
	}{
		{
			name:      "create new subject",
			login:     "user1",
			password:  "pass1",
			expectErr: false,
		},
		{
			name:     "duplicate login fails",
			login:    "user_dup",
			password: "pass123",
			setup: func() {
				_, err := s.CreateSubject(ctx, "user_dup", "old_pass")
				if err != nil {
					t.Fatalf("setup duplicate: %v", err)
				}
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.setup != nil {
				tt.setup()
			}

			id, err := s.CreateSubject(ctx, tt.login, tt.password)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			// проверяем корректность создания
			subj, err := s.GetSubjectByID(ctx, id)
			if err != nil {
				t.Fatalf("GetSubjectByID failed: %v", err)
			}

			if subj.Login != tt.login {
				t.Errorf("login mismatch: got %s want %s", subj.Login, tt.login)
			}
			if subj.PasswordHash != tt.password {
				t.Errorf("password mismatch: got %s want %s", subj.PasswordHash, tt.password)
			}
			if subj.DeletedAt != nil {
				t.Errorf("expected DeletedAt nil, got %v", subj.DeletedAt)
			}
		})
	}
}

func TestStorage_GetSubjectByID_Integration(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	defer db.Close()

	s := storage.New(db)
	ctx := context.Background()

	id1, err := s.CreateSubject(ctx, "user3", "hash1")
	if err != nil {
		t.Fatalf("CreateSubject failed: %v", err)
	}
	id2, err := s.CreateSubject(ctx, "user4", "hash2")
	if err != nil {
		t.Fatalf("CreateSubject failed: %v", err)
	}

	tests := []struct {
		name      string
		subjID    int
		want      *entities.Subject
		expectErr bool
	}{
		{
			name:   "existing user1",
			subjID: id1,
			want: &entities.Subject{
				ID:           id1,
				Login:        "user3",
				PasswordHash: "hash1",
			},
			expectErr: false,
		},
		{
			name:   "existing user2",
			subjID: id2,
			want: &entities.Subject{
				ID:           id2,
				Login:        "user4",
				PasswordHash: "hash2",
			},
			expectErr: false,
		},
		{
			name:      "non-existent user",
			subjID:    9999,
			want:      nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetSubjectByID(ctx, tt.subjID)
			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error=%v, got %v", tt.expectErr, err)
			}
			if !tt.expectErr {
				if got.ID != tt.want.ID || got.Login != tt.want.Login || got.PasswordHash != tt.want.PasswordHash {
					t.Errorf("GetSubjectByID() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestStorage_DeleteSubject_Integration(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	defer db.Close()

	s := storage.New(db)
	ctx := context.Background()

	id, err := s.CreateSubject(ctx, "user_del", "hash_del")
	if err != nil {
		t.Fatalf("CreateSubject failed: %v", err)
	}

	tests := []struct {
		name      string
		subjID    int
		expectErr bool
	}{
		{
			name:      "soft delete existing",
			subjID:    id,
			expectErr: false,
		},
		{
			name:      "delete non-existent",
			subjID:    99999,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.DeleteSubject(ctx, tt.subjID)
			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error=%v got %v", tt.expectErr, err)
			}

			if tt.subjID == id && !tt.expectErr {
				_, err := s.GetSubjectByID(ctx, id)
				if err == nil {
					t.Fatalf("expected error when fetching deleted subject, got nil")
				}
			}
		})
	}
}

func TestStorage_ChangePassword_Integration(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	defer db.Close()

	s := storage.New(db)
	ctx := context.Background()

	existingID, err := s.CreateSubject(ctx, "user_change", "old_hash")
	if err != nil {
		t.Fatalf("CreateSubject failed: %v", err)
	}

	tests := []struct {
		name        string
		subjID      int
		newPassword string
		expectErr   bool
	}{
		{
			name:        "change password ok",
			subjID:      existingID,
			newPassword: "new_hash_123",
			expectErr:   false,
		},
		{
			name:        "change password ok",
			subjID:      existingID,
			newPassword: "old_hash",
			expectErr:   false,
		},
		{
			name:        "non-existent user",
			subjID:      99999,
			newPassword: "whatever",
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ChangePassword(ctx, tt.subjID, tt.newPassword)
			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error=%v got %v", tt.expectErr, err)
			}

			if tt.expectErr {
				return
			}

			subj, err := s.GetSubjectByID(ctx, tt.subjID)
			if err != nil {
				t.Fatalf("GetSubjectByID failed: %v", err)
			}

			if subj.PasswordHash != tt.newPassword {
				t.Fatalf("password not updated: got=%s want=%s", subj.PasswordHash, tt.newPassword)
			}
		})
	}
}
