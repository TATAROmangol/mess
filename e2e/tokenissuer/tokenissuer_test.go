package tokenissuer

import (
	"context"
	pb "e2e/tokenissuer/pb/tokenissuerpb"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Token struct {
	Access string `json:"access_token"`
	Type   string `json:"token_type"`
}

func getTokens(ctx context.Context, t *testing.T) *Token {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client := resty.New()

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"grant_type": "password",
			"client_id":  CFG.Adapter.ClientID,
			"client_secret": CFG.Adapter.ClientSecret,
			"username":   CFG.Adapter.Login,
			"password":   CFG.Adapter.Password,
		}).
		Post(CFG.Adapter.AuthURL)

	if err != nil {
		t.Fatalf("failed to request token: %v", err)
	}

	if resp.StatusCode() != 200 {
		t.Fatalf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.Body())
	}

	var res Token

	if err := json.Unmarshal(resp.Body(), &res); err != nil {
		t.Fatalf("failed to unmarshal token response: %v", err)
	}

	return &res
}

func getSubjectID(ctx context.Context, t *testing.T, client pb.TokenVerifierClient, token Token) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Verify(ctx, &pb.VerifyRequest{
		AccessToken: token.Access,
		TokenType:   token.Type,
	})
	if err != nil {
		return "", err
	}

	return resp.GetSubjectId(), nil
}

func TestTokenIssuer_LoginInServices(t *testing.T) {
	ctx := context.Background()

	clientConn, err := grpc.NewClient(CFG.Issuer.VerifyGrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to create gRPC client: %v", err)
	}
	defer clientConn.Close()
	clientGRPC := pb.NewTokenVerifierClient(clientConn)
	t.Log("gRPC client ready")

	tok := getTokens(ctx, t)

	id, err := getSubjectID(ctx, t, clientGRPC, *tok)
	if err != nil {
		t.Fatalf("failed to get subject ID: %v", err)
	}
	if id != CFG.Adapter.SubjectID {
		t.Fatalf("unexpected subject ID: got %s, want %s", id, CFG.Adapter.SubjectID)
	}

	time.Sleep(CFG.Adapter.TokenDuration)
	_, err = getSubjectID(ctx, t, clientGRPC, *tok)
	if err == nil {
		t.Fatalf("expected error for expired token, got none")
	}
}
