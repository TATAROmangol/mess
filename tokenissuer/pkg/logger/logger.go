package logger

import (
	"context"
	"io"
	"log/slog"
)

const (
	OkMessage = "OK"
)

type Logger interface {
	InfoContext(ctx context.Context, msg string)
	ErrorContext(ctx context.Context, err error)
}

type Log struct {
	slog  *slog.Logger
	parse func(ctx context.Context) map[string]any
}

func New(w io.Writer, parseFunc func(ctx context.Context) map[string]any) *Log {
	handler := slog.NewTextHandler(w, nil)
	return &Log{
		slog:  slog.New(handler),
		parse: parseFunc,
	}
}

func (l *Log) InfoContext(ctx context.Context, msg string) {
	attrs := []slog.Attr{}
	for k, v := range l.parse(ctx) {
		attrs = append(attrs, slog.Any(k, v))
	}
	l.slog.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

func (l *Log) ErrorContext(ctx context.Context, err error) {
	attrs := []slog.Attr{}
	for k, v := range l.parse(ctx) {
		attrs = append(attrs, slog.Any(k, v))
	}
	l.slog.LogAttrs(ctx, slog.LevelError, err.Error(), attrs...)
}
