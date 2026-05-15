package customapp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/user/local-service-panel/agent/internal/domain"
)

// ProcessManager handles starting, stopping, and managing custom app processes.
type ProcessManager struct {
	logsDir string
}

// NewProcessManager creates a new process manager.
// logsDir is the base directory for per-app log files (e.g. .data/logs/apps).
func NewProcessManager(logsDir string) *ProcessManager {
	return &ProcessManager{logsDir: logsDir}
}

// StartResult holds the result of starting a process.
type StartResult struct {
	PID int
}

// Start launches the given custom app and returns the PID.
// It validates the executable path and working directory before starting.
func (pm *ProcessManager) Start(app *domain.CustomApp) (*StartResult, error) {
	// Validate executable path - check if accessible
	if _, err := exec.LookPath(app.ExecutablePath); err != nil {
		return nil, NewServiceError(ErrCodeInvalidExecutable,
			fmt.Sprintf("executable not found: %s", app.ExecutablePath), nil)
	}

	// Validate working directory if set
	if app.WorkingDir != "" {
		if info, err := os.Stat(app.WorkingDir); err != nil || !info.IsDir() {
			return nil, NewServiceError(ErrCodeInvalidWorkingDir,
				fmt.Sprintf("working directory not found: %s", app.WorkingDir), nil)
		}
	}

	// Create log directory
	appLogDir := filepath.Join(pm.logsDir, app.ID)
	if err := os.MkdirAll(appLogDir, 0755); err != nil {
		return nil, NewServiceError(ErrCodeStartFailed,
			fmt.Sprintf("failed to create log directory: %s", appLogDir), err)
	}

	// Open log files
	stdoutPath := filepath.Join(appLogDir, "stdout.log")
	stderrPath := filepath.Join(appLogDir, "stderr.log")

	stdoutFile, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, NewServiceError(ErrCodeStartFailed, "failed to open stdout log", err)
	}

	stderrFile, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		stdoutFile.Close()
		return nil, NewServiceError(ErrCodeStartFailed, "failed to open stderr log", err)
	}

	// Build command - use executablePath + args array (no shell)
	cmd := exec.Command(app.ExecutablePath, app.Args...)
	cmd.Dir = app.WorkingDir
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	// Start the process
	if err := cmd.Start(); err != nil {
		stdoutFile.Close()
		stderrFile.Close()
		return nil, NewServiceError(ErrCodeStartFailed,
			fmt.Sprintf("failed to start %s", app.ExecutablePath), err)
	}

	// Store PID
	result := &StartResult{PID: cmd.Process.Pid}

	// Wait in background to handle process exit (cleanup log files later)
	go func() {
		cmd.Wait()
		stdoutFile.Close()
		stderrFile.Close()
	}()

	return result, nil
}

// Stop kills a process by PID.
func (pm *ProcessManager) Stop(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		// Process already exited - treat as success
		return nil
	}

	if err := proc.Kill(); err != nil {
		// May already be terminated
		return nil
	}

	// Wait for process to exit (non-blocking if already gone)
	proc.Wait()
	return nil
}
