// Package domain holds the core domain types for the Agent.
package domain

// CustomApp represents a user-managed application/script/command.
type CustomApp struct {
	ID             string
	Name           string
	ExecutablePath string
	WorkingDir     string
	Args           []string
	StopCommand    string
	AutoStart      bool
	PID            int
	Status         RunStatus
	LastStartedAt  string
	LastStoppedAt  string
	LastError      string
	CreatedAt      string
	UpdatedAt      string
}

// RunStatus represents the runtime status of a Custom App.
type RunStatus string

const (
	RunStatusRunning RunStatus = "running"
	RunStatusStopped RunStatus = "stopped"
	RunStatusError   RunStatus = "error"
	RunStatusUnknown RunStatus = "unknown"
)
