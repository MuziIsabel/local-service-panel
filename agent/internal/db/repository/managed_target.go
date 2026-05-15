// Package repository provides data access layer for Agent entities.
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/user/local-service-panel/agent/internal/db"
)

// ManagedTarget represents a user-added custom application record.
type ManagedTarget struct {
	ID             string
	Name           string
	Type           string
	ExecutablePath string
	WorkingDir     string
	ArgsJSON       string
	StartCommand   string
	StopCommand    string
	AutoStart      bool
	HealthCheckJSON string
	PID            sql.NullInt64
	LastStartedAt  sql.NullString
	LastStoppedAt  sql.NullString
	LastError      sql.NullString
	CreatedAt      string
	UpdatedAt      string
}

// ManagedTargetRepo provides CRUD operations for managed_targets.
type ManagedTargetRepo struct {
	db *db.DB
}

// NewManagedTargetRepo creates a new repository.
func NewManagedTargetRepo(database *db.DB) *ManagedTargetRepo {
	return &ManagedTargetRepo{db: database}
}

// Create inserts a new managed target.
func (r *ManagedTargetRepo) Create(t *ManagedTarget) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.CreatedAt = now
	t.UpdatedAt = now

	autoStart := 0
	if t.AutoStart {
		autoStart = 1
	}

	_, err := r.db.Exec(`
		INSERT INTO managed_targets (
			id, name, type, executable_path, working_dir, args_json,
			start_command, stop_command, auto_start, health_check_json,
			pid, last_started_at, last_stopped_at, last_error,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Name, t.Type, t.ExecutablePath, t.WorkingDir, t.ArgsJSON,
		t.StartCommand, t.StopCommand, autoStart, t.HealthCheckJSON,
		t.PID, t.LastStartedAt, t.LastStoppedAt, t.LastError,
		t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert managed target: %w", err)
	}
	return nil
}

// GetByID retrieves a managed target by ID.
func (r *ManagedTargetRepo) GetByID(id string) (*ManagedTarget, error) {
	t := &ManagedTarget{}
	autoStart := 0
	err := r.db.QueryRow(`
		SELECT id, name, type, executable_path, working_dir, args_json,
			start_command, stop_command, auto_start, health_check_json,
			pid, last_started_at, last_stopped_at, last_error,
			created_at, updated_at
		FROM managed_targets WHERE id = ?`, id,
	).Scan(
		&t.ID, &t.Name, &t.Type, &t.ExecutablePath, &t.WorkingDir, &t.ArgsJSON,
		&t.StartCommand, &t.StopCommand, &autoStart, &t.HealthCheckJSON,
		&t.PID, &t.LastStartedAt, &t.LastStoppedAt, &t.LastError,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get managed target %s: %w", id, err)
	}
	t.AutoStart = autoStart == 1
	return t, nil
}

// List returns all managed targets of a given type.
// Pass empty type to return all.
func (r *ManagedTargetRepo) List(targetType string) ([]*ManagedTarget, error) {
	query := `
		SELECT id, name, type, executable_path, working_dir, args_json,
			start_command, stop_command, auto_start, health_check_json,
			pid, last_started_at, last_stopped_at, last_error,
			created_at, updated_at
		FROM managed_targets`
	var args []interface{}
	if targetType != "" {
		query += " WHERE type = ?"
		args = append(args, targetType)
	}
	query += " ORDER BY created_at"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list managed targets: %w", err)
	}
	defer rows.Close()

	var targets []*ManagedTarget
	for rows.Next() {
		t := &ManagedTarget{}
		autoStart := 0
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Type, &t.ExecutablePath, &t.WorkingDir, &t.ArgsJSON,
			&t.StartCommand, &t.StopCommand, &autoStart, &t.HealthCheckJSON,
			&t.PID, &t.LastStartedAt, &t.LastStoppedAt, &t.LastError,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan managed target: %w", err)
		}
		t.AutoStart = autoStart == 1
		targets = append(targets, t)
	}
	return targets, rows.Err()
}

// Delete removes a managed target by ID.
func (r *ManagedTargetRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM managed_targets WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete managed target %s: %w", id, err)
	}
	return nil
}

// Update modifies an existing managed target.
// Only non-zero fields are updated.
func (r *ManagedTargetRepo) Update(t *ManagedTarget) error {
	now := time.Now().UTC().Format(time.RFC3339)
	t.UpdatedAt = now

	autoStart := 0
	if t.AutoStart {
		autoStart = 1
	}

	_, err := r.db.Exec(`
		UPDATE managed_targets SET
			name = ?, type = ?, executable_path = ?, working_dir = ?, args_json = ?,
			start_command = ?, stop_command = ?, auto_start = ?, health_check_json = ?,
			pid = ?, last_started_at = ?, last_stopped_at = ?, last_error = ?,
			updated_at = ?
		WHERE id = ?`,
		t.Name, t.Type, t.ExecutablePath, t.WorkingDir, t.ArgsJSON,
		t.StartCommand, t.StopCommand, autoStart, t.HealthCheckJSON,
		t.PID, t.LastStartedAt, t.LastStoppedAt, t.LastError,
		t.UpdatedAt, t.ID,
	)
	if err != nil {
		return fmt.Errorf("update managed target %s: %w", t.ID, err)
	}
	return nil
}

// ListByKeyword searches managed targets by name with an optional type filter.
func (r *ManagedTargetRepo) ListByKeyword(keyword, targetType string) ([]*ManagedTarget, error) {
	query := `
		SELECT id, name, type, executable_path, working_dir, args_json,
			start_command, stop_command, auto_start, health_check_json,
			pid, last_started_at, last_stopped_at, last_error,
			created_at, updated_at
		FROM managed_targets
		WHERE name LIKE ?`
	var args []interface{}
	args = append(args, "%"+keyword+"%")

	if targetType != "" {
		query += " AND type = ?"
		args = append(args, targetType)
	}
	query += " ORDER BY created_at"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list managed targets by keyword: %w", err)
	}
	defer rows.Close()

	var targets []*ManagedTarget
	for rows.Next() {
		t := &ManagedTarget{}
		autoStart := 0
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Type, &t.ExecutablePath, &t.WorkingDir, &t.ArgsJSON,
			&t.StartCommand, &t.StopCommand, &autoStart, &t.HealthCheckJSON,
			&t.PID, &t.LastStartedAt, &t.LastStoppedAt, &t.LastError,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan managed target: %w", err)
		}
		t.AutoStart = autoStart == 1
		targets = append(targets, t)
	}
	return targets, rows.Err()
}
