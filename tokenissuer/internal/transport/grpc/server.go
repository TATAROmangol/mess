package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Server struct {
	cfg    Config
	server *grpc.Server
}

func NewServer(cfg Config, interceptor Interceptor, svc Handler) *Server {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.SetMethodName,
			interceptor.SetRequestID,
			interceptor.Loggining,
		),
	)

	Register(srv, svc)
	reflection.Register(srv)

	return &Server{
		cfg:    cfg,
		server: srv,
	}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%v", s.cfg.Host, s.cfg.Port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	s.server.Stop()
}
