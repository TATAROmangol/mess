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

	"github.com/TATAROmangol/mess/shared/auth/keycloak"
	"github.com/TATAROmangol/mess/shared/logger"
	"github.com/TATAROmangol/mess/websocket/config"
	"github.com/TATAROmangol/mess/websocket/internal/ctxkey"
	"github.com/TATAROmangol/mess/websocket/internal/loglables"
	"github.com/TATAROmangol/mess/websocket/internal/model"
	"github.com/TATAROmangol/mess/websocket/internal/transport"
	"github.com/TATAROmangol/mess/websocket/internal/worker"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lg := logger.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	lg = lg.With(loglables.Service, "websocket_microservice")

	ctx = ctxkey.WithLogger(ctx, lg)

	cfg, err := config.LoadConfig()
	if err != nil {
		lg.Error(fmt.Errorf("load config: %w", err))
		return
	}

	msgs := make(chan *model.Message)

	keycloak, err := keycloak.New(cfg.Keycloak, lg)
	if err != nil {
		lg.Error(fmt.Errorf("keycloak new: %w", err))
		return
	}

	messageWorkerLg := lg.With(loglables.Layer, "message worker")
	messageWorker, err := worker.NewMessageWorker(cfg.MessageWorker, msgs, messageWorkerLg)
	if err != nil {
		lg.Error(fmt.Errorf("new message worker: %w", err))
		return
	}
	go messageWorker.Run(ctx)

	lastreadWorkerLg := lg.With(loglables.Layer, "lastread worker")
	lastreadWorker, err := worker.NewLastReadWorker(cfg.LastReadWorker, msgs, lastreadWorkerLg)
	if err != nil {
		lg.Error(fmt.Errorf("new message worker: %w", err))
		return
	}
	go lastreadWorker.Run(ctx)

	hubLg := lg.With(loglables.Layer, "hub")
	hub := transport.NewHub(msgs, hubLg)
	go hub.Run()

	handler := transport.NewHandler(cfg.WSConfig, hub)
	server := transport.NewServer(cfg.HTTP, keycloak, handler)
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
