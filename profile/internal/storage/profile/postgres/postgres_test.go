package postgres_test

import (
	"context"
	"fmt"
	"os"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	// storage "profile/internal/storage/profile/postgres"
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

	fmt.Println("its ok")

	os.Exit(m.Run())
}

// func TestStorage_GetProfileFromSubjectID(t *testing.T) {

// 	tests := []struct {
// 		name    string
// 		cfg     pg.Config
// 		subjID  string
// 		want    *model.Profile
// 		wantErr bool
// 	}{}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// s, err := postgres.New(tt.cfg)
// 			// if err != nil {
// 			// 	t.Fatalf("could not construct receiver type: %v", err)
// 			// }
// 			// got, gotErr := s.GetProfileFromSubjectID(tt.subjID)
// 			// if gotErr != nil {
// 			// 	if !tt.wantErr {
// 			// 		t.Errorf("GetProfileFromSubjectID() failed: %v", gotErr)
// 			// 	}
// 			// 	return
// 			// }
// 			// if tt.wantErr {
// 			// 	t.Fatal("GetProfileFromSubjectID() succeeded unexpectedly")
// 			// }
// 			// // TODO: update the condition below to compare got with tt.want.
// 			// if true {
// 			// 	t.Errorf("GetProfileFromSubjectID() = %v, want %v", got, tt.want)
// 			// }
// 		})
// 	}
// }
