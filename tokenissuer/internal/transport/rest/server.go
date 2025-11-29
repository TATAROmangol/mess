package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Server struct {
	server *http.Server
	cfg    Config
}

func NewServer(cfg Config, handlers Handler, middleware Middleware) *Server {
	router := gin.New()

	router.Use(middleware.SetMethodName())
	router.Use(middleware.SetRequestID())
	router.Use(middleware.Log())

	api := router.Group("/api")
	api.POST("/token", handlers.GetToken)
	api.GET("/refresh", handlers.Refresh)

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
