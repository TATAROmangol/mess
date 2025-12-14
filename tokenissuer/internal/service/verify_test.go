package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"reflect"
	"sync"
	"testing"
	"time"
	jwksloadermocks "tokenissuer/internal/adapter/jwksloader/mocks"
	"tokenissuer/internal/model"
	"tokenissuer/pkg/jwks"
	jwksmocks "tokenissuer/pkg/jwks/mocks"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
)

const (
	KID       = "test_kid"
	SecondKID = "second_test_kid"
)

func initTestKeyAndToken(t *testing.T, kid string, claims jwt.MapClaims) (*rsa.PrivateKey, *rsa.PublicKey, string) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("cannot generate rsa: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signedToken, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("cannot sign token: %v", err)
	}

	return priv, &priv.PublicKey, signedToken
}

func TestVerifyImpl_updateJWKSKeys(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	jwksMock := jwksmocks.NewMockJWKS(ctrl)

	jwksLoader := jwksloadermocks.NewMockService(ctrl)

	jwksLoader.EXPECT().LoadJWKS(gomock.Any()).Return(
		map[string]jwks.JWKS{"kid1": jwksMock},
		nil,
	).Times(2)

	cfg := VerifyConfig{JwksRateLimit: 500 * time.Millisecond}
	verify, err := NewVerifyImpl(ctx, jwksLoader, cfg)
	if err != nil {
		t.Fatal(err)
	}

	verify.jwksLastUpdated = time.Now().Add(-1 * time.Hour)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	start := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start
			if err := verify.updateJWKSKeys(ctx); err != nil {
				t.Errorf("updateJWKSKeys() error = %v", err)
			}
		}()
	}

	close(start)
	wg.Wait()
}

func TestVerifyImpl_findKeyByKid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, fKey, _ := initTestKeyAndToken(t, KID, jwt.MapClaims{})
	_, sKey, _ := initTestKeyAndToken(t, SecondKID, jwt.MapClaims{})

	fJwksMock := jwksmocks.NewMockJWKS(ctrl)
	fJwksMock.EXPECT().GetPublicKey().Return(fKey, nil).AnyTimes()

	sJwksMock := jwksmocks.NewMockJWKS(ctrl)
	sJwksMock.EXPECT().GetPublicKey().Return(sKey, nil).AnyTimes()

	jwksLoader := jwksloadermocks.NewMockService(ctrl)

	firstCall := jwksLoader.EXPECT().LoadJWKS(gomock.Any()).Return(
		map[string]jwks.JWKS{
			KID: fJwksMock,
		},
		nil,
	)
	jwksLoader.EXPECT().LoadJWKS(gomock.Any()).Return(
		map[string]jwks.JWKS{
			KID:       fJwksMock,
			SecondKID: sJwksMock,
		},
		nil,
	).After(firstCall).AnyTimes()

	cfg := VerifyConfig{JwksRateLimit: 1 * time.Microsecond}
	verify, err := NewVerifyImpl(context.Background(), jwksLoader, cfg)
	if err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	ctx := context.Background()

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			<-start
			defer wg.Done()
			kid := KID
			if i%2 != 0 {
				kid = SecondKID
			}
			_, _ = verify.findKeyByKid(ctx, kid)
		}(i)
	}

	close(start)

	wg.Wait()
}

func TestVerifyImpl_VerifyToken(t *testing.T) {
	claims := jwt.MapClaims{
		"sub":                "123",
		"email":              "test@example.com",
		"preferred_username": "tester",
	}

	_, key, token := initTestKeyAndToken(t, KID, claims)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwksMock := jwksmocks.NewMockJWKS(ctrl)
	jwksMock.EXPECT().GetPublicKey().Return(key, nil).AnyTimes()

	jwksLoader := jwksloadermocks.NewMockService(ctrl)
	jwksLoader.EXPECT().LoadJWKS(gomock.Any()).Return(
		map[string]jwks.JWKS{
			KID: jwksMock,
		},
		nil,
	).AnyTimes()

	cfg := VerifyConfig{JwksRateLimit: time.Hour}
	verify, err := NewVerifyImpl(context.Background(), jwksLoader, cfg)
	if err != nil {
		t.Error(err)
	}

	type args struct {
		ctx         context.Context
		typeToken   string
		accessToken string
	}
	tests := []struct {
		name    string
		args    args
		want    *model.User
		wantErr bool
	}{
		{
			name: "valid token",
			args: args{
				ctx:         context.Background(),
				typeToken:   BearerType,
				accessToken: token,
			},
			want: &model.User{
				ID:    "123",
				Name:  "tester",
				Email: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid token",
			args: args{
				ctx:         context.Background(),
				typeToken:   BearerType,
				accessToken: "234",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid token type",
			args: args{
				ctx:         context.Background(),
				typeToken:   "InvalidType",
				accessToken: token,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wrong kid",
			args: args{
				ctx:       context.Background(),
				typeToken: BearerType,
				accessToken: func() string {
					_, _, tok := initTestKeyAndToken(t, "wrong_kid", claims)
					return tok
				}(),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := verify.VerifyToken(tt.args.ctx, tt.args.typeToken, tt.args.accessToken)
			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifyToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VerifyToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
