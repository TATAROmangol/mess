package logger

import (
	"log/slog"
)

type Logger interface {
	Info(msg string)
	Error(err error)
	With(key string, val any) Logger
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
