package logger

import (
	"context"
	"fmt"
	"log/slog"
)

type Logger interface {
	Info(msg string)
	Error(err error)
	With(key string, val any) Logger
	PushFromContext(ctx context.Context) context.Context
}

type loggerCtxKey struct{}

var LoggerKey = loggerCtxKey{}

type Log struct {
	lg *slog.Logger
}

func New(handler slog.Handler) *Log {
	lg := slog.New(handler)
	return &Log{
		lg: lg,
	}
}

func (l *Log) Info(msg string) {
	l.lg.Info(msg)
}

func (l *Log) Error(err error) {
	l.lg.Error(err.Error())
}

func (l *Log) With(key string, val any) Logger {
	return &Log{
		lg: l.lg.With(slog.Any(key, val)),
	}
}

func (l *Log) PushFromContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, LoggerKey, l)
}

func GetFromContext(ctx context.Context) (*Log, error) {
	lg, ok := ctx.Value(LoggerKey).(*Log)
	if !ok {
		return nil, fmt.Errorf("logger not found in context")
	}
	return lg, nil
}
