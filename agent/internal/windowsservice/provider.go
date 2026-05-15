// Package windowsservice defines the Provider interface and DTOs
// for Windows Service management. Platform-specific implementations
// live in separate files.
package windowsservice

import (
	"context"
	"fmt"

	"github.com/user/local-service-panel/agent/internal/domain"
)

// Error codes for Windows Service operations.
// These should be returned via API error responses.
const (
	ErrCodeNotFound          = "WINDOWS_SERVICE_NOT_FOUND"
	ErrCodeQueryFailed       = "WINDOWS_SERVICE_QUERY_FAILED"
	ErrCodeStartFailed       = "WINDOWS_SERVICE_START_FAILED"
	ErrCodeStopFailed        = "WINDOWS_SERVICE_STOP_FAILED"
	ErrCodeRestartFailed     = "WINDOWS_SERVICE_RESTART_FAILED"
	ErrCodeStartTypeFailed   = "WINDOWS_SERVICE_START_TYPE_FAILED"
	ErrCodeProtected         = "WINDOWS_SERVICE_PROTECTED"
	ErrCodePermissionDenied  = "WINDOWS_SERVICE_PERMISSION_DENIED"
	ErrCodeUnsupported       = "WINDOWS_SERVICE_UNSUPPORTED_PLATFORM"
	ErrCodeInvalidStartType  = "INVALID_START_TYPE"
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

func (e *ServiceError) Unwrap() error {
	return e.Err
}

// NewServiceError creates a ServiceError.
func NewServiceError(code, message string, err error) *ServiceError {
	return &ServiceError{Code: code, Message: message, Err: err}
}

// Filter specifies criteria for listing Windows Services.
type Filter struct {
	Keyword           string
	Status            domain.ServiceStatus // empty means no filter
	StartType         domain.StartType     // empty means no filter
	IncludeProtected  bool                 // true to include protected services
}

// Provider defines the interface for Windows Service operations.
// Platform-specific implementations (e.g. Windows SCM) should implement this.
type Provider interface {
	// List returns all Windows Services matching the given filter.
	List(ctx context.Context, filter Filter) ([]domain.Service, error)

	// Get returns a single Windows Service by name.
	Get(ctx context.Context, serviceName string) (*domain.Service, error)

	// Start attempts to start the named service.
	Start(ctx context.Context, serviceName string) error

	// Stop attempts to stop the named service.
	Stop(ctx context.Context, serviceName string) error

	// Restart stops then starts the named service.
	Restart(ctx context.Context, serviceName string) error

	// SetStartType modifies the startup type of the named service.
	SetStartType(ctx context.Context, serviceName string, startType domain.StartType) error
}

// DTO is the API-facing representation of a Windows Service.
// It is used for JSON serialization in HTTP responses.
type DTO struct {
	ID                string `json:"id"`
	ServiceName       string `json:"serviceName"`
	DisplayName       string `json:"displayName"`
	Description       string `json:"description,omitempty"`
	Status            string `json:"status"`
	StartType         string `json:"startType"`
	ExecutablePath    string `json:"executablePath,omitempty"`
	CanStop           bool   `json:"canStop"`
	CanPauseAndContinue bool `json:"canPauseAndContinue"`
	Protected         bool   `json:"protected"`
}

// ToDTO converts a domain.Service to an API DTO, marking protected status.
func ToDTO(svc domain.Service, protected bool) DTO {
	return DTO{
		ID:                  "windows_service:" + string(svc.ServiceName),
		ServiceName:         string(svc.ServiceName),
		DisplayName:         svc.DisplayName,
		Description:         svc.Description,
		Status:              string(svc.Status),
		StartType:           string(svc.StartType),
		ExecutablePath:      svc.ExecutablePath,
		CanStop:             svc.CanStop,
		CanPauseAndContinue: svc.CanPauseAndContinue,
		Protected:           protected,
	}
}

// ToDTOList converts a slice of domain.Service to a slice of DTO.
func ToDTOList(services []domain.Service, isProtected func(domain.ServiceName) bool) []DTO {
	dtos := make([]DTO, len(services))
	for i, svc := range services {
		dtos[i] = ToDTO(svc, isProtected(svc.ServiceName))
	}
	return dtos
}

// ParseStartType converts a string to StartType, returning an error if invalid.
func ParseStartType(s string) (domain.StartType, error) {
	switch s {
	case "automatic":
		return domain.StartTypeAutomatic, nil
	case "automatic_delayed":
		return domain.StartTypeAutomaticDelayed, nil
	case "manual":
		return domain.StartTypeManual, nil
	case "disabled":
		return domain.StartTypeDisabled, nil
	default:
		return domain.StartTypeUnknown, NewServiceError(ErrCodeInvalidStartType,
			fmt.Sprintf("invalid start type: %s", s), nil)
	}
}
