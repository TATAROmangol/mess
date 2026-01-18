package ctxkey

import (
	"context"
	"fmt"

	"github.com/TATAROmangol/mess/shared/logger"
	"github.com/TATAROmangol/mess/shared/model"
)

type loggerKeyStruct struct{}

var loggerKey = loggerKeyStruct{}

func WithLogger(ctx context.Context, log logger.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

func ExtractLogger(ctx context.Context) (logger.Logger, error) {
	v := ctx.Value(loggerKey)
	if v == nil {
		return nil, fmt.Errorf("not have logger in context")
	}

	l, ok := v.(logger.Logger)
	if !ok {
		return nil, fmt.Errorf("value is not logger: %T", v)
	}

	return l, nil
}

type subjectKeyStruct struct{}

var subjectKey = subjectKeyStruct{}

func WithSubject(ctx context.Context, subj model.Subject) context.Context {
	return context.WithValue(ctx, subjectKey, subj)
}

func ExtractSubject(ctx context.Context) (model.Subject, error) {
	v := ctx.Value(subjectKey)
	if v == nil {
		return nil, fmt.Errorf("not have subject in context")
	}

	s, ok := v.(model.Subject)
	if !ok {
		return nil, fmt.Errorf("value is not subject: %T", v)
	}

	return s, nil
}
