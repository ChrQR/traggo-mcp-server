package mcp

import (
	"context"
	"net/http"
	"strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Unauthorized: Invalid header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Inject the token into the request context.
		// The MCP SDK will automatically pass this context down to your Tool functions.
		ctx := context.WithValue(r.Context(), "traggo_token", token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
