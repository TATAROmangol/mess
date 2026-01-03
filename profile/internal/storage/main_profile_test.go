package storage_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/TATAROmangol/mess/profile/internal/model"
	p "github.com/TATAROmangol/mess/profile/internal/storage"
	pq "github.com/TATAROmangol/mess/shared/postgres"
	"github.com/TATAROmangol/mess/shared/utils"
	"github.com/stretchr/testify/assert"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var CFG pq.Config

const (
	MigrationsPath = "file://../../migrations/"
)

var InitProfiles = []*model.Profile{
	{
		SubjectID: "subject_id1",
		Alias:     "al",
		AvatarKey: utils.StringPtr("url"),
		Version:   1,
	},
	{
		SubjectID: "subject_id2",
		Alias:     "alias",
		AvatarKey: utils.StringPtr("url"),
		Version:   1,
	},
	{
		SubjectID: "subject_id3",
		Alias:     "alias pro",
		AvatarKey: utils.StringPtr("url"),
		Version:   1,
	},
}

var InitAvatarKeys = []*model.AvatarKeyOutbox{
	{
		SubjectID: "subject_id1",
		Key:       "key1",
	},
	{
		SubjectID: "subject_id2",
		Key:       "key2",
	},
	{
		SubjectID: "subject_id3",
		Key:       "key3",
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

	_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", p.AvatarKeyOutboxTable))
	if err != nil {
		t.Fatalf("cleanup db: %v", err)
	}
}

func initData(t *testing.T) {
	t.Helper()

	s, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	for _, prof := range InitProfiles {
		_, err = s.Profile().AddProfile(t.Context(), prof.SubjectID, prof.Alias)
		if err != nil {
			t.Fatalf("init add: %v", err)
		}
	}

	for _, k := range InitAvatarKeys {
		_, err = s.AvatarKeyOutbox().AddKey(t.Context(), k.SubjectID, k.Key)
		if err != nil {
			t.Fatalf("init add: %v", err)
		}
	}
}

func TestStorage_Transaction_Commit(t *testing.T) {
	st, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	defer cleanupDB(t)

	s, err := st.WithTransaction(t.Context())
	if err != nil {
		t.Fatalf("failed create storage with trasaction: %v", err)
	}

	prof, err := s.Profile().AddProfile(t.Context(), InitProfiles[0].SubjectID, InitProfiles[0].Alias)
	if err != nil {
		t.Fatalf("delete avatar key: %v", err)
	}

	key, err := s.AvatarKeyOutbox().AddKey(t.Context(), InitProfiles[0].SubjectID, "test")
	if err != nil {
		t.Fatalf("add key: %v", err)
	}

	err = s.Commit()
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	// imitation defer
	err = s.Rollback()
	if err == nil {
		t.Fatalf("rollback not have error from commit")
	}

	prof2, err := st.Profile().GetProfileFromSubjectID(t.Context(), InitProfiles[0].SubjectID)
	if err != nil {
		t.Fatalf("get profile from subject id: %v", err)
	}
	assert.Equal(t, prof, prof2)

	keys, err := st.AvatarKeyOutbox().GetKeys(t.Context(), 100)
	if err != nil {
		t.Fatalf("get keys: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("mot correct len, wait 1, have: %v", len(keys))
	}
	assert.Equal(t, key, keys[0])
}

func TestStorage_Transaction_Rollback(t *testing.T) {
	st, err := p.New(CFG)
	if err != nil {
		t.Fatalf("could not construct receiver type: %v", err)
	}

	defer cleanupDB(t)

	s, err := st.WithTransaction(t.Context())
	if err != nil {
		t.Fatalf("failed create storage with trasaction: %v", err)
	}

	_, err = s.Profile().AddProfile(t.Context(), InitProfiles[0].SubjectID, InitProfiles[0].Alias)
	if err != nil {
		t.Fatalf("delete avatar key: %v", err)
	}

	_, err = s.AvatarKeyOutbox().AddKey(t.Context(), InitProfiles[0].SubjectID, "test")
	if err != nil {
		t.Fatalf("add key: %v", err)
	}

	err = s.Rollback()
	if err != nil {
		t.Fatalf("rollback: %v", err)
	}

	_, err = st.Profile().GetProfileFromSubjectID(t.Context(), InitProfiles[0].SubjectID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("get profile from subject id: %v", err)
	}

	_, err = st.AvatarKeyOutbox().GetKeys(t.Context(), 100)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("get keys: %v", err)
	}
}
