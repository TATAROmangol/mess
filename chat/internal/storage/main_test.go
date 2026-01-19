package storage_test

import (
	"github.com/TATAROmangol/mess/chat/internal/model"
	"github.com/TATAROmangol/mess/chat/internal/storage"
	"context"
	"fmt"
	"os"
	"testing"

	pq "github.com/TATAROmangol/mess/shared/postgres"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var CFG pq.Config

const (
	MigrationsPath = "file://../../migrations/"
)

var InitChats = []*model.Chat{
	{
		ID:              1,
		FirstSubjectID:  "subj-1",
		SecondSubjectID: "subj-2",
	},
	{
		ID:              2,
		FirstSubjectID:  "subj-1",
		SecondSubjectID: "subj-3",
	},
}

var InitLastReads = []*model.LastRead{
	{
		SubjectID:     "subj-1",
		ChatID:        1,
		MessageNumber: 1,
	},
	{
		SubjectID:     "subj-2",
		ChatID:        1,
		MessageNumber: 2,
	},
	{
		SubjectID:     "subj-1",
		ChatID:        2,
		MessageNumber: 0,
	},
	{
		SubjectID:     "subj-3",
		ChatID:        2,
		MessageNumber: 1,
	},
}

var InitMessages = []*model.Message{
	{
		ID:              1,
		ChatID:          1,
		SenderSubjectID: "subj-1",
		Content:         "test-content-1",
		Version:         1,
	},
	{
		ID:              2,
		ChatID:          1,
		SenderSubjectID: "subj-2",
		Content:         "test-content-2",
		Version:         1,
	},
	{
		ID:              3,
		ChatID:          2,
		SenderSubjectID: "subj-2",
		Content:         "test-content-1",
		Version:         1,
	},
}

var InitMessageOutboxes = []*model.MessageOutbox{
	{
		ID:        1,
		ChatID:    1,
		MessageID: 1,
		Operation: model.AddOperation,
	},
	{
		ID:        2,
		ChatID:    1,
		MessageID: 2,
		Operation: model.AddOperation,
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

	_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", storage.ChatTable))
	if err != nil {
		t.Fatalf("cleanup db: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", storage.LastReadTable))
	if err != nil {
		t.Fatalf("cleanup db: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", storage.MessageTable))
	if err != nil {
		t.Fatalf("cleanup db: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", storage.MessageOutboxTable))
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

	for _, chat := range InitChats {
		_, err = s.Chat().CreateChat(t.Context(), chat.FirstSubjectID, chat.SecondSubjectID)
		if err != nil {
			t.Fatalf("create chat: %v", err)
		}
	}

	for _, lr := range InitLastReads {
		_, err = s.LastRead().CreateLastRead(t.Context(), lr.SubjectID, lr.ChatID)
		if err != nil {
			t.Fatalf("create last read: %v", err)
		}
	}

	for _, ms := range InitMessages {
		_, err = s.Message().CreateMessage(t.Context(), ms.ChatID, ms.SenderSubjectID, ms.Content, ms.Number)
		if err != nil {
			t.Fatalf("create message: %v", err)
		}
	}

	for _, ms := range InitMessageOutboxes {
		_, err = s.MessageOutbox().AddMessageOutbox(t.Context(), ms.ChatID, ms.MessageID, ms.Operation)
		if err != nil {
			t.Fatalf("create message outbox: %v", err)
		}
	}
}
