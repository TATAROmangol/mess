package tokenissuer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-resty/resty/v2"
)

func TestTokenIssuer_LoginWithChromedp(t *testing.T) {
	ch := make(chan string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "code not provided", http.StatusBadRequest)
			return
		}
		ch <- code
		fmt.Fprintf(w, "Received code: %s", code)
	}))
	defer server.Close()

	u, _ := url.Parse(CFG.Adapter.AuthURL)
	params := url.Values{}
	params.Set("client_id", CFG.Adapter.ClientID)
	params.Set("redirect_uri", server.URL)
	params.Set("response_type", "code")
	params.Set("scope", CFG.Adapter.Scope)
	params.Set("state", "xyz123")
	u.RawQuery = params.Encode()

	authURL := u.String()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var finalURL string

	err := chromedp.Run(ctx,
		chromedp.Navigate(authURL),
		chromedp.WaitVisible(`input[name="username"]`),
		chromedp.SendKeys(`input[name="username"]`, CFG.Adapter.Login),
		chromedp.SendKeys(`input[name="password"]`, CFG.Adapter.Password),
		chromedp.Click(`input[type="submit"]`),
		chromedp.WaitVisible(`body`),
		chromedp.Location(&finalURL),
	)
	if err != nil {
		t.Fatalf("chromedp run failed: %v", err)
	}

	t.Log("Final redirected URL:", finalURL)

	codeCh := <-ch
	close(ch)
	t.Log("Authorization code received:", codeCh)

	if codeCh == "" {
		t.Fatal("expected code from callback, got empty")
	}

	client := resty.New()
	verifyResp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"code":         codeCh,
			"redirect_url": server.URL,
		}).
		Post(CFG.VerifyCodeEndpoint)

	if err != nil {
		t.Fatalf("failed to call VerifyCodeEndpoint: %v", err)
	}

	t.Log("VerifyCodeEndpoint response:", verifyResp.String())
}
