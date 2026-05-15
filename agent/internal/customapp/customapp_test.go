package customapp

import (
	"testing"

	"github.com/user/local-service-panel/agent/internal/domain"
)

func TestToDTO(t *testing.T) {
	app := &domain.CustomApp{
		ID:             "uuid-123",
		Name:           "Test App",
		ExecutablePath: "C:\\apps\\test.exe",
		WorkingDir:     "C:\\apps",
		Args:           []string{"--port", "8080"},
		Status:         domain.RunStatusStopped,
		AutoStart:      true,
		CreatedAt:      "2026-01-01T00:00:00Z",
	}

	dto := ToDTO(app)
	if dto.ID != "uuid-123" {
		t.Errorf("ID = %q, want %q", dto.ID, "uuid-123")
	}
	if dto.Name != "Test App" {
		t.Errorf("Name = %q, want %q", dto.Name, "Test App")
	}
	if dto.Type != "custom_app" {
		t.Errorf("Type = %q, want %q", dto.Type, "custom_app")
	}
	if dto.Status != "stopped" {
		t.Errorf("Status = %q, want %q", dto.Status, "stopped")
	}
	if dto.Args[0] != "--port" || dto.Args[1] != "8080" {
		t.Errorf("Args = %v", dto.Args)
	}
	if !dto.AutoStart {
		t.Error("AutoStart should be true")
	}
}

func TestToDTOList(t *testing.T) {
	apps := []*domain.CustomApp{
		{ID: "1", Name: "App1", ExecutablePath: "C:\\a.exe", Status: domain.RunStatusRunning},
		{ID: "2", Name: "App2", ExecutablePath: "C:\\b.exe", Status: domain.RunStatusStopped},
	}

	dtos := ToDTOList(apps)
	if len(dtos) != 2 {
		t.Fatalf("len = %d, want 2", len(dtos))
	}
	if dtos[0].Status != "running" {
		t.Errorf("Status[0] = %q, want %q", dtos[0].Status, "running")
	}
	if dtos[1].Status != "stopped" {
		t.Errorf("Status[1] = %q, want %q", dtos[1].Status, "stopped")
	}
}

func TestValidateCreate(t *testing.T) {
	// Valid
	err := ValidateCreate(&CreateRequest{
		Name:           "My App",
		ExecutablePath: "C:\\app.exe",
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Missing name
	err = ValidateCreate(&CreateRequest{
		ExecutablePath: "C:\\app.exe",
	})
	if err == nil {
		t.Error("expected error for missing name")
	}

	// Missing executablePath
	err = ValidateCreate(&CreateRequest{
		Name: "My App",
	})
	if err == nil {
		t.Error("expected error for missing executablePath")
	}
}

func TestServiceError(t *testing.T) {
	err := NewServiceError(ErrCodeNotFound, "app not found", nil)
	if err.Code != ErrCodeNotFound {
		t.Errorf("Code = %q, want %q", err.Code, ErrCodeNotFound)
	}
	if err.Error() == "" {
		t.Error("Error() should not be empty")
	}

	wrappedErr := NewServiceError(ErrCodeStartFailed, "start failed", nil)
	if wrappedErr.Code != ErrCodeStartFailed {
		t.Errorf("Code = %q", wrappedErr.Code)
	}
}
