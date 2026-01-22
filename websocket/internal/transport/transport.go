package transport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/TATAROmangol/mess/shared/auth"
	"github.com/TATAROmangol/mess/shared/logger"
	"github.com/gorilla/mux"
)

type Server struct {
	cfg         HTTPConfig
	Router      *mux.Router
	AuthService auth.Service
	Logger      logger.Logger
	httpServer  *http.Server
}

func NewServer(cfg HTTPConfig, authService auth.Service, handler *Handler) *Server {
	r := mux.NewRouter()

	s := &Server{
		cfg:         cfg,
		Router:      r,
		AuthService: authService,
	}

	r.Use(SubjectMiddleware(authService))

	// WS endpoint
	r.HandleFunc("/ws", handler.WSHandler)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%v:%v", cfg.Host, cfg.Port),
		Handler: r,
	}

	return s
}

func (s *Server) Run() error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
