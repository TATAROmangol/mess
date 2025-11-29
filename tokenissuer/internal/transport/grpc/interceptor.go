package grpc

import (
	"context"
	"tokenissuer/internal/ctxkey"
	"tokenissuer/pkg/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Interceptor interface {
	SetRequestID(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
	SetMethodName(ctx context.Context, eq interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
	Loggining(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
}

type InterceptorImpl struct {
	log logger.Logger
}

func NewInterceptorImpl(log logger.Logger) *InterceptorImpl {
	return &InterceptorImpl{
		log: log,
	}
}

func (i *InterceptorImpl) SetRequestID(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	reqID := uuid.NewString()
	ctx = ctxkey.WithRequestID(ctx, reqID)
	return handler(ctx, req)
}

func (i *InterceptorImpl) SetMethodName(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	ctx = ctxkey.WithMethodName(ctx, info.FullMethod)
	return handler(ctx, req)
}

func (i *InterceptorImpl) Loggining(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		i.log.ErrorContext(ctx, err)
	} else {
		i.log.InfoContext(ctx, logger.OkMessage)
	}

	return resp, err
}
