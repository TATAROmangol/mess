package storage_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/TATAROmangol/mess/profile/internal/model"
	p "github.com/TATAROmangol/mess/profile/internal/storage"
	pq "github.com/TATAROmangol/mess/shared/postgres"
	"github.com/TATAROmangol/mess/shared/utils"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var CFG pq.Config

const (
	MigrationsPath = "file://../../../migrations/"
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
		_, err = s.AddProfile(t.Context(), prof.SubjectID, prof.Alias)
		if err != nil {
			t.Fatalf("init add: %v", err)
		}
	}

	for _, k := range InitAvatarKeys {
		_, err = s.AddKey(t.Context(), k.SubjectID, k.Key)
		if err != nil {
			t.Fatalf("init add: %v", err)
		}
	}
}
