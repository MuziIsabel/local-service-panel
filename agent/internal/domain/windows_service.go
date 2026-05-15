// Package domain holds the core domain types for the Agent.
package domain

// ServiceName is the internal Windows Service name (e.g. "Spooler").
type ServiceName string

// ServiceStatus represents the current state of a Windows Service.
type ServiceStatus string

const (
	ServiceStatusRunning         ServiceStatus = "running"
	ServiceStatusStopped         ServiceStatus = "stopped"
	ServiceStatusStartPending    ServiceStatus = "start_pending"
	ServiceStatusStopPending     ServiceStatus = "stop_pending"
	ServiceStatusPausePending    ServiceStatus = "pause_pending"
	ServiceStatusPaused          ServiceStatus = "paused"
	ServiceStatusContinuePending ServiceStatus = "continue_pending"
	ServiceStatusUnknown         ServiceStatus = "unknown"
)

// StartType represents how a Windows Service starts.
type StartType string

const (
	StartTypeAutomatic        StartType = "automatic"
	StartTypeAutomaticDelayed StartType = "automatic_delayed"
	StartTypeManual           StartType = "manual"
	StartTypeDisabled         StartType = "disabled"
	StartTypeUnknown          StartType = "unknown"
)

// Service is the internal representation of a Windows Service.
type Service struct {
	ServiceName       ServiceName
	DisplayName       string
	Description       string
	Status            ServiceStatus
	StartType         StartType
	ExecutablePath    string
	PID               uint32
	CanStop           bool
	CanPauseAndContinue bool
}
