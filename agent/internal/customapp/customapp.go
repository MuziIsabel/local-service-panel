// Package customapp handles Custom App management: CRUD, process lifecycle, logging.
package customapp

import (
	"fmt"

	"github.com/user/local-service-panel/agent/internal/domain"
)

// Error codes for Custom App operations.
const (
	ErrCodeNotFound            = "CUSTOM_APP_NOT_FOUND"
	ErrCodeCreateFailed        = "CUSTOM_APP_CREATE_FAILED"
	ErrCodeUpdateFailed        = "CUSTOM_APP_UPDATE_FAILED"
	ErrCodeDeleteFailed        = "CUSTOM_APP_DELETE_FAILED"
	ErrCodeInvalidExecutable   = "CUSTOM_APP_INVALID_EXECUTABLE"
	ErrCodeInvalidWorkingDir   = "CUSTOM_APP_INVALID_WORKING_DIR"
	ErrCodeAlreadyRunning      = "CUSTOM_APP_ALREADY_RUNNING"
	ErrCodeNotRunning          = "CUSTOM_APP_NOT_RUNNING"
	ErrCodeStartFailed         = "CUSTOM_APP_START_FAILED"
	ErrCodeStopFailed          = "CUSTOM_APP_STOP_FAILED"
	ErrCodeLogReadFailed       = "CUSTOM_APP_LOG_READ_FAILED"
	ErrCodeDeleteRunningDenied = "CUSTOM_APP_RUNNING_DELETE_DENIED"
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

// --- DTOs ---

// DTO is the API-facing representation of a Custom App.
type DTO struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Status         string   `json:"status"`
	ExecutablePath string   `json:"executablePath"`
	WorkingDir     string   `json:"workingDir,omitempty"`
	Args           []string `json:"args,omitempty"`
	AutoStart      bool     `json:"autoStart"`
	PID            int      `json:"pid,omitempty"`
	LastStartedAt  string   `json:"lastStartedAt,omitempty"`
	LastStoppedAt  string   `json:"lastStoppedAt,omitempty"`
	LastError      string   `json:"lastError,omitempty"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
}

// CreateRequest is the request body for POST /api/custom-apps.
type CreateRequest struct {
	Name           string   `json:"name"`
	ExecutablePath string   `json:"executablePath"`
	WorkingDir     string   `json:"workingDir,omitempty"`
	Args           []string `json:"args,omitempty"`
	AutoStart      bool     `json:"autoStart,omitempty"`
}

// UpdateRequest is the request body for PATCH /api/custom-apps/{id}.
type UpdateRequest struct {
	Name           *string  `json:"name,omitempty"`
	ExecutablePath *string  `json:"executablePath,omitempty"`
	WorkingDir     *string  `json:"workingDir,omitempty"`
	Args           []string `json:"args,omitempty"`
	AutoStart      *bool    `json:"autoStart,omitempty"`
}

// --- Domain conversions ---

// ToDTO converts a domain.CustomApp to an API DTO.
func ToDTO(app *domain.CustomApp) DTO {
	return DTO{
		ID:             app.ID,
		Name:           app.Name,
		Type:           "custom_app",
		Status:         string(app.Status),
		ExecutablePath: app.ExecutablePath,
		WorkingDir:     app.WorkingDir,
		Args:           app.Args,
		AutoStart:      app.AutoStart,
		PID:            app.PID,
		LastStartedAt:  app.LastStartedAt,
		LastStoppedAt:  app.LastStoppedAt,
		LastError:      app.LastError,
		CreatedAt:      app.CreatedAt,
		UpdatedAt:      app.UpdatedAt,
	}
}

// ToDTOList converts a slice of domain.CustomApp to a slice of DTO.
func ToDTOList(apps []*domain.CustomApp) []DTO {
	dtos := make([]DTO, len(apps))
	for i, app := range apps {
		dtos[i] = ToDTO(app)
	}
	return dtos
}

// --- Validation ---

// ValidateCreate checks that a CreateRequest is valid.
func ValidateCreate(req *CreateRequest) error {
	if req.Name == "" {
		return NewServiceError("VALIDATION_ERROR", "name is required", nil)
	}
	if req.ExecutablePath == "" {
		return NewServiceError("VALIDATION_ERROR", "executablePath is required", nil)
	}
	return nil
}
