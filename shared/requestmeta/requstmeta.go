package requestmeta

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func GetFromHTTPRequest(r *http.Request) *HttpMetadata {
	return &HttpMetadata{
		Method:    r.Method,
		URL:       r.URL.String(),
		ClientIP:  r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}
}

func GetFromGRPCRequest(ctx context.Context, info *grpc.UnaryServerInfo) *GrpcMetadata {
	p, ok := peer.FromContext(ctx)
	clientAddr := ""
	if ok {
		clientAddr = p.Addr.String() 
	}

	return &GrpcMetadata{
		Method:   info.FullMethod,
		PeerAddr: clientAddr,
	}
}
