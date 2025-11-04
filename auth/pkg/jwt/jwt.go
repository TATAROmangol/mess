package jwt

import (
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type Service interface {
	GenerateAccessTokenWithUserID(userID string) (string, error)
	GenerateRefreshTokenWithUserID(userID string) (string, error)
	GetUserIDFromToken(token string) (string, error)
}

type Config struct {
	SecretKey       string        `json:"secret_key"`
	AccessTokenTTL  time.Duration `json:"access_token_ttl"`
	RefreshTokenTTL time.Duration `json:"refresh_token_ttl"`
}

type JWT struct {
	cfg Config
}

func New(cfg Config) *JWT {
	return &JWT{cfg: cfg}
}

const (
	SubjectKey = "sub"
	ExpiredKey = "exp"
)

func (j *JWT) GenerateAccessTokenWithUserID(userID string) (string, error) {
	claims := jwtv5.MapClaims{
		SubjectKey: userID,
		ExpiredKey: time.Now().Add(j.cfg.AccessTokenTTL).Unix(),
	}

	return j.generateTokensWithClaims(claims)
}

func (j *JWT) GenerateRefreshTokenWithUserID(userID string) (string, error) {
	claims := jwtv5.MapClaims{
		SubjectKey: userID,
		ExpiredKey: time.Now().Add(j.cfg.RefreshTokenTTL).Unix(),
	}

	return j.generateTokensWithClaims(claims)
}

func (j *JWT) generateTokensWithClaims(claims jwtv5.MapClaims) (string, error) {
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.cfg.SecretKey))
}

func (j *JWT) GetUserIDFromToken(token string) (string, error) {
	parsedToken, err := jwtv5.Parse(token, func(t *jwtv5.Token) (any, error) {
		return []byte(j.cfg.SecretKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("jwt parse: %w", err)
	}

	return parsedToken.Claims.GetSubject()
}
