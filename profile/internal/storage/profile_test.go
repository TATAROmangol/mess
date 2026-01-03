package storage_test

import (
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"testing"

	pq "github.com/TATAROmangol/mess/shared/postgres"

	"github.com/TATAROmangol/mess/profile/internal/model"
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
	err = s.Profile().UpdateAvatarKey(t.Context(), InitProfiles[0].SubjectID, key)
	if err != nil {
		t.Fatalf("update avatar key: %v", err)
	}

	prof, err := s.Profile().GetProfileFromSubjectID(t.Context(), InitProfiles[0].SubjectID)
	if err != nil {
		t.Fatalf("get profile from subject id: %v", err)
	}

	if *prof.AvatarKey != key {
		t.Fatalf("avatar keys not equals, new: %v, cur %v", *prof.AvatarKey, key)
	}
}

func TestStorage_GetProfilesWithPagination_Sort(t *testing.T) {
	s, err := p.New(CFG)
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
			token, res, err := s.Profile().GetProfilesFromAlias(
				t.Context(),
				3,
				tt.asc,
				p.ProfileAliasLabel,
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

func TestStorage_GetProfilesWithPagination_Pagination(t *testing.T) {
	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	alias := "al"

	last := postgres.NewLast(p.ProfileSubjectIDLabel, nil)
	sort := postgres.NewSort(p.ProfileAliasLabel, true)

	pag := postgres.NewPagination(
		2,
		sort,
		last,
	)

	token, res, err := s.Profile().GetProfilesFromAliasWithToken(t.Context(), pag.Token(), alias)
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

	token, res, err = s.Profile().GetProfilesFromAliasWithToken(t.Context(), nP.Token(), alias)
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

func TestStorage_DeleteProfile(t *testing.T) {
	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	initData(t)
	defer cleanupDB(t)

	delID := InitProfiles[0].SubjectID

	err = s.Profile().DeleteProfile(t.Context(), delID)
	if err != nil {
		t.Fatalf("delete profile from subjectID: %v", err)
	}

	_, res, err := s.Profile().GetProfilesFromAlias(t.Context(), 100, true, p.ProfileAliasLabel, "")
	if err != nil {
		t.Fatalf("get profiles from alias: %v", err)
	}

	if len(res) == len(InitProfiles) {
		t.Fatalf("not delete profile")
	}

	err = s.Profile().DeleteProfile(t.Context(), "not")
	if err != nil {
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

	err = s.Profile().DeleteAvatarKey(t.Context(), InitProfiles[0].SubjectID)
	if err != nil {
		t.Fatalf("delete avatar key: %v", err)
	}

	prof, err := s.Profile().GetProfileFromSubjectID(t.Context(), InitProfiles[0].SubjectID)
	if err != nil {
		t.Fatalf("get profile from subject id: %v", err)
	}

	if prof.AvatarKey != nil {
		t.Fatalf("avatar key not nil")
	}
}
