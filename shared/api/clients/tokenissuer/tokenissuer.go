package tokenissuer

import (
	"context"
	"fmt"

	pb "github.com/TATAROmangol/mess/shared/api/pb/tokenissuer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Subject interface {
	GetSubjectId() string
	GetName() string
	GetEmail() string
}

type Client interface {
	Close() error
	Verify(ctx context.Context, token string, tokenType string) (Subject, error)
}

type ClientIMPL struct {
	conn *grpc.ClientConn
	c    pb.TokenVerifierClient
}

func NewClient(cfg Config) (Client, error) {
	addr := fmt.Sprintf("%v:%v", cfg.Host, cfg.Port)
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	c := pb.NewTokenVerifierClient(conn)

	client := ClientIMPL{
		conn: conn,
		c:    c,
	}

	return &client, nil
}

func (c *ClientIMPL) Close() error {
	return c.conn.Close()
}

func (c *ClientIMPL) Verify(ctx context.Context, token string, tokenType string) (Subject, error) {
	req := pb.VerifyRequest{
		AccessToken: token,
		TokenType:   tokenType,
	}

	resp, err := c.c.Verify(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("verify request failed: %w", err)
	}

	return resp, nil
}
