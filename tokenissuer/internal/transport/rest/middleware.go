package rest

import (
	"tokenissuer/internal/ctxkey"
	"tokenissuer/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Middleware interface {
	SetMethodName() gin.HandlerFunc
	SetRequestID() gin.HandlerFunc
	SetPath() gin.HandlerFunc
	Log() gin.HandlerFunc
}

type MiddlewareImpl struct {
	log logger.Logger
}

func NewMiddleware(log logger.Logger) *MiddlewareImpl {
	return &MiddlewareImpl{
		log: log,
	}
}

func (m *MiddlewareImpl) SetMethodName() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := ctxkey.WithMethodName(c.Request.Context(), c.Request.Method)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (m *MiddlewareImpl) SetPath() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := ctxkey.WithPath(c.Request.Context(), c.Request.URL.String())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (m *MiddlewareImpl) SetRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := uuid.NewString()
		ctx := ctxkey.WithRequestID(c.Request.Context(), reqID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (m *MiddlewareImpl) Log() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			m.log.InfoContext(c.Request.Context(), logger.OkMessage)
			return
		}

		for _, e := range c.Errors {
			m.log.ErrorContext(c.Request.Context(), e.Err)
		}
	}
}
