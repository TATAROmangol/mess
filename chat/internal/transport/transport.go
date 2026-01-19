package transport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/TATAROmangol/mess/chat/internal/domain"
	"github.com/TATAROmangol/mess/shared/auth"
	"github.com/TATAROmangol/mess/shared/logger"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Host      string `yaml:"host"`
	Port      string `yaml:"port"`
	DebugMode bool   `yaml:"debug_mode"`
}

type HTTPServer struct {
	cfg    *Config
	srv    *gin.Engine
	httpSv *http.Server
}

func NewServer(cfg Config, lg logger.Logger, domain domain.Service, auth auth.Service) *HTTPServer {
	h := NewHandler(domain)

	if !cfg.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(InitLoggerMiddleware(lg))
	r.Use(SetRequestMetadataMiddleware())
	r.Use(LogResponseMiddleware())
	r.Use(InitSubjectMiddleware(auth))

	r.GET("/chat/subject/:subject_id", h.GetChatBySubjectID)
	r.POST("/chat/subject/:subject_id", h.AddChat)
	r.GET("/chat/:chat_id", h.GetChatByID)
	r.GET("/chats", h.GetChats)

	r.GET("/messages", h.GetMessages)
	r.POST("/message", h.AddMessage)
	r.PATCH("/message", h.UpdateMessage)

	r.PATCH("/lastread", h.UpdateLastRead)

	return &HTTPServer{
		cfg: &cfg,
		srv: r,
	}
}

func (s *HTTPServer) Run() error {
	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)

	s.httpSv = &http.Server{
		Addr:    addr,
		Handler: s.srv,
	}

	return s.httpSv.ListenAndServe()
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.httpSv.Shutdown(ctx)
}
