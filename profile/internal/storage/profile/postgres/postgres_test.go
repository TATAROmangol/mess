package postgres_test

import (
	"context"
	"fmt"
	"os"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"profile/internal/model"
	storage "profile/internal/storage/profile/postgres"
	pg "profile/pkg/postgres"
	"testing"

	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var CFG pg.Config

const (
	MigrationsPath = "file://../../../../migrations/"
)

// init pg container
func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := pgcontainer.Run(
		ctx,
		"postgres:15-alpine",
		pgcontainer.WithDatabase("test"),
		pgcontainer.WithUsername("test"),
		pgcontainer.WithPassword("test"),
		pgcontainer.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to start postgres container:", err)
		os.Exit(1)
	}
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get container host:", err)
		os.Exit(1)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get container port:", err)
		os.Exit(1)
	}

	CFG = pg.Config{
		Host:     host,
		Port:     port.Int(),
		User:     "test",
		Password: "test",
		DBName:   "test",
		SSLMode:  "disable",
	}

	mig, err := pg.NewMigrator(CFG, MigrationsPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create migrator:", err)
		os.Exit(1)
	}

	if err := mig.Up(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to run migrations:", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func cleanupDB(t *testing.T) {
	t.Helper()

	db, err := pg.New(CFG)
	if err != nil {
		t.Fatalf("connect to db: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", storage.ProfileTable))
	if err != nil {
		t.Fatalf("cleanup db: %v", err)
	}
}

func TestStorage_AddProfile_GetProfileFromSubjectID(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	profileToAdd := &model.Profile{
		SubjectID: "subject_id",
		Alias:     "alias",
		AvatarURL: "url",
		Version:   1,
		UpdatedAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}

	err = s.AddProfile(profileToAdd)
	if err != nil {
		t.Fatalf("first add: %v", err)
	}

	profileFromDB, err := s.GetProfileFromSubjectID(profileToAdd.SubjectID)
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}

	if profileFromDB.SubjectID != profileToAdd.SubjectID ||
		profileFromDB.Alias != profileToAdd.Alias ||
		profileFromDB.AvatarURL != profileToAdd.AvatarURL ||
		profileFromDB.Version != profileToAdd.Version {
		t.Fatalf("retrieved profile does not match added profile")
	}

	err = s.AddProfile(profileToAdd)
	if err == nil {
		t.Fatalf("expected error on duplicate add, got nil")
	}

	cleanupDB(t)
}

func TestStorage_UpdateProfile(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	profileToAdd := &model.Profile{
		SubjectID: "subject_id",
		Alias:     "alias",
		AvatarURL: "url",
		Version:   1,
		UpdatedAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}

	err = s.AddProfile(profileToAdd)
	if err != nil {
		t.Fatalf("first add: %v", err)
	}

	tests := []struct {
		name    string
		profile *model.Profile
		wantErr bool
	}{
		{
			name: "successful update",
			profile: &model.Profile{
				SubjectID: "subject_id",
				Alias:     "new_alias",
				AvatarURL: "new_url",
				Version:   2,
				UpdatedAt: time.Now().UTC(),
				CreatedAt: profileToAdd.CreatedAt,
			},
			wantErr: false,
		},
		{
			name: "nont version update",
			profile: &model.Profile{
				SubjectID: "subject_id",
				Alias:     "another_alias",
				AvatarURL: "another_url",
				Version:   2,
				UpdatedAt: time.Now().UTC(),
				CreatedAt: profileToAdd.CreatedAt,
			},
			wantErr: true,
		},
		{
			name: "update non-existing profile",
			profile: &model.Profile{
				SubjectID: "non_existing_subject_id",
				Alias:     "alias",
				AvatarURL: "url",
				Version:   1,
				UpdatedAt: time.Now().UTC(),
				CreatedAt: time.Now().UTC(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := s.UpdateProfile(tt.profile)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateProfile() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateProfile() succeeded unexpectedly")
			}

			updatedProfile, err := s.GetProfileFromSubjectID(tt.profile.SubjectID)
			if err != nil {
				t.Fatalf("GetProfileFromSubjectID() failed: %v", err)
			}

			if updatedProfile.Alias != tt.profile.Alias ||
				updatedProfile.AvatarURL != tt.profile.AvatarURL ||
				updatedProfile.Version != tt.profile.Version {
				t.Errorf("Profile not updated correctly")
			}
		})
	}
}
