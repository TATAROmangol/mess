package grpc

import (
	"context"
	"fmt"
	"tokenissuer/internal/service"
	pb "tokenissuer/internal/transport/grpc/pb/tokenissuerpb"

	"google.golang.org/grpc"
)

type Handler interface {
	pb.TokenVerifierServer
	Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error)
}

type HandlerImpl struct {
	pb.UnimplementedTokenVerifierServer
	srv service.Verify
}

func NewHandlerImpl(srv service.Verify) *HandlerImpl {
	return &HandlerImpl{
		srv: srv,
	}
}

func Register(gRPCServer *grpc.Server, handler Handler) {
	pb.RegisterTokenVerifierServer(gRPCServer, handler)
}

func (h *HandlerImpl) Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	if req.GetTokenType() == "" {
		return nil, fmt.Errorf("not have token type")
	}
	if req.GetAccessToken() == "" {
		return nil, fmt.Errorf("not have access token")
	}

	user, err := h.srv.VerifyToken(ctx, req.GetTokenType(), req.GetAccessToken())
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}

	return &pb.VerifyResponse{
		SubjectId: user.ID,
		Name:      user.Name,
		Email:     user.Email,
	}, nil
}
