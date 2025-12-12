package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tokenissuer/internal/adapter/identifier/keycloak"
	"tokenissuer/internal/config"
	"tokenissuer/internal/ctxkey"
	"tokenissuer/internal/service"
	"tokenissuer/internal/transport/grpc"
	"tokenissuer/internal/transport/rest"
	"tokenissuer/pkg/logger"

	"github.com/gin-gonic/gin"
)

const (
	ConfigPath = "CONFIG_PATH"
)

func main() {
	localPath := flag.String("config-path", "", "Path to local config file")
	flag.Parse()

	path := *localPath
	if path == "" {
		gin.SetMode(gin.ReleaseMode)
		path = os.Getenv(ConfigPath)
		if path == "" {
			log.Fatal("Error: provide --config-path or set CONFIG_PATH environment variable")
			os.Exit(1)
		}
	}

	cfg, err := config.Load(path)
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	l := logger.New(os.Stdout, ctxkey.Parse)
	ctx := context.Background()

	iden := keycloak.NewKeycloak(cfg.Keycloak)
	token := service.NewTokenImpl(iden)
	ver, err := service.NewVerifyImpl(ctx, iden, cfg.VerifyService)
	if err != nil {
		l.Error(fmt.Errorf("new verify impl: %w", err))
		os.Exit(1)
	}
	service := service.NewServiceImpl(token, ver)

	httpHandler := rest.NewHandler(service.Token)
	httpMiddleware := rest.NewMiddleware(l)
	httpServer := rest.NewServer(cfg.HTTP, httpHandler, httpMiddleware)

	grpcHandler := grpc.NewHandlerImpl(service.Verify)
	grpcInterceptor := grpc.NewInterceptorImpl(l)
	grpcServer := grpc.NewServer(cfg.GRPC, grpcInterceptor, grpcHandler)

	go func() {
		if err := httpServer.Run(); err != nil && err != http.ErrServerClosed {
			l.Error(err)
			os.Exit(1)
		}
	}()
	l.Info(fmt.Sprintf("http server start - host: %v, port: %v", cfg.HTTP.Host, cfg.HTTP.Port))

	go func() {
		if err := grpcServer.Run(); err != nil {
			l.Error(err)
			os.Exit(1)
		}
	}()
	l.Info(fmt.Sprintf("grpc server start - host: %v, port: %v", cfg.GRPC.Host, cfg.GRPC.Port))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	l.Info("Shutting down servers...")
	httpServer.Stop(ctx)
	grpcServer.Stop()
	l.Info("Servers stopped")
}
