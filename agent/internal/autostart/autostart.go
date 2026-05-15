// Package autostart manages startup entries for Custom Apps.
// Windows: uses HKCU\Software\Microsoft\Windows\CurrentVersion\Run.
// Non-Windows: returns ErrCodeUnsupported.
package autostart

import (
	"fmt"
	"strings"
)

// Error codes.
const (
	ErrCodeUnsupported          = "AUTOSTART_UNSUPPORTED_PLATFORM"
	ErrCodeRegistryOpenFailed   = "AUTOSTART_REGISTRY_OPEN_FAILED"
	ErrCodeRegistryWriteFailed  = "AUTOSTART_REGISTRY_WRITE_FAILED"
	ErrCodeRegistryDeleteFailed = "AUTOSTART_REGISTRY_DELETE_FAILED"
	ErrCodeInvalidTarget        = "AUTOSTART_INVALID_TARGET"
	ErrCodeExecutableNotFound   = "AUTOSTART_EXECUTABLE_NOT_FOUND"
	ErrCodeCommandBuildFailed   = "AUTOSTART_COMMAND_BUILD_FAILED"
)

// ServiceError wraps an operation error with a well-known error code.
type ServiceError struct {
	Code    string
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *ServiceError) Unwrap() error { return e.Err }

// NewServiceError creates a ServiceError.
func NewServiceError(code, message string, err error) *ServiceError {
	return &ServiceError{Code: code, Message: message, Err: err}
}

// Provider defines the interface for managing startup entries.
type Provider interface {
	// Enable creates or updates the startup entry for the given target.
	Enable(entry Entry) error
	// Disable removes the startup entry for the given target.
	Disable(entry Entry) error
	// IsEnabled checks whether a startup entry exists.
	IsEnabled(entry Entry) (bool, error)
}

// Entry describes an autostart target.
type Entry struct {
	// ID is the unique identifier (e.g. Custom App UUID).
	ID string
	// Name is the display name.
	Name string
	// ExecutablePath is the full path to the executable.
	ExecutablePath string
	// Args are the command-line arguments.
	Args []string
	// WorkingDir is the working directory (stored in the entry for reference, but HKCU Run doesn't support it directly).
	WorkingDir string
}

// RegistryKeyName returns the stable name used in the Windows Registry Run key.
func (e Entry) RegistryKeyName() string {
	return fmt.Sprintf("LocalServicePanel_CustomApp_%s", e.ID)
}

// BuildCommand constructs the command string for a Run entry.
// Rules:
// - executablePath is wrapped in double quotes.
// - Each argument is space-separated.
// - Arguments containing spaces are wrapped in double quotes.
func BuildCommand(entry Entry) (string, error) {
	if entry.ExecutablePath == "" {
		return "", NewServiceError(ErrCodeCommandBuildFailed, "executable path is empty", nil)
	}

	var parts []string
	// Quote the executable path.
	parts = append(parts, fmt.Sprintf(`"%s"`, entry.ExecutablePath))

	for _, arg := range entry.Args {
		if strings.Contains(arg, " ") && !strings.HasPrefix(arg, `"`) {
			parts = append(parts, fmt.Sprintf(`"%s"`, arg))
		} else {
			parts = append(parts, arg)
		}
	}

	return strings.Join(parts, " "), nil
}

// ParseCommand splits a command string back into executable path and args.
// This is the reverse of BuildCommand, used when reading existing Run entries.
func ParseCommand(cmd string) (executablePath string, args []string) {
	parts := splitCommand(cmd)
	if len(parts) == 0 {
		return "", nil
	}
	// Remove surrounding quotes from executable path.
	executablePath = strings.Trim(parts[0], `"`)
	for _, p := range parts[1:] {
		args = append(args, p)
	}
	return
}

// splitCommand splits a command string by spaces, respecting quoted sections.
func splitCommand(cmd string) []string {
	var result []string
	var current strings.Builder
	inQuote := false

	for _, c := range cmd {
		switch {
		case c == '"':
			inQuote = !inQuote
			current.WriteRune(c)
		case c == ' ' && !inQuote:
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(c)
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}
