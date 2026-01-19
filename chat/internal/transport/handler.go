package transport

import (
	"errors"
	"net/http"

	"github.com/TATAROmangol/mess/chat/internal/domain"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	domain domain.Service
}

func NewHandler(domain domain.Service) *Handler {
	return &Handler{
		domain: domain,
	}
}

func (h *Handler) GetProfile(c *gin.Context) {

}

func (h *Handler) sendError(c *gin.Context, err error) {
	var code int

	if errors.Is(err, InvalidRequestError) {
		code = http.StatusBadRequest
	}

	if errors.Is(err, domain.ErrNotFound) {
		code = http.StatusNoContent
	}

	if code == 0 {
		code = http.StatusInternalServerError
	}

	c.AbortWithError(code, err)
}
