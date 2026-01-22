package transport

import (
	"fmt"
	"net/http"

	"github.com/TATAROmangol/mess/shared/auth"
	"github.com/TATAROmangol/mess/websocket/internal/ctxkey"
)

func SubjectMiddleware(auth auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")

			sub, err := auth.Verify(token)
			if err != nil {
				http.Error(w, fmt.Sprintf("verify token: %v", err), http.StatusUnauthorized)
				return
			}

			ctx := ctxkey.WithSubject(r.Context(), sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
