package transport

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/1ocknight/mess/chat/internal/domain"
	"github.com/1ocknight/mess/shared/logger"
	"github.com/1ocknight/mess/shared/verify"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Config struct {
	CorsUrl   []string `yaml:"cors_url"`
	Host      string   `yaml:"host"`
	Port      string   `yaml:"port"`
	DebugMode bool     `yaml:"debug_mode"`
}

type HTTPServer struct {
	cfg    *Config
	srv    *gin.Engine
	httpSv *http.Server
}

func NewServer(cfg Config, lg logger.Logger, domain domain.Service, verify verify.Service) *HTTPServer {
	h := NewHandler(domain)

	if !cfg.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	if len(cfg.CorsUrl) != 0 {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.CorsUrl,
			AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}

	r.Use(InitLoggerMiddleware(lg))
	r.Use(SetRequestMetadataMiddleware())
	r.Use(LogResponseMiddleware())
	r.Use(InitSubjectMiddleware(verify))

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
