package transport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/TATAROmangol/mess/profile/internal/domain"
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

	r.GET("/profile", h.GetProfile)
	r.GET("/profile/:id", h.GetProfile)
	r.GET("/profiles/:alias", h.GetProfiles)

	r.POST("/add/profile", h.AddProfile)

	r.PUT("/put/profile", h.UpdateProfileMetadata)
	r.PUT("/upload/avatar", h.UploadAvatar)

	r.DELETE("/delete/avatar", h.DeleteAvatar)
	r.DELETE("/delete/profile", h.DeleteProfile)

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
