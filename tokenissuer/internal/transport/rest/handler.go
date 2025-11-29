package rest

import (
	"io"
	"net/http"
	"tokenissuer/internal/adapter/identifier"
	"tokenissuer/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	ErrorLabel = "error"

	RefreshCookieName = "refresh_token"
)

type Handler interface {
	Refresh(c *gin.Context)
	GetToken(c *gin.Context)
}

type HandlerImpl struct {
	service service.Token
}

func NewHandler(src service.Token) *HandlerImpl {
	return &HandlerImpl{
		service: src,
	}
}

func (h *HandlerImpl) sendError(c *gin.Context, code int, err error) {
	if err != io.EOF {
		c.Error(err)
	}

	c.JSON(code, gin.H{
		ErrorLabel: err.Error(),
	})
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (h *HandlerImpl) respondTokenPair(c *gin.Context, tokenPair identifier.TokenPair) {
	c.SetCookie(
		RefreshCookieName,
		tokenPair.GetRefreshToken(),
		tokenPair.GetRefreshExpiresIn(),
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, AccessTokenResponse{
		AccessToken: tokenPair.GetAccessToken(),
		TokenType:   tokenPair.GetTokenType(),
	})
}

func (h *HandlerImpl) Refresh(c *gin.Context) {
	refresh, err := c.Cookie(RefreshCookieName)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, err)
		return
	}

	tokenPair, err := h.service.RefreshTokenPair(c.Request.Context(), refresh)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, err)
		return
	}

	h.respondTokenPair(c, tokenPair)
}

type GetTokenRequest struct {
	Code        string `json:"code"`
	RedirectURL string `json:"redirect_url"`
}

func (h *HandlerImpl) GetToken(c *gin.Context) {
	var req GetTokenRequest

	if err := c.BindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, err)
		return
	}

	tokenPair, err := h.service.GetTokenPair(c.Request.Context(), req.Code, req.RedirectURL)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, err)
		return
	}

	h.respondTokenPair(c, tokenPair)
}
