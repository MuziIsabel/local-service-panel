package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken: %v", err)
	}
	// 32 bytes = 64 hex chars
	if len(token) != 64 {
		t.Errorf("token length = %d, want 64", len(token))
	}

	// Should be hex.
	for _, c := range token {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("unexpected char in token: %c", c)
			break
		}
	}

	// Second token should be different.
	token2, _ := generateToken()
	if token == token2 {
		t.Error("two generated tokens should not be equal")
	}
}

func TestLoadOrCreateToken_New(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "auth-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tokenPath := filepath.Join(tmpDir, "config", "token")
	token, err := LoadOrCreateToken(tokenPath)
	if err != nil {
		t.Fatalf("LoadOrCreateToken new: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	// File should exist.
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("read token file: %v", err)
	}
	savedToken := string(data)
	// TrimSpace because we add a newline.
	if savedToken[:len(savedToken)-1] != token {
		t.Errorf("saved token = %q, want %q", savedToken, token)
	}
}

func TestLoadOrCreateToken_Existing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "auth-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tokenPath := filepath.Join(tmpDir, "token")
	existingToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	os.WriteFile(tokenPath, []byte(existingToken+"\n"), 0600)

	token, err := LoadOrCreateToken(tokenPath)
	if err != nil {
		t.Fatalf("LoadOrCreateToken existing: %v", err)
	}
	if token != existingToken {
		t.Errorf("token = %q, want %q", token, existingToken)
	}
}

func TestTokenPath(t *testing.T) {
	got := TokenPath("/data")
	want := filepath.Join("/data", "config", "token")
	if got != want {
		t.Errorf("TokenPath(/data) = %q, want %q", got, want)
	}
}

func TestMiddleware_SkipPaths(t *testing.T) {
	skipPaths := map[string]bool{
		"/api/healthz": true,
	}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	handler := Middleware("secret-token", skipPaths, "", inner)

	// healthz should pass without token.
	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("healthz without token: status = %d, want 200", w.Code)
	}
}

func TestMiddleware_ValidToken(t *testing.T) {
	skipPaths := map[string]bool{}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	handler := Middleware("secret-token", skipPaths, "", inner)

	// Valid token should pass.
	req := httptest.NewRequest(http.MethodGet, "/api/targets", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("valid token: status = %d, want 200", w.Code)
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	skipPaths := map[string]bool{}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Middleware("secret-token", skipPaths, "", inner)

	// Invalid token should be rejected.
	req := httptest.NewRequest(http.MethodGet, "/api/targets", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("invalid token: status = %d, want 401", w.Code)
	}
}

func TestMiddleware_NoAuth(t *testing.T) {
	skipPaths := map[string]bool{}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Middleware("secret-token", skipPaths, "", inner)

	// No auth header.
	req := httptest.NewRequest(http.MethodGet, "/api/targets", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("no auth: status = %d, want 401", w.Code)
	}
}

func TestMiddleware_DevToken(t *testing.T) {
	skipPaths := map[string]bool{}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Middleware("secret-token", skipPaths, "dev-token-123", inner)

	// Dev token should be accepted.
	req := httptest.NewRequest(http.MethodGet, "/api/targets", nil)
	req.Header.Set("Authorization", "Bearer dev-token-123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("dev token: status = %d, want 200", w.Code)
	}

	// Random token should still be rejected.
	req2 := httptest.NewRequest(http.MethodGet, "/api/targets", nil)
	req2.Header.Set("Authorization", "Bearer random")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Errorf("random token: status = %d, want 401", w2.Code)
	}
}
