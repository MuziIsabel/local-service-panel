package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/local-service-panel/agent/internal/db"
	"github.com/user/local-service-panel/agent/internal/db/repository"
	"github.com/user/local-service-panel/agent/internal/domain"
	"github.com/user/local-service-panel/agent/internal/events"
	"github.com/user/local-service-panel/agent/internal/settings"
	"github.com/user/local-service-panel/agent/internal/windowsservice"
)

// testEventSvc is a shared events.Service backed by a temp database for all tests.
var testEventSvc *events.Service

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "lsp-api-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create temp dir: %v\n", err)
		os.Exit(1)
	}

	database, err := db.Open(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "open database: %v\n", err)
		os.RemoveAll(tmpDir)
		os.Exit(1)
	}

	repo := repository.NewEventLogRepo(database)
	testEventSvc = events.NewService(repo)

	code := m.Run()

	database.Close()
	os.RemoveAll(tmpDir)
	os.Exit(code)
}

// mockProvider implements windowsservice.Provider for testing.
type mockProvider struct {
	listFn          func(ctx context.Context, filter windowsservice.Filter) ([]domain.Service, error)
	getFn           func(ctx context.Context, serviceName string) (*domain.Service, error)
	startFn         func(ctx context.Context, serviceName string) error
	stopFn          func(ctx context.Context, serviceName string) error
	restartFn       func(ctx context.Context, serviceName string) error
	setStartTypeFn  func(ctx context.Context, serviceName string, startType domain.StartType) error
}

func (m *mockProvider) List(ctx context.Context, filter windowsservice.Filter) ([]domain.Service, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockProvider) Get(ctx context.Context, serviceName string) (*domain.Service, error) {
	if m.getFn != nil {
		return m.getFn(ctx, serviceName)
	}
	return nil, windowsservice.NewServiceError(windowsservice.ErrCodeNotFound, "not found", nil)
}

func (m *mockProvider) Start(ctx context.Context, serviceName string) error {
	if m.startFn != nil {
		return m.startFn(ctx, serviceName)
	}
	return nil
}

func (m *mockProvider) Stop(ctx context.Context, serviceName string) error {
	if m.stopFn != nil {
		return m.stopFn(ctx, serviceName)
	}
	return nil
}

func (m *mockProvider) Restart(ctx context.Context, serviceName string) error {
	if m.restartFn != nil {
		return m.restartFn(ctx, serviceName)
	}
	return nil
}

func (m *mockProvider) SetStartType(ctx context.Context, serviceName string, startType domain.StartType) error {
	if m.setStartTypeFn != nil {
		return m.setStartTypeFn(ctx, serviceName, startType)
	}
	return nil
}

func newTestHandler(m *mockProvider) *Handler {
	tmpDir, err := os.MkdirTemp("", "lsp-api-*")
	if err != nil {
		panic(err)
	}
	settingsStore, err := settings.NewStore(tmpDir)
	if err != nil {
		panic(err)
	}
	return NewHandler(m, nil, testEventSvc, settingsStore)
}

func TestHealthzReturnsOK(t *testing.T) {
	h := newTestHandler(&mockProvider{})
	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %T", resp.Data)
	}

	if data["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", data["status"])
	}

	if data["version"] == "" {
		t.Error("expected non-empty version")
	}
}

func TestHealthzContentType(t *testing.T) {
	h := newTestHandler(&mockProvider{})
	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
}

func TestListServices(t *testing.T) {
	h := newTestHandler(&mockProvider{
		listFn: func(ctx context.Context, filter windowsservice.Filter) ([]domain.Service, error) {
			return []domain.Service{
				{ServiceName: "Spooler", DisplayName: "Print Spooler", Status: domain.ServiceStatusRunning, StartType: domain.StartTypeAutomatic},
				{ServiceName: "WinDefend", DisplayName: "Windows Defender", Status: domain.ServiceStatusRunning, StartType: domain.StartTypeAutomatic},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/windows/services", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	list, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be array, got %T", resp.Data)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 services, got %d", len(list))
	}

	first := list[0].(map[string]interface{})
	if first["serviceName"] != "Spooler" {
		t.Errorf("serviceName = %q, want %q", first["serviceName"], "Spooler")
	}
	if first["status"] != "running" {
		t.Errorf("status = %q, want %q", first["status"], "running")
	}
}

func TestGetService(t *testing.T) {
	h := newTestHandler(&mockProvider{
		getFn: func(ctx context.Context, serviceName string) (*domain.Service, error) {
			if serviceName != "Spooler" {
				t.Errorf("get called with serviceName = %q, want %q", serviceName, "Spooler")
			}
			return &domain.Service{
				ServiceName:  "Spooler",
				DisplayName:  "Print Spooler",
				Status:       domain.ServiceStatusRunning,
				StartType:    domain.StartTypeAutomatic,
				CanStop:      true,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/windows/services/Spooler", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	data := resp.Data.(map[string]interface{})
	if data["serviceName"] != "Spooler" {
		t.Errorf("serviceName = %q, want %q", data["serviceName"], "Spooler")
	}
	if data["status"] != "running" {
		t.Errorf("status = %q, want %q", data["status"], "running")
	}
}

func TestGetServiceNotFound(t *testing.T) {
	h := newTestHandler(&mockProvider{
		getFn: func(ctx context.Context, serviceName string) (*domain.Service, error) {
			return nil, windowsservice.NewServiceError(windowsservice.ErrCodeNotFound, "service not found", nil)
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/windows/services/NonExistent", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestListServicesError(t *testing.T) {
	h := newTestHandler(&mockProvider{
		listFn: func(ctx context.Context, filter windowsservice.Filter) ([]domain.Service, error) {
			return nil, windowsservice.NewServiceError(windowsservice.ErrCodeQueryFailed, "test error", nil)
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/windows/services", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	var errResp APIError
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error.Code != windowsservice.ErrCodeQueryFailed {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, windowsservice.ErrCodeQueryFailed)
	}
}

func TestStartService(t *testing.T) {
	h := newTestHandler(&mockProvider{
		startFn: func(ctx context.Context, serviceName string) error {
			if serviceName != "Spooler" {
				t.Errorf("start called with serviceName = %q, want %q", serviceName, "Spooler")
			}
			return nil
		},
		getFn: func(ctx context.Context, serviceName string) (*domain.Service, error) {
			return &domain.Service{
				ServiceName: "Spooler",
				DisplayName: "Print Spooler",
				Status:      domain.ServiceStatusRunning,
				StartType:   domain.StartTypeAutomatic,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/windows/services/Spooler/start", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)

	data := resp.Data.(map[string]interface{})
	if data["serviceName"] != "Spooler" {
		t.Errorf("serviceName = %q, want %q", data["serviceName"], "Spooler")
	}
	if data["status"] != "running" {
		t.Errorf("status = %q, want %q", data["status"], "running")
	}
}

func TestStartService_NotFound(t *testing.T) {
	h := newTestHandler(&mockProvider{
		startFn: func(ctx context.Context, serviceName string) error {
			return windowsservice.NewServiceError(windowsservice.ErrCodeNotFound, "service not found", nil)
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/windows/services/NonExistent/start", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestStartService_Protected(t *testing.T) {
	h := newTestHandler(&mockProvider{
		startFn: func(ctx context.Context, serviceName string) error {
			// Start is allowed even for protected services, so should succeed.
			return nil
		},
		getFn: func(ctx context.Context, serviceName string) (*domain.Service, error) {
			return &domain.Service{
				ServiceName: "WinDefend",
				Status:      domain.ServiceStatusRunning,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/windows/services/WinDefend/start", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for protected service start, got %d", w.Code)
	}
}

func TestStartService_InternalError(t *testing.T) {
	h := newTestHandler(&mockProvider{
		startFn: func(ctx context.Context, serviceName string) error {
			return windowsservice.NewServiceError(windowsservice.ErrCodeStartFailed, "access denied", nil)
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/windows/services/Spooler/start", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	var errResp APIError
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error.Code != windowsservice.ErrCodeStartFailed {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, windowsservice.ErrCodeStartFailed)
	}
}

func TestStopService(t *testing.T) {
	h := newTestHandler(&mockProvider{
		stopFn: func(ctx context.Context, serviceName string) error {
			if serviceName != "Spooler" {
				t.Errorf("stop called with serviceName = %q, want %q", serviceName, "Spooler")
			}
			return nil
		},
		getFn: func(ctx context.Context, serviceName string) (*domain.Service, error) {
			return &domain.Service{
				ServiceName: "Spooler",
				Status:      domain.ServiceStatusStopped,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/windows/services/Spooler/stop", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	data := resp.Data.(map[string]interface{})
	if data["serviceName"] != "Spooler" {
		t.Errorf("serviceName = %q, want %q", data["serviceName"], "Spooler")
	}
	if data["status"] != "stopped" {
		t.Errorf("status = %q, want %q", data["status"], "stopped")
	}
}

func TestStopService_NotFound(t *testing.T) {
	h := newTestHandler(&mockProvider{
		stopFn: func(ctx context.Context, serviceName string) error {
			return windowsservice.NewServiceError(windowsservice.ErrCodeNotFound, "not found", nil)
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/windows/services/NonExistent/stop", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestStopService_Protected(t *testing.T) {
	h := newTestHandler(&mockProvider{
		stopFn: func(ctx context.Context, serviceName string) error {
			return windowsservice.NewServiceError(windowsservice.ErrCodeProtected, "protected", nil)
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/windows/services/WinDefend/stop", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	var errResp APIError
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error.Code != windowsservice.ErrCodeProtected {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, windowsservice.ErrCodeProtected)
	}
}

func TestRestartService(t *testing.T) {
	h := newTestHandler(&mockProvider{
		restartFn: func(ctx context.Context, serviceName string) error {
			if serviceName != "Spooler" {
				t.Errorf("restart called with serviceName = %q, want %q", serviceName, "Spooler")
			}
			return nil
		},
		getFn: func(ctx context.Context, serviceName string) (*domain.Service, error) {
			return &domain.Service{
				ServiceName: "Spooler",
				Status:      domain.ServiceStatusRunning,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/windows/services/Spooler/restart", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	data := resp.Data.(map[string]interface{})
	if data["serviceName"] != "Spooler" {
		t.Errorf("serviceName = %q, want %q", data["serviceName"], "Spooler")
	}
	if data["status"] != "running" {
		t.Errorf("status = %q, want %q", data["status"], "running")
	}
}

func TestRestartService_NotFound(t *testing.T) {
	h := newTestHandler(&mockProvider{
		restartFn: func(ctx context.Context, serviceName string) error {
			return windowsservice.NewServiceError(windowsservice.ErrCodeNotFound, "not found", nil)
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/windows/services/NonExistent/restart", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRestartService_Protected(t *testing.T) {
	h := newTestHandler(&mockProvider{
		restartFn: func(ctx context.Context, serviceName string) error {
			return windowsservice.NewServiceError(windowsservice.ErrCodeProtected, "protected", nil)
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/windows/services/WinDefend/restart", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	var errResp APIError
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error.Code != windowsservice.ErrCodeProtected {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, windowsservice.ErrCodeProtected)
	}
}

func TestSetStartType(t *testing.T) {
	h := newTestHandler(&mockProvider{
		setStartTypeFn: func(ctx context.Context, serviceName string, startType domain.StartType) error {
			if serviceName != "Spooler" {
				t.Errorf("serviceName = %q, want %q", serviceName, "Spooler")
			}
			if startType != domain.StartTypeManual {
				t.Errorf("startType = %q, want %q", startType, domain.StartTypeManual)
			}
			return nil
		},
	})

	body := `{"startType":"manual"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/windows/services/Spooler/start-type", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	data := resp.Data.(map[string]interface{})
	if data["serviceName"] != "Spooler" {
		t.Errorf("serviceName = %q, want %q", data["serviceName"], "Spooler")
	}
	if data["startType"] != "manual" {
		t.Errorf("startType = %q, want %q", data["startType"], "manual")
	}
}

func TestSetStartType_InvalidBody(t *testing.T) {
	h := newTestHandler(&mockProvider{})
	req := httptest.NewRequest(http.MethodPatch, "/api/windows/services/Spooler/start-type", http.NoBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSetStartType_InvalidValue(t *testing.T) {
	h := newTestHandler(&mockProvider{})
	body := `{"startType":"invalid_value"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/windows/services/Spooler/start-type", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var errResp APIError
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error.Code != windowsservice.ErrCodeInvalidStartType {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, windowsservice.ErrCodeInvalidStartType)
	}
}

func TestSetStartType_NotFound(t *testing.T) {
	h := newTestHandler(&mockProvider{
		setStartTypeFn: func(ctx context.Context, serviceName string, startType domain.StartType) error {
			return windowsservice.NewServiceError(windowsservice.ErrCodeNotFound, "not found", nil)
		},
	})

	body := `{"startType":"manual"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/windows/services/NonExistent/start-type", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestSetStartType_Protected(t *testing.T) {
	h := newTestHandler(&mockProvider{
		setStartTypeFn: func(ctx context.Context, serviceName string, startType domain.StartType) error {
			return windowsservice.NewServiceError(windowsservice.ErrCodeProtected, "protected", nil)
		},
	})

	body := `{"startType":"disabled"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/windows/services/WinDefend/start-type", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	var errResp APIError
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error.Code != windowsservice.ErrCodeProtected {
		t.Errorf("error code = %q, want %q", errResp.Error.Code, windowsservice.ErrCodeProtected)
	}
}

func TestRouterNotFound(t *testing.T) {
	h := newTestHandler(&mockProvider{})
	req := httptest.NewRequest(http.MethodGet, "/api/unknown", nil)
	w := httptest.NewRecorder()

	h.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
