package storage_test

import (
	"database/sql"
	"errors"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"testing"

	"github.com/TATAROmangol/mess/profile/internal/model"
	"github.com/TATAROmangol/mess/profile/internal/storage"
	p "github.com/TATAROmangol/mess/profile/internal/storage"
	"github.com/TATAROmangol/mess/shared/postgres"
)

func TestStorage_AddProfile(t *testing.T) {
	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}
	defer cleanupDB(t)

	profileToAdd := &model.Profile{
		SubjectID: "subject_id",
		Alias:     "alias",
		Version:   1,
	}

	profileFromDB, err := s.Profile().AddProfile(t.Context(), profileToAdd.SubjectID, profileToAdd.Alias)
	if err != nil {
		t.Fatalf("first add: %v", err)
	}

	if profileFromDB.SubjectID != profileToAdd.SubjectID ||
		profileFromDB.Alias != profileToAdd.Alias ||
		profileFromDB.AvatarKey != profileToAdd.AvatarKey ||
		profileFromDB.Version != profileToAdd.Version ||
		profileFromDB.DeletedAt != nil {
		t.Fatalf("retrieved profile does not match added profile")
	}

	_, err = s.Profile().AddProfile(t.Context(), profileToAdd.SubjectID, profileToAdd.Alias)
	if err == nil {
		t.Fatalf("expected error on duplicate add, got nil")
	}
}

func TestStorage_UpdateProfileMetadata(t *testing.T) {
	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	type temp struct {
		SubjectID   string
		Alias       string
		PrevVersion int
	}

	tests := []struct {
		name    string
		profile temp
		wantErr bool
	}{
		{
			name: "successful update",
			profile: temp{
				SubjectID:   InitProfiles[0].SubjectID,
				Alias:       "new_alias",
				PrevVersion: 1,
			},
			wantErr: false,
		},
		{
			name: "non version update",
			profile: temp{
				SubjectID:   InitProfiles[0].SubjectID,
				Alias:       "another_alias",
				PrevVersion: 1,
			},
			wantErr: true,
		},
		{
			name: "update non-existing profile",
			profile: temp{
				SubjectID:   "non_existing_subject_id",
				Alias:       "alias",
				PrevVersion: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedProfile, gotErr := s.Profile().UpdateProfileMetadata(t.Context(), tt.profile.SubjectID, tt.profile.PrevVersion, tt.profile.Alias)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateProfile() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateProfile() succeeded unexpectedly")
			}

			if updatedProfile.Alias != tt.profile.Alias ||
				updatedProfile.Version != tt.profile.PrevVersion+1 {
				t.Errorf("Profile not updated correctly new: %v, prev: %v", updatedProfile, tt.profile)
			}
		})
	}
}

func TestStorage_UpdateAvatarKey(t *testing.T) {
	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	key := "test-key"
	prof, err := s.Profile().UpdateAvatarKey(t.Context(), InitProfiles[0].SubjectID, key)
	if err != nil {
		t.Fatalf("update avatar key: %v", err)
	}

	if *prof.AvatarKey != key {
		t.Fatalf("avatar keys not equals, new: %v, cur %v", *prof.AvatarKey, key)
	}
}

func TestStorage_GetProfilesFromAlias_PaginationForward(t *testing.T) {
	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	filer := postgres.PaginationFilter{
		Limit:     10,
		Asc:       true,
		SortLabel: storage.ProfileAliasLabel,
		IDLabel:   storage.ProfileSubjectIDLabel,
		LastID:    &InitProfiles[0].SubjectID,
	}

	profiles, err := s.Profile().GetProfilesFromAlias(t.Context(), "al", &filer)
	if err != nil {
		t.Fatalf("get profiles from alias: %v", err)
	}

	if len(profiles) != 2 {
		t.Fatalf("not wait len: %v, wait 2", len(profiles))
	}

	if profiles[1].SubjectID != InitProfiles[2].SubjectID {
		t.Fatalf("not equal profiles: %v, want %v", profiles[2].SubjectID, InitProfiles[2].SubjectID)
	}
}

func TestStorage_GetProfilesFromAlias_PaginationBack(t *testing.T) {
	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	filer := postgres.PaginationFilter{
		Limit:     10,
		Asc:       false,
		SortLabel: storage.ProfileAliasLabel,
		IDLabel:   storage.ProfileSubjectIDLabel,
		LastID:    &InitProfiles[2].SubjectID,
	}

	profiles, err := s.Profile().GetProfilesFromAlias(t.Context(), "al", &filer)
	if err != nil {
		t.Fatalf("get profiles from alias: %v", err)
	}

	if len(profiles) != 2 {
		t.Fatalf("not wait len: %v, wait 2", len(profiles))
	}

	if profiles[1].SubjectID != InitProfiles[0].SubjectID {
		t.Fatalf("not equal profiles: %v, want %v", profiles[1].SubjectID, InitProfiles[0].SubjectID)
	}
}

func TestStorage_GetProfilesFromAlias_WithoutPagination(t *testing.T) {
	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	filer := postgres.PaginationFilter{
		Limit:     10,
		Asc:       true,
		SortLabel: storage.ProfileAliasLabel,
		IDLabel:   storage.ProfileSubjectIDLabel,
	}

	profiles, err := s.Profile().GetProfilesFromAlias(t.Context(), "al", &filer)
	if err != nil {
		t.Fatalf("get profiles from alias: %v", err)
	}

	if len(profiles) != 3 {
		t.Fatalf("not wait len: %v, wait 3", len(profiles))
	}

	if profiles[2].SubjectID != InitProfiles[2].SubjectID {
		t.Fatalf("not equal profiles: %v, want %v", profiles[2].SubjectID, InitProfiles[2].SubjectID)
	}
}

func TestStorage_DeleteProfile(t *testing.T) {
	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	delID := InitProfiles[0].SubjectID

	prof, err := s.Profile().DeleteProfile(t.Context(), delID)
	if err != nil {
		t.Fatalf("delete profile subject id: %v", err)
	}
	if prof == nil {
		t.Fatalf("expected nil profile on delete non-existing, got %+v", prof)
	}

	res, err := s.Profile().GetProfileFromSubjectID(t.Context(), prof.SubjectID)
	if err != nil && !errors.Is(err, storage.ErrNoRows) {
		t.Fatalf("get profiles from alias: %v", err)
	}
	if res != nil {
		t.Fatalf("not delete profile: %v", res)
	}

	_, err = s.Profile().DeleteProfile(t.Context(), "not")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("delete profile subject id: %v", err)
	}
}

func TestStorage_DeleteAvatarKey(t *testing.T) {
	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	prof, err := s.Profile().DeleteAvatarKey(t.Context(), InitProfiles[0].SubjectID)
	if err != nil {
		t.Fatalf("delete avatar key: %v", err)
	}

	if prof.AvatarKey != nil {
		t.Fatalf("avatar key not nil")
	}
}
