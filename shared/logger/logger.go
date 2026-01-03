package logger

import (
	"log/slog"
	"strings"
)

type Logger interface {
	Info(msg string)
	Error(err error)
	Errors(msg string, errors []error)
	With(key string, val any) Logger
}

type log struct {
	lg *slog.Logger
}

func New(handler slog.Handler) Logger {
	lg := slog.New(handler)
	return &log{
		lg: lg,
	}
}

func (l *log) Info(msg string) {
	l.lg.Info(msg)
}

func (l *log) Error(err error) {
	l.lg.Error(err.Error())
}

func (l *log) Errors(msg string, errors []error) {
	var b strings.Builder
	for i, err := range errors {
		if err != nil {
			if i > 0 {
				b.WriteString("; ")
			}
			b.WriteString(err.Error())
		}
	}

	l.lg.Error(msg, slog.String("errors", b.String()))
}

func (l *log) With(key string, val any) Logger {
	return &log{
		lg: l.lg.With(key, val),
	}
}
