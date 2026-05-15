package windowsservice

import (
	"testing"

	"github.com/user/local-service-panel/agent/internal/domain"
)

func TestParseStartType(t *testing.T) {
	tests := []struct {
		input string
		want  domain.StartType
		err   bool
	}{
		{"automatic", domain.StartTypeAutomatic, false},
		{"automatic_delayed", domain.StartTypeAutomaticDelayed, false},
		{"manual", domain.StartTypeManual, false},
		{"disabled", domain.StartTypeDisabled, false},
		{"unknown", domain.StartTypeUnknown, true},
		{"invalid", domain.StartTypeUnknown, true},
	}
	for _, tt := range tests {
		got, err := ParseStartType(tt.input)
		if (err != nil) != tt.err {
			t.Errorf("ParseStartType(%q) error = %v, want err=%v", tt.input, err, tt.err)
		}
		if err == nil && got != tt.want {
			t.Errorf("ParseStartType(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestToDTO(t *testing.T) {
	svc := domain.Service{
		ServiceName: "Spooler",
		DisplayName: "Print Spooler",
		Status:      domain.ServiceStatusRunning,
		StartType:   domain.StartTypeAutomatic,
		CanStop:     true,
	}

	dto := ToDTO(svc, false)
	if dto.ID != "windows_service:Spooler" {
		t.Errorf("ID = %q, want %q", dto.ID, "windows_service:Spooler")
	}
	if dto.ServiceName != "Spooler" {
		t.Errorf("ServiceName = %q, want %q", dto.ServiceName, "Spooler")
	}
	if dto.DisplayName != "Print Spooler" {
		t.Errorf("DisplayName = %q, want %q", dto.DisplayName, "Print Spooler")
	}
	if dto.Status != "running" {
		t.Errorf("Status = %q, want %q", dto.Status, "running")
	}
	if dto.StartType != "automatic" {
		t.Errorf("StartType = %q, want %q", dto.StartType, "automatic")
	}
	if dto.Protected {
		t.Error("Expected protected = false")
	}
}

func TestToDTOProtected(t *testing.T) {
	svc := domain.Service{
		ServiceName: "WinDefend",
		DisplayName: "Windows Defender",
		Status:      domain.ServiceStatusRunning,
		StartType:   domain.StartTypeAutomatic,
	}

	dto := ToDTO(svc, true)
	if !dto.Protected {
		t.Error("Expected protected = true")
	}
}

func TestToDTOList(t *testing.T) {
	services := []domain.Service{
		{ServiceName: "Spooler", DisplayName: "Print Spooler", Status: domain.ServiceStatusRunning},
		{ServiceName: "WinDefend", DisplayName: "Windows Defender", Status: domain.ServiceStatusRunning},
	}

	isProtected := func(name domain.ServiceName) bool {
		return name == "WinDefend"
	}

	dtos := ToDTOList(services, isProtected)
	if len(dtos) != 2 {
		t.Fatalf("len = %d, want 2", len(dtos))
	}

	if dtos[0].Protected {
		t.Error("Spooler should not be protected")
	}
	if !dtos[1].Protected {
		t.Error("WinDefend should be protected")
	}
}

func TestIsProtected(t *testing.T) {
	if !IsProtected("WinDefend") {
		t.Error("WinDefend should be protected")
	}
	if !IsProtected("EventLog") {
		t.Error("EventLog should be protected")
	}
	if !IsProtected("RpcSs") {
		t.Error("RpcSs should be protected")
	}
	if IsProtected("Spooler") {
		t.Error("Spooler should not be protected")
	}
	if IsProtected("NonExistent") {
		t.Error("NonExistent service should not be protected")
	}
}

func TestIsHighRiskAction(t *testing.T) {
	// Protected service + high-risk action
	if !IsHighRiskAction("WinDefend", "stop") {
		t.Error("stop on WinDefend should be high risk")
	}
	if !IsHighRiskAction("WinDefend", "restart") {
		t.Error("restart on WinDefend should be high risk")
	}
	if !IsHighRiskAction("WinDefend", "set_start_type") {
		t.Error("set_start_type on WinDefend should be high risk")
	}

	// Start action should be low risk
	if IsHighRiskAction("WinDefend", "start") {
		t.Error("start on WinDefend should not be high risk")
	}

	// Non-protected service should not be high risk
	if IsHighRiskAction("Spooler", "stop") {
		t.Error("stop on Spooler should not be high risk")
	}
	if IsHighRiskAction("Spooler", "restart") {
		t.Error("restart on Spooler should not be high risk")
	}
}

func TestServiceError(t *testing.T) {
	err := NewServiceError(ErrCodeNotFound, "service not found", nil)
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
