package rest

import (
	"context"
	"fmt"
	"net/http"
	"tokenissuer/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Port int
	Host string
}

type Server struct {
	server *http.Server
	cfg    Config
}

func NewServer(cfg Config, log logger.Logger, handlers Handler, middleware Middleware) *Server {
	router := gin.New()

	router.Use(middleware.SetMethodName())
	router.Use(middleware.SetRequestID())
	router.Use(middleware.Log())

	api := router.Group("/api")
	api.GET("/token", handlers.GetToken)
	api.POST("/refresh", handlers.Refresh)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%v:%v", cfg.Host, cfg.Port),
		Handler: router,
	}

	return &Server{
		server: httpServer,
		cfg:    cfg,
	}
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
