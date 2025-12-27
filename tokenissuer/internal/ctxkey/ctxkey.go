package ctxkey

import (
	"context"
	"fmt"

	"github.com/TATAROmangol/mess/shared/logger"
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
