package jwt_test

import (
	"auth/pkg/jwt"
	"strings"
	"testing"
	"time"
)

var TestConfig = jwt.Config{
	SecretKey:       "test_secret_key",
	AccessTokenTTL:  time.Minute * 15,
	RefreshTokenTTL: time.Hour * 24 * 7,
}

func TestJWT_GenerateAccessTokenWithUserID(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{"valid userID 12345", "12345", false},
		{"valid userID abcde", "abcde", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := jwt.New(TestConfig)
			token, err := j.GenerateAccessTokenWithUserID(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GenerateAccessTokenWithUserID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if token == "" {
				t.Fatal("GenerateAccessTokenWithUserID() returned empty token")
			}

			parts := strings.Split(token, ".")
			if len(parts) != 3 {
				t.Fatalf("GenerateAccessTokenWithUserID() returned invalid JWT: %v", token)
			}

			gotUserID, err := j.GetUserIDFromToken(token)
			if err != nil {
				t.Fatalf("GetUserIDFromToken() error = %v", err)
			}
			if gotUserID != tt.userID {
				t.Errorf("GetUserIDFromToken() = %v, want %v", gotUserID, tt.userID)
			}
		})
	}
}

func TestJWT_GenerateRefreshTokenWithUserID(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{"valid userID 54321", "54321", false},
		{"valid userID xyz", "xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := jwt.New(TestConfig)
			token, err := j.GenerateRefreshTokenWithUserID(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GenerateRefreshTokenWithUserID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if token == "" {
				t.Fatal("GenerateRefreshTokenWithUserID() returned empty token")
			}

			parts := strings.Split(token, ".")
			if len(parts) != 3 {
				t.Fatalf("GenerateRefreshTokenWithUserID() returned invalid JWT: %v", token)
			}

			gotUserID, err := j.GetUserIDFromToken(token)
			if err != nil {
				t.Fatalf("GetUserIDFromToken() error = %v", err)
			}
			if gotUserID != tt.userID {
				t.Errorf("GetUserIDFromToken() = %v, want %v", gotUserID, tt.userID)
			}
		})
	}
}

func TestJWT_GetUserIDFromToken(t *testing.T) {
	j := jwt.New(TestConfig)

	validToken, _ := j.GenerateAccessTokenWithUserID("123")

	tests := []struct {
		name    string
		token   string
		want    string
		wantErr bool
	}{
		{"valid token", validToken, "123", false},
		{"invalid token", "invalid.token.value", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := j.GetUserIDFromToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetUserIDFromToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("GetUserIDFromToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
