package domain

import "testing"

func TestServiceStatusConstants(t *testing.T) {
	tests := []struct {
		status ServiceStatus
		want   string
	}{
		{ServiceStatusRunning, "running"},
		{ServiceStatusStopped, "stopped"},
		{ServiceStatusStartPending, "start_pending"},
		{ServiceStatusStopPending, "stop_pending"},
		{ServiceStatusPausePending, "pause_pending"},
		{ServiceStatusPaused, "paused"},
		{ServiceStatusContinuePending, "continue_pending"},
		{ServiceStatusUnknown, "unknown"},
	}
	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("ServiceStatus(%s) = %q, want %q", string(tt.status), tt.status, tt.want)
		}
	}
}

func TestStartTypeConstants(t *testing.T) {
	tests := []struct {
		st   StartType
		want string
	}{
		{StartTypeAutomatic, "automatic"},
		{StartTypeAutomaticDelayed, "automatic_delayed"},
		{StartTypeManual, "manual"},
		{StartTypeDisabled, "disabled"},
		{StartTypeUnknown, "unknown"},
	}
	for _, tt := range tests {
		if string(tt.st) != tt.want {
			t.Errorf("StartType(%s) = %q, want %q", string(tt.st), tt.st, tt.want)
		}
	}
}
