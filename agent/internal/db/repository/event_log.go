// Package repository provides data access layer for Agent entities.
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/user/local-service-panel/agent/internal/db"
)

// EventLog represents a single event log record.
type EventLog struct {
	ID         string
	TargetID   string
	TargetType string
	Action     string
	Status     string // "success", "failed", "info"
	Message    string
	Details    string
	CreatedAt  string
}

// EventFilter specifies criteria for querying event logs.
type EventFilter struct {
	Limit      int
	TargetID   string
	TargetType string
	Action     string
	Status     string
}

// EventLogRepo provides CRUD operations for event_logs.
type EventLogRepo struct {
	db *db.DB
}

// NewEventLogRepo creates a new event log repository.
func NewEventLogRepo(database *db.DB) *EventLogRepo {
	return &EventLogRepo{db: database}
}

// Create inserts a new event log.
func (r *EventLogRepo) Create(e *EventLog) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	if e.CreatedAt == "" {
		e.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	_, err := r.db.Exec(`
		INSERT INTO event_logs (id, target_id, target_type, action, status, message, details, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, nz(e.TargetID), nz(e.TargetType), e.Action, e.Status,
		nz(e.Message), nz(e.Details), e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert event log: %w", err)
	}
	return nil
}

// List returns event logs matching the given filter.
func (r *EventLogRepo) List(filter EventFilter) ([]*EventLog, error) {
	query := `SELECT id, target_id, target_type, action, status, message, details, created_at
		FROM event_logs WHERE 1=1`
	var args []interface{}

	if filter.TargetID != "" {
		query += " AND target_id = ?"
		args = append(args, filter.TargetID)
	}
	if filter.TargetType != "" {
		query += " AND target_type = ?"
		args = append(args, filter.TargetType)
	}
	if filter.Action != "" {
		query += " AND action = ?"
		args = append(args, filter.Action)
	}
	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	} else {
		query += " LIMIT 100"
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list event logs: %w", err)
	}
	defer rows.Close()

	var logs []*EventLog
	for rows.Next() {
		e := &EventLog{}
		var tid, ttype, msg, det sql.NullString
		if err := rows.Scan(&e.ID, &tid, &ttype, &e.Action, &e.Status, &msg, &det, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan event log: %w", err)
		}
		e.TargetID = tid.String
		e.TargetType = ttype.String
		e.Message = msg.String
		e.Details = det.String
		logs = append(logs, e)
	}
	return logs, rows.Err()
}

// nz returns empty string if s is empty, useful for nullable DB fields.
func nz(s string) string {
	if s == "" {
		return ""
	}
	return s
}
