package grpc

import (
	"context"
	"fmt"

	"github.com/TATAROmangol/mess/shared/logger"
	"github.com/TATAROmangol/mess/shared/requestmeta"
	"github.com/TATAROmangol/mess/tokenissuer/internal/ctxkey"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

const (
	OKMessage      = "OK"
	RequestIDLabel = "request_id"
	MetadataLabel  = "metadata"
)

type Interceptor interface {
	InitLogger(log logger.Logger) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
	SetMetadataWithRequestID(ctx context.Context, eq interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
	LogResponse(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
}

type InterceptorImpl struct {
}

func NewInterceptorImpl(log logger.Logger) Interceptor {
	return &InterceptorImpl{}
}

func (i *InterceptorImpl) InitLogger(log logger.Logger) func(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		ctx = ctxkey.WithLogger(ctx, log)
		return handler(ctx, req)
	}
}

func (i *InterceptorImpl) SetMetadataWithRequestID(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	l, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return nil, fmt.Errorf("logger get from context: %w", err)
	}

	reqID := uuid.NewString()
	l = l.With(RequestIDLabel, reqID)

	metadata := requestmeta.GetFromGRPCRequest(ctx, info)
	l = l.With(MetadataLabel, metadata)

	ctx = ctxkey.WithLogger(ctx, l)
	return handler(ctx, req)
}

func (i *InterceptorImpl) LogResponse(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	l, err := ctxkey.ExtractLogger(ctx)
	if err != nil {
		return nil, fmt.Errorf("logger get from context: %w", err)
	}

	resp, err = handler(ctx, req)
	if err != nil {
		l.Error(err)
	} else {
		l.Info(OKMessage)
	}

	return resp, err
}
