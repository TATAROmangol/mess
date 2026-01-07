package transport

import (
	"fmt"
	"net/http"

	"github.com/TATAROmangol/mess/profile/internal/ctxkey"
	"github.com/TATAROmangol/mess/profile/internal/loglables"
	"github.com/TATAROmangol/mess/shared/logger"
	"github.com/TATAROmangol/mess/shared/requestmeta"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func InitLoggerMiddleware(lg logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := ctxkey.WithLogger(c.Request.Context(), lg)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func SetRequestMetadataMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		meta := requestmeta.GetFromHTTPRequest(c.Request)

		id := uuid.New().String()

		lg, err := ctxkey.ExtractLogger(c.Request.Context())
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("extract logger: %v", err))
			return
		}

		lg = lg.With(loglables.RequestMetadata, *meta)
		lg = lg.With(loglables.RequestID, id)

		ctx := ctxkey.WithLogger(c.Request.Context(), lg)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

type bodyWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func LogResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bw := &bodyWriter{ResponseWriter: c.Writer}
		c.Writer = bw
		c.Next()

		lg, err := ctxkey.ExtractLogger(c.Request.Context())
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("extract logger: %v", err))
			return
		}

		if len(c.Errors) > 0 {
			lg.Error(err)
			return
		}

		lg.With(loglables.StatusResponse, c.Writer.Status())
		lg.With(loglables.Response, string(bw.body))
		lg.Info("request completed")
	}
}
