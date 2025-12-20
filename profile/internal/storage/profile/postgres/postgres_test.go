package postgres_test

import (
	"context"
	"fmt"
	"os"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"profile/internal/model"
	p "profile/internal/storage/profile"
	storage "profile/internal/storage/profile/postgres"
	"profile/pkg/postgres"
	pq "profile/pkg/postgres"
	"testing"

	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var CFG pq.Config

const (
	MigrationsPath = "file://../../../../migrations/"
)

var InitProfiles = []*model.Profile{
	{
		SubjectID: "subject_id1",
		Alias:     "al",
		AvatarURL: "url",
		Version:   1,
		UpdatedAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	},
	{
		SubjectID: "subject_id2",
		Alias:     "alias",
		AvatarURL: "url",
		Version:   1,
		UpdatedAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	},
	{
		SubjectID: "subject_id3",
		Alias:     "alias pro",
		AvatarURL: "url",
		Version:   1,
		UpdatedAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	},
}

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

	CFG = pq.Config{
		Host:     host,
		Port:     port.Int(),
		User:     "test",
		Password: "test",
		DBName:   "test",
		SSLMode:  "disable",
	}

	mig, err := pq.NewMigrator(CFG, MigrationsPath)
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

	db, err := pq.New(CFG)
	if err != nil {
		t.Fatalf("connect to db: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", p.ProfileTable))
	if err != nil {
		t.Fatalf("cleanup db: %v", err)
	}
}

func initData(t *testing.T) {
	t.Helper()

	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	for _, prof := range InitProfiles {
		err = s.AddProfile(t.Context(), prof)
		if err != nil {
			t.Fatalf("first add: %v", err)
		}
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

	err = s.AddProfile(t.Context(), profileToAdd)
	if err != nil {
		t.Fatalf("first add: %v", err)
	}

	profileFromDB, err := s.GetProfileFromSubjectID(t.Context(), profileToAdd.SubjectID)
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}

	if profileFromDB.SubjectID != profileToAdd.SubjectID ||
		profileFromDB.Alias != profileToAdd.Alias ||
		profileFromDB.AvatarURL != profileToAdd.AvatarURL ||
		profileFromDB.Version != profileToAdd.Version {
		t.Fatalf("retrieved profile does not match added profile")
	}

	err = s.AddProfile(t.Context(), profileToAdd)
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

	initData(t)
	defer cleanupDB(t)

	tests := []struct {
		name    string
		profile *model.Profile
		wantErr bool
	}{
		{
			name: "successful update",
			profile: &model.Profile{
				SubjectID: "subject_id1",
				Alias:     "new_alias",
				AvatarURL: "new_url",
				Version:   2,
				UpdatedAt: time.Now().UTC(),
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
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := s.UpdateProfile(t.Context(), tt.profile)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateProfile() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateProfile() succeeded unexpectedly")
			}

			updatedProfile, err := s.GetProfileFromSubjectID(t.Context(), tt.profile.SubjectID)
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

func TestStorage_getProfilesWithPagination_Sort(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	tests := []struct {
		name string
		asc  bool
		ea   []string
	}{
		{
			name: "asc",
			asc:  true,
			ea:   []string{"al", "alias", "alias pro"},
		},
		{
			name: "desc",
			asc:  false,
			ea:   []string{"alias pro", "alias", "al"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, res, err := s.GetProfilesFromAlias(
				t.Context(),
				3,
				tt.asc,
				p.AliasLabel,
				"al",
			)
			if err != nil {
				t.Fatalf("get profiles from alias: %v", err)
			}

			if len(res) != len(tt.ea) {
				t.Fatalf("expected %d profiles, got %d", len(tt.ea), len(res))
			}

			for i, prof := range res {
				if prof.Alias != tt.ea[i] {
					t.Errorf(
						"unexpected alias at index %d: want %q, got %q",
						i,
						tt.ea[i],
						prof.Alias,
					)
				}
			}

			pag, err := pq.ParsePaginationToken(token)
			if err != nil {
				t.Fatalf("parse pagination token: %v", err)
			}

			if pag.Last.Key != nil {
				t.Fatalf("invalid last val")
			}
		})
	}
}

func TestStorage_getProfilesWithPagination_Pagination(t *testing.T) {
	s, err := storage.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	alias := "al"

	last := postgres.NewLast(p.SubjectIDLabel, nil)
	sort := postgres.NewSort(p.AliasLabel, true)

	pag := postgres.NewPagination(
		2,
		sort,
		last,
	)

	token, res, err := s.GetProfilesFromAliasWithToken(t.Context(), pag.Token(), alias)
	if len(res) != 2 {
		t.Fatalf("invalid len res: %v", len(res))
	}
	if res[0].SubjectID != InitProfiles[0].SubjectID || res[1].SubjectID != InitProfiles[1].SubjectID {
		t.Fatalf("invalid res data: %v", res)
	}

	nP, err := postgres.ParsePaginationToken(token)
	if err != nil {
		t.Fatalf("parse pagination token: %v", err)
	}

	token, res, err = s.GetProfilesFromAliasWithToken(t.Context(), nP.Token(), alias)
	if len(res) != 1 {
		t.Fatalf("invalid len res: %v", res)
	}
	if res[0].SubjectID != InitProfiles[2].SubjectID {
		t.Fatalf("invalid res data: %v", res)
	}

	pag, err = pq.ParsePaginationToken(token)
	if err != nil {
		t.Fatalf("parse pagination token: %v", err)
	}

	if pag.Last.Key != nil {
		t.Fatalf("invalid last val")
	}
}
