package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/TATAROmangol/mess/profile/config"
	"github.com/TATAROmangol/mess/profile/internal/adapter/avatar"
	"github.com/TATAROmangol/mess/profile/internal/ctxkey"
	"github.com/TATAROmangol/mess/profile/internal/domain"
	"github.com/TATAROmangol/mess/profile/internal/loglables"
	"github.com/TATAROmangol/mess/profile/internal/storage"
	"github.com/TATAROmangol/mess/profile/internal/transport"
	workers "github.com/TATAROmangol/mess/profile/internal/wokers"
	"github.com/TATAROmangol/mess/shared/auth/keycloak"
	"github.com/TATAROmangol/mess/shared/logger"
	"github.com/TATAROmangol/mess/shared/postgres"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lg := logger.New(slog.NewJSONHandler(os.Stdout, nil))
	lg = lg.With(loglables.ServiceName, "profile_microservice")

	ctx = ctxkey.WithLogger(ctx, lg)

	cfg, err := config.LoadConfig()
	if err != nil {
		lg.Error(fmt.Errorf("load config: %v", err))
		return
	}

	storage, err := storage.New(cfg.Postgres)
	if err != nil {
		lg.Error(fmt.Errorf("storage new: %v", err))
		return
	}

	mig, err := postgres.NewMigrator(cfg.Postgres, cfg.MigrationsPath)
	if err != nil {
		lg.Error(fmt.Errorf("migrator new: %v", err))
		return
	}
	defer mig.Close()

	if err = mig.Up(); err != nil {
		lg.Error(fmt.Errorf("migrator up: %v", err))
		return
	}

	avatar, err := avatar.New(ctx, cfg.S3)
	if err != nil {
		lg.Error(fmt.Errorf("avatar new: %v", err))
		return
	}

	dom := domain.New(storage, avatar)

	ad := workers.NewAvatarDeleter(cfg.AvatarDeleter, avatar, storage.AvatarOutbox())
	err = ad.Start(ctx)
	if err != nil {
		lg.Error(fmt.Errorf("avatar deleter start: %v", err))
		return
	}

	au := workers.NewAvatarUploader(cfg.AvatarUploader, storage)
	err = au.Start(ctx)
	if err != nil {
		lg.Error(fmt.Errorf("avatar uploader start: %v", err))
		return
	}

	pd := workers.NewProfileDeleter(cfg.ProfileDeleter, storage.Profile())
	err = pd.Start(ctx)
	if err != nil {
		lg.Error(fmt.Errorf("profile deleter start: %v", err))
		return
	}

	keycloak, err := keycloak.New(cfg.Keycloak, lg)
	if err != nil {
		lg.Error(fmt.Errorf("keycloak new: %v", err))
		return
	}

	server := transport.NewServer(cfg.HTTP, lg, dom, keycloak)
	go func() {
		if err := server.Run(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			lg.Error(fmt.Errorf("server run: %v", err))
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	lg.Info("start graceful shutdown")

	err = server.Stop(ctx)
	if err != nil {
		lg.Error(fmt.Errorf("server stop: %v", err))
	}
	lg.Info("server is stop")

	cancel()
	lg.Info("successful stop")
}
