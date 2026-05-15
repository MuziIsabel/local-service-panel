// Package auth handles Bearer token generation, loading, and HTTP middleware.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	// tokenBytes is the number of random bytes for the token (32 bytes = 64 hex chars).
	tokenBytes = 32

	// Header prefix for Bearer token.
	bearerPrefix = "Bearer "
)

// TokenPath returns the full path to the token file within the data directory.
func TokenPath(dataDir string) string {
	return filepath.Join(dataDir, "config", "token")
}

// LoadOrCreateToken reads the token from file, or generates and saves a new one.
// The token is a hex-encoded string of 32 random bytes.
func LoadOrCreateToken(path string) (string, error) {
	// Try to read existing token.
	data, err := os.ReadFile(path)
	if err == nil {
		token := strings.TrimSpace(string(data))
		if token != "" {
			return token, nil
		}
	}

	// Generate new token.
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("create token directory: %w", err)
	}

	// Write token file.
	if err := os.WriteFile(path, []byte(token+"\n"), 0600); err != nil {
		return "", fmt.Errorf("write token file: %w", err)
	}

	return token, nil
}

// generateToken creates a cryptographically random hex string.
func generateToken() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// Middleware returns an HTTP middleware that validates Bearer tokens.
// Paths in skipPaths are allowed without authentication (e.g. "/api/healthz").
// devToken, if non-empty, is also accepted (for development convenience).
func Middleware(token string, skipPaths map[string]bool, devToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for whitelisted paths.
		if skipPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// Extract Authorization header.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeUnauthorized(w, "Missing Authorization header")
			return
		}

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			writeUnauthorized(w, "Invalid Authorization header format")
			return
		}

		providedToken := strings.TrimPrefix(authHeader, bearerPrefix)

		// Check against real token and optional dev token.
		if providedToken != token && (devToken == "" || providedToken != devToken) {
			writeUnauthorized(w, "Invalid token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(fmt.Sprintf(
		`{"error":{"code":"UNAUTHORIZED","message":"%s"}}`, message,
	)))
}
