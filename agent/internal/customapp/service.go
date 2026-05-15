package customapp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/user/local-service-panel/agent/internal/domain"
	"github.com/user/local-service-panel/agent/internal/db/repository"
	"github.com/user/local-service-panel/agent/internal/autostart"
)

// Service provides Custom App business logic, bridging DB repository and domain models.
type Service struct {
	repo    *repository.ManagedTargetRepo
	procMgr *ProcessManager
	auto    autostart.Provider
}

// NewService creates a new CustomApp service.
func NewService(repo *repository.ManagedTargetRepo, dataDir string, auto autostart.Provider) *Service {
	logsDir := dataDir + "/logs/apps"
	return &Service{
		repo:    repo,
		procMgr: NewProcessManager(logsDir),
		auto:    auto,
	}
}

// List returns all custom apps, optionally filtered by keyword.
func (s *Service) List(keyword string) ([]*domain.CustomApp, error) {
	var records []*repository.ManagedTarget
	var err error

	if keyword != "" {
		records, err = s.repo.ListByKeyword(keyword, "custom_app")
	} else {
		records, err = s.repo.List("custom_app")
	}
	if err != nil {
		return nil, NewServiceError(ErrCodeNotFound, "failed to list custom apps", err)
	}

	apps := make([]*domain.CustomApp, 0, len(records))
	for _, r := range records {
		app, err := repoToDomain(r)
		if err != nil {
			return nil, NewServiceError(ErrCodeCreateFailed, "failed to parse app record", err)
		}
		apps = append(apps, app)
	}
	return apps, nil
}

// GetByID returns a single custom app by ID.
func (s *Service) GetByID(id string) (*domain.CustomApp, error) {
	if id == "" {
		return nil, NewServiceError(ErrCodeNotFound, "id is required", nil)
	}
	record, err := s.repo.GetByID(id)
	if err != nil {
		return nil, NewServiceError(ErrCodeNotFound, fmt.Sprintf("custom app %q not found", id), err)
	}
	return repoToDomain(record)
}

// Create adds a new custom app from a create request.
func (s *Service) Create(req *CreateRequest) (*domain.CustomApp, error) {
	if err := ValidateCreate(req); err != nil {
		return nil, err
	}

	app := &domain.CustomApp{
		Name:           req.Name,
		ExecutablePath: req.ExecutablePath,
		WorkingDir:     req.WorkingDir,
		Args:           req.Args,
		AutoStart:      req.AutoStart,
		Status:         domain.RunStatusStopped,
	}

	repoRecord, err := domainToRepo(app)
	if err != nil {
		return nil, NewServiceError(ErrCodeCreateFailed, "failed to prepare record", err)
	}

	if err := s.repo.Create(repoRecord); err != nil {
		return nil, NewServiceError(ErrCodeCreateFailed, "failed to create custom app", err)
	}

	return repoToDomain(repoRecord)
}

// Update applies partial updates to an existing custom app.
func (s *Service) Update(id string, req *UpdateRequest) (*domain.CustomApp, error) {
	existing, err := s.GetByID(id)
	if err != nil {
		return nil, err // already wrapped as ErrCodeNotFound
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.ExecutablePath != nil {
		existing.ExecutablePath = *req.ExecutablePath
	}
	if req.WorkingDir != nil {
		existing.WorkingDir = *req.WorkingDir
	}
	if req.Args != nil {
		existing.Args = req.Args
	}
	if req.AutoStart != nil {
		existing.AutoStart = *req.AutoStart
	}

	repoRecord, err := domainToRepo(existing)
	if err != nil {
		return nil, NewServiceError(ErrCodeUpdateFailed, "failed to prepare record", err)
	}

	if err := s.repo.Update(repoRecord); err != nil {
		return nil, NewServiceError(ErrCodeUpdateFailed, "failed to update custom app", err)
	}

	return repoToDomain(repoRecord)
}

// Delete removes a custom app by ID.
// Returns an error if the app is currently running.
func (s *Service) Delete(id string) error {
	app, err := s.GetByID(id)
	if err != nil {
		return err
	}

	if app.Status == domain.RunStatusRunning {
		return NewServiceError(ErrCodeDeleteRunningDenied,
			fmt.Sprintf("custom app %q is currently running, stop it first", id), nil)
	}

	if err := s.repo.Delete(id); err != nil {
		return NewServiceError(ErrCodeDeleteFailed, "failed to delete custom app", err)
	}
	return nil
}

// Start launches the given custom app and records the PID.
func (s *Service) Start(id string) (*domain.CustomApp, error) {
	app, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if app.Status == domain.RunStatusRunning {
		return nil, NewServiceError(ErrCodeAlreadyRunning,
			fmt.Sprintf("custom app %q is already running", id), nil)
	}

	result, err := s.procMgr.Start(app)
	if err != nil {
		now := time.Now().UTC().Format(time.RFC3339)
		app.LastError = err.Error()
		app.Status = domain.RunStatusError
		app.LastStoppedAt = now

		repoRecord, _ := domainToRepo(app)
		if repoRecord != nil {
			s.repo.Update(repoRecord)
		}
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	app.PID = result.PID
	app.Status = domain.RunStatusRunning
	app.LastStartedAt = now
	app.LastError = ""

	repoRecord, err := domainToRepo(app)
	if err != nil {
		return nil, NewServiceError(ErrCodeStartFailed, "failed to marshal record", err)
	}

	if err := s.repo.Update(repoRecord); err != nil {
		return nil, NewServiceError(ErrCodeStartFailed, "failed to update record after start", err)
	}

	return app, nil
}

// Stop stops a running custom app by ID.
func (s *Service) Stop(id string) (*domain.CustomApp, error) {
	app, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if app.Status != domain.RunStatusRunning || app.PID == 0 {
		return nil, NewServiceError(ErrCodeNotRunning,
			fmt.Sprintf("custom app %q is not running", id), nil)
	}

	if err := s.procMgr.Stop(app.PID); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	app.PID = 0
	app.Status = domain.RunStatusStopped
	app.LastStoppedAt = now

	repoRecord, err := domainToRepo(app)
	if err != nil {
		return nil, NewServiceError(ErrCodeStopFailed, "failed to marshal record", err)
	}

	if err := s.repo.Update(repoRecord); err != nil {
		return nil, NewServiceError(ErrCodeStopFailed, "failed to update record after stop", err)
	}

	return app, nil
}

// GetLogs reads the last N lines of stdout and stderr for a custom app.
func (s *Service) GetLogs(id string, lines int) (*LogsResponse, error) {
	// Verify the app exists
	if _, err := s.GetByID(id); err != nil {
		return nil, err
	}

	logs, err := s.procMgr.ReadLogs(id, lines)
	if err != nil {
		return nil, NewServiceError(ErrCodeLogReadFailed, "failed to read logs", err)
	}
	return logs, nil
}

// SetAutoStart enables or disables autostart for a custom app.
func (s *Service) SetAutoStart(id string, enabled bool) (*domain.CustomApp, error) {
	app, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	entry := autostart.Entry{
		ID:             app.ID,
		Name:           app.Name,
		ExecutablePath: app.ExecutablePath,
		Args:           app.Args,
		WorkingDir:     app.WorkingDir,
	}

	if enabled {
		if err := s.auto.Enable(entry); err != nil {
			return nil, err
		}
	} else {
		if err := s.auto.Disable(entry); err != nil {
			return nil, err
		}
	}

	app.AutoStart = enabled
	repoRecord, err := domainToRepo(app)
	if err != nil {
		return nil, NewServiceError(ErrCodeUpdateFailed, "failed to marshal record", err)
	}

	if err := s.repo.Update(repoRecord); err != nil {
		return nil, NewServiceError(ErrCodeUpdateFailed, "failed to update auto_start", err)
	}

	return app, nil
}

// repoToDomain converts a repository ManagedTarget to a domain CustomApp.
func repoToDomain(r *repository.ManagedTarget) (*domain.CustomApp, error) {
	var args []string
	if r.ArgsJSON != "" {
		if err := json.Unmarshal([]byte(r.ArgsJSON), &args); err != nil {
			// Non-fatal: treat unparseable args as empty
			args = nil
		}
	}

	status := domain.RunStatusUnknown
	if r.PID.Valid && r.PID.Int64 > 0 {
		status = domain.RunStatusRunning
	} else if r.LastError.Valid && r.LastError.String != "" {
		status = domain.RunStatusError
	} else {
		status = domain.RunStatusStopped
	}

	app := &domain.CustomApp{
		ID:             r.ID,
		Name:           r.Name,
		ExecutablePath: r.ExecutablePath,
		WorkingDir:     r.WorkingDir,
		Args:           args,
		StopCommand:    r.StopCommand,
		AutoStart:      r.AutoStart,
		Status:         status,
		LastStartedAt:  r.LastStartedAt.String,
		LastStoppedAt:  r.LastStoppedAt.String,
		LastError:      r.LastError.String,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
	if r.PID.Valid {
		app.PID = int(r.PID.Int64)
	}
	return app, nil
}

// domainToRepo converts a domain CustomApp to a repository ManagedTarget.
func domainToRepo(app *domain.CustomApp) (*repository.ManagedTarget, error) {
	argsJSON := ""
	if len(app.Args) > 0 {
		data, err := json.Marshal(app.Args)
		if err != nil {
			return nil, NewServiceError(ErrCodeCreateFailed, "failed to marshal args", err)
		}
		argsJSON = string(data)
	}

	pid := sql.NullInt64{Valid: app.PID > 0, Int64: int64(app.PID)}

	return &repository.ManagedTarget{
		ID:             app.ID,
		Name:           app.Name,
		Type:           "custom_app",
		ExecutablePath: app.ExecutablePath,
		WorkingDir:     app.WorkingDir,
		ArgsJSON:       argsJSON,
		StartCommand:   "",
		StopCommand:    app.StopCommand,
		AutoStart:      app.AutoStart,
		HealthCheckJSON: "",
		PID:            pid,
		LastStartedAt:  sql.NullString{Valid: app.LastStartedAt != "", String: app.LastStartedAt},
		LastStoppedAt:  sql.NullString{Valid: app.LastStoppedAt != "", String: app.LastStoppedAt},
		LastError:      sql.NullString{Valid: app.LastError != "", String: app.LastError},
	}, nil
}
