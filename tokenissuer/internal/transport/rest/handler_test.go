package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	identifiermocks "tokenissuer/internal/adapter/identifier/mocks"
	servicemocks "tokenissuer/internal/service/mocks"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlerImpl_Refresh_SetsCookie(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSrv := servicemocks.NewMockToken(ctrl)

	tokenPair := identifiermocks.NewMockTokenPair(ctrl)
	tokenPair.EXPECT().GetRefreshToken().Return("refresh123").AnyTimes()
	tokenPair.EXPECT().GetRefreshExpiresIn().Return(3600).AnyTimes()
	tokenPair.EXPECT().GetAccessToken().Return("access123").AnyTimes()
	tokenPair.EXPECT().GetTokenType().Return("Bearer").AnyTimes()

	mockSrv.EXPECT().
		RefreshTokenPair(gomock.Any(), "refresh_cookie").
		Return(tokenPair, nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewHandler(mockSrv)
	r.GET("/refresh", handler.Refresh)

	req := httptest.NewRequest(http.MethodGet, "/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  RefreshCookieName,
		Value: "refresh_cookie",
	})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == RefreshCookieName && c.Value == "refresh123" {
			found = true
			break
		}
	}
	assert.True(t, found, "Refresh cookie was not set correctly")

	expected := `{"access_token":"access123","token_type":"Bearer"}`
	assert.JSONEq(t, expected, w.Body.String())
}

func TestHandlerImpl_GetToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSrv := servicemocks.NewMockToken(ctrl)

	tokenPair := identifiermocks.NewMockTokenPair(ctrl)
	tokenPair.EXPECT().GetRefreshToken().Return("refresh123").AnyTimes()
	tokenPair.EXPECT().GetRefreshExpiresIn().Return(3600).AnyTimes()
	tokenPair.EXPECT().GetAccessToken().Return("access123").AnyTimes()
	tokenPair.EXPECT().GetTokenType().Return("Bearer").AnyTimes()

	mockSrv.EXPECT().
		GetTokenPair(gomock.Any(), "auth_code", "https://redirect").
		Return(tokenPair, nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewHandler(mockSrv)
	r.POST("/token", handler.GetToken)

	body := `{"code":"auth_code","redirect_url":"https://redirect"}`
	req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == RefreshCookieName && c.Value == "refresh123" {
			found = true
			break
		}
	}
	assert.True(t, found, "Refresh cookie was not set correctly")

	expected := `{"access_token":"access123","token_type":"Bearer"}`
	assert.JSONEq(t, expected, w.Body.String())
}
