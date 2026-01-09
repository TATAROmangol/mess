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
	lg = lg.With(loglables.Service, "profile_microservice")

	ctx = ctxkey.WithLogger(ctx, lg)

	cfg, err := config.LoadConfig()
	if err != nil {
		lg.Error(fmt.Errorf("load config: %w", err))
		return
	}

	storage, err := storage.New(cfg.Postgres)
	if err != nil {
		lg.Error(fmt.Errorf("storage new: %w", err))
		return
	}

	mig, err := postgres.NewMigrator(cfg.Postgres, cfg.MigrationsPath)
	if err != nil {
		lg.Error(fmt.Errorf("migrator new: %w", err))
		return
	}
	defer mig.Close()

	if err = mig.Up(); err != nil {
		lg.Error(fmt.Errorf("migrator up: %w", err))
		return
	}

	avatar, err := avatar.New(ctx, cfg.S3)
	if err != nil {
		lg.Error(fmt.Errorf("avatar new: %w", err))
		return
	}

	dom := domain.New(storage, avatar)

	ad := workers.NewAvatarDeleter(cfg.AvatarDeleter, avatar, storage.AvatarOutbox())
	avdelLog := lg.With(loglables.Layer, "worker_avatar_deleter")
	err = ad.Start(ctxkey.WithLogger(ctx, avdelLog))
	if err != nil {
		lg.Error(fmt.Errorf("avatar deleter start: %w", err))
		return
	}
	lg.Info("avatar deleter started")

	au := workers.NewAvatarUploader(cfg.AvatarUploader, storage)
	upLog := lg.With(loglables.Layer, "worker_avatar_uploader")
	err = au.Start(ctxkey.WithLogger(ctx, upLog))
	if err != nil {
		lg.Error(fmt.Errorf("avatar uploader start: %w", err))
		return
	}
	lg.Info("avatar uploader started")

	pd := workers.NewProfileDeleter(cfg.ProfileDeleter, storage.Profile())
	pdelLog := lg.With(loglables.Layer, "worker_profile_deleter")
	err = pd.Start(ctxkey.WithLogger(ctx, pdelLog))
	if err != nil {
		lg.Error(fmt.Errorf("profile deleter start: %w", err))
		return
	}
	lg.Info("profile deleter started")

	keycloak, err := keycloak.New(cfg.Keycloak, lg)
	if err != nil {
		lg.Error(fmt.Errorf("keycloak new: %w", err))
		return
	}

	server := transport.NewServer(cfg.HTTP, lg, dom, keycloak)
	go func() {
		if err := server.Run(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			lg.Error(fmt.Errorf("server run: %w", err))
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	lg.Info("start graceful shutdown")

	err = server.Stop(ctx)
	if err != nil {
		lg.Error(fmt.Errorf("server stop: %w", err))
	}
	lg.Info("server is stop")

	cancel()
	lg.Info("successful stop")
}
