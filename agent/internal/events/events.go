// Package events provides event log domain types, DTOs, and service.
package events

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/user/local-service-panel/agent/internal/db/repository"
)

// Event status constants.
const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusInfo    = "info"
)

// Event is the domain type for an event log entry.
type Event struct {
	ID         string
	TargetID   string
	TargetType string
	Action     string
	Status     string
	Message    string
	Details    string
	CreatedAt  string
}

// DTO is the API-facing representation of an event log.
type DTO struct {
	ID         string `json:"id"`
	TargetID   string `json:"targetId,omitempty"`
	TargetType string `json:"targetType,omitempty"`
	Action     string `json:"action"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
	Details    string `json:"details,omitempty"`
	CreatedAt  string `json:"createdAt"`
}

// Filter specifies criteria for querying events.
type Filter struct {
	Limit      int
	TargetID   string
	TargetType string
	Action     string
	Status     string
}

// Service provides event log operations.
type Service struct {
	repo *repository.EventLogRepo
}

// NewService creates a new event log service.
func NewService(repo *repository.EventLogRepo) *Service {
	return &Service{repo: repo}
}

// Record creates an event log. Errors are logged to stderr but not returned,
// to avoid blocking the main operation.
func (s *Service) Record(targetID, targetType, action, status, message, details string) {
	e := &Event{
		ID:         uuid.New().String(),
		TargetID:   targetID,
		TargetType: targetType,
		Action:     action,
		Status:     status,
		Message:    message,
		Details:    details,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	repoRecord := &repository.EventLog{
		ID:         e.ID,
		TargetID:   e.TargetID,
		TargetType: e.TargetType,
		Action:     e.Action,
		Status:     e.Status,
		Message:    e.Message,
		Details:    e.Details,
		CreatedAt:  e.CreatedAt,
	}

	if err := s.repo.Create(repoRecord); err != nil {
		// Event write failures should not block the main operation.
		// The error is logged to stderr.
		fmt.Fprintf(os.Stderr, "failed to record event: %v\n", err)
	}
}

// List returns events matching the given filter.
func (s *Service) List(filter Filter) ([]*DTO, error) {
	repoFilter := repository.EventFilter{
		Limit:      filter.Limit,
		TargetID:   filter.TargetID,
		TargetType: filter.TargetType,
		Action:     filter.Action,
		Status:     filter.Status,
	}

	records, err := s.repo.List(repoFilter)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	dtos := make([]*DTO, len(records))
	for i, r := range records {
		dtos[i] = &DTO{
			ID:         r.ID,
			TargetID:   r.TargetID,
			TargetType: r.TargetType,
			Action:     r.Action,
			Status:     r.Status,
			Message:    r.Message,
			Details:    r.Details,
			CreatedAt:  r.CreatedAt,
		}
	}
	return dtos, nil
}
