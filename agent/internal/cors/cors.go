// Package cors provides CORS middleware for HTTP handlers.
package cors

import (
	"net/http"
	"strings"
)

// AllowedOrigins is the list of allowed CORS origins.
var AllowedOrigins = []string{
	"http://localhost:5173",
	"http://127.0.0.1:5173",
	"tauri://localhost",
	"http://localhost:1420",
}

// Middleware returns an HTTP handler that adds CORS headers.
// If the request's Origin matches an allowed origin, it is echoed back.
// Preflight OPTIONS requests are handled directly with a 204 response.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		// Handle preflight requests.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(origin string) bool {
	for _, allowed := range AllowedOrigins {
		if strings.EqualFold(origin, allowed) {
			return true
		}
	}
	return false
}
