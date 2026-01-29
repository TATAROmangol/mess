package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/1ocknight/mess/chat/config"
	"github.com/1ocknight/mess/chat/internal/ctxkey"
	"github.com/1ocknight/mess/chat/internal/domain"
	"github.com/1ocknight/mess/chat/internal/loglables"
	"github.com/1ocknight/mess/chat/internal/storage"
	"github.com/1ocknight/mess/chat/internal/transport"
	"github.com/1ocknight/mess/chat/internal/worker"
	"github.com/1ocknight/mess/shared/auth/keycloak"
	"github.com/1ocknight/mess/shared/logger"
	"github.com/1ocknight/mess/shared/postgres"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("load config: %w", err))
		return
	}

	var logHandler *slog.HandlerOptions
	if cfg.LoggerDebug {
		logHandler = &slog.HandlerOptions{Level: slog.LevelDebug}
	}

	lg := logger.New(slog.NewJSONHandler(os.Stdout, logHandler))
	lg = lg.With(loglables.Service, "chat_microservice")

	ctx = ctxkey.WithLogger(ctx, lg)

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
	lg.Info("up migrations")

	dom := domain.New(storage)

	keycloak, err := keycloak.New(cfg.Keycloak, lg)
	if err != nil {
		lg.Error(fmt.Errorf("keycloak new: %w", err))
		return
	}

	messageWorkerLg := lg.With(loglables.Service, "message worker")
	messageWorker, err := worker.NewMessageWorker(storage, messageWorkerLg, &cfg.MessageWorker)
	if err != nil {
		lg.Error(fmt.Errorf("new message worker: %w", err))
		return
	}
	go messageWorker.Run(ctx)

	lastreadWorkerLg := lg.With(loglables.Service, "lastread worker")
	lastreadWorker, err := worker.NewLastReadWorker(storage, lastreadWorkerLg, &cfg.LastReadWorker)
	if err != nil {
		lg.Error(fmt.Errorf("new lastread worker: %w", err))
		return
	}
	go lastreadWorker.Run(ctx)

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
