package tokenissuer

import (
	"context"
	pb "e2e/tokenissuer/pb/tokenissuerpb"
	"encoding/json"
	"fmt"
	"net/http/cookiejar"
	"net/url"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/go-rod/rod"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	ScopeOpenID      = "openid"
	ResponseTypeCode = "code"
	State            = "e2e"
	RedirectURL      = "e2e"
)

func getSubjectID(ctx context.Context, t *testing.T, client pb.TokenVerifierClient, token Token) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Verify(ctx, &pb.VerifyRequest{
		AccessToken: token.Access,
		TokenType:   token.Type,
	})
	if err != nil {
		t.Fatalf("gRPC Verify call failed: %v", err)
	}

	return resp.GetSubjectId()
}

func rodRequest(t *testing.T) string {
	// Собираем полный authURL
	u, _ := url.Parse(CFG.Adapter.AuthURL)
	params := url.Values{}
	params.Set("client_id", CFG.Adapter.ClientID)
	params.Set("redirect_uri", RedirectURL)
	params.Set("response_type", ResponseTypeCode)
	params.Set("scope", ScopeOpenID)
	params.Set("state", State)
	u.RawQuery = params.Encode()
	authURL := u.String()

	browser := rod.New().Timeout(15 * time.Second).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(authURL).Timeout(15 * time.Second)
	defer page.MustClose()

	// Вводим логин и пароль
	page.MustElement(`input[name="username"]`).MustInput(CFG.Adapter.Login)
	page.MustElement(`input[name="password"]`).MustInput(CFG.Adapter.Password)
	page.MustElement(`input[type="submit"]`).MustClick()

	page.MustWaitLoad()

	finalURL := page.MustEval(`() => window.location.href`).Str()
	t.Log("Final URL after login:", finalURL)

	return finalURL
}

type Token struct {
	Access string `json:"access_token"`
	Type   string `json:"token_type"`
}

func exchangeCodeForToken(ctx context.Context, t *testing.T, client *resty.Client, code string, redirectURL string) Token {
	verifyResp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"code":         code,
			"redirect_url": redirectURL,
		}).
		Post(CFG.Issuer.VerifyCodeEndpoint)

	if err != nil {
		t.Fatalf("failed to call VerifyCodeEndpoint: %v", err)
	}

	var tokenResp Token
	if err := json.Unmarshal(verifyResp.Body(), &tokenResp); err != nil {
		t.Fatalf("failed to unmarshal token response: %v", err)
	}

	return tokenResp
}

func refreshToken(ctx context.Context, t *testing.T, client *resty.Client) Token {
	refreshResp, err := client.R().
		SetHeader("Content-Type", "application/json").
		Post(CFG.Issuer.RefreshEndpoint)

	if err != nil {
		t.Fatalf("failed to call RefreshEndpoint: %v", err)
	}

	var tokenResp Token
	if err := json.Unmarshal(refreshResp.Body(), &tokenResp); err != nil {
		t.Fatalf("failed to unmarshal refresh token response: %v", err)
	}

	return tokenResp
}

func TestTokenIssuer_LoginInServices(t *testing.T) {
	ctx := context.Background()

	// gRPC клиент
	clientConn, err := grpc.NewClient(CFG.Issuer.VerifyGrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to create gRPC client: %v", err)
	}
	defer clientConn.Close()
	clientGRPC := pb.NewTokenVerifierClient(clientConn)
	t.Log("gRPC client ready")

	// REST клиент
	jar, _ := cookiejar.New(nil)
	clientREST := resty.New()
	clientREST.SetCookieJar(jar)
	t.Log("REST client ready")

	// авторизация через rod
	finalURL := rodRequest(t)

	// Если сервер OAuth возвращает код через URL, можно его выдернуть
	redirectURL, _ := url.Parse(finalURL)
	code := redirectURL.Query().Get("code")
	if code == "" {
		t.Fatal("expected code from URL, got empty")
	}
	t.Logf("Received code: %s", code)

	// обмен кода на токен
	tokenResp := exchangeCodeForToken(ctx, t, clientREST, code, RedirectURL)
	fmt.Println(tokenResp.Type)
	t.Logf("Received token: access=%s", tokenResp.Access)

	// проверка токена через gRPC
	firstSubjID := getSubjectID(ctx, t, clientGRPC, tokenResp)
	t.Logf("Subject ID before refresh: %s", firstSubjID)

	// рефреш
	tokenResp = refreshToken(ctx, t, clientREST)
	t.Logf("Token refreshed: access=%s", tokenResp.Access)

	// проверка токена после рефреша
	secondSubjID := getSubjectID(ctx, t, clientGRPC, tokenResp)
	t.Logf("Subject ID after refresh: %s", secondSubjID)

	if firstSubjID != secondSubjID {
		t.Fatalf("Subject ID mismatch: %s vs %s", firstSubjID, secondSubjID)
	}

	t.Log("Test completed successfully")
}
