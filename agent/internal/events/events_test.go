package events

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/local-service-panel/agent/internal/db"
	"github.com/user/local-service-panel/agent/internal/db/repository"
)

func setupTestDB(t *testing.T) *db.DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "lsp-events-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	database, err := db.Open(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func setupTestRepo(t *testing.T) *repository.EventLogRepo {
	t.Helper()
	database := setupTestDB(t)
	return repository.NewEventLogRepo(database)
}

func TestRecordAndList(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewService(repo)

	// Record several events
	svc.Record("target-1", "custom_app", "CUSTOM_APP_STARTED", StatusSuccess, "Started", "")
	svc.Record("target-1", "custom_app", "CUSTOM_APP_STOPPED", StatusSuccess, "Stopped", "")
	svc.Record("target-2", "windows_service", "WINDOWS_SERVICE_START_FAILED", StatusFailed, "Failed to start", "access denied")

	// List all
	dtos, err := svc.List(Filter{Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(dtos) != 3 {
		t.Fatalf("expected 3 events, got %d", len(dtos))
	}

	// Most recent first (DESC order).
	// If events have same timestamp, order is not guaranteed, so check all are present.
	actions := make(map[string]bool)
	for _, d := range dtos {
		actions[d.Action] = true
	}
	for _, a := range []string{"CUSTOM_APP_STARTED", "CUSTOM_APP_STOPPED", "WINDOWS_SERVICE_START_FAILED"} {
		if !actions[a] {
			t.Errorf("missing action %q", a)
		}
	}

	// Filter by target type
	filtered, err := svc.List(Filter{Limit: 10, TargetType: "custom_app"})
	if err != nil {
		t.Fatalf("List filtered: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 custom_app events, got %d", len(filtered))
	}

	// Filter by status
	failed, err := svc.List(Filter{Limit: 10, Status: StatusFailed})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(failed) != 1 {
		t.Fatalf("expected 1 failed event, got %d", len(failed))
	}
	if failed[0].Action != "WINDOWS_SERVICE_START_FAILED" {
		t.Errorf("failed action = %q", failed[0].Action)
	}
}

func TestEventDTORoundTrip(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewService(repo)

	svc.Record("test-id", "test-type", "TEST_ACTION", StatusInfo, "test message", "detail info")

	dtos, err := svc.List(Filter{Limit: 1})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(dtos) != 1 {
		t.Fatalf("expected 1 event, got %d", len(dtos))
	}

	dto := dtos[0]
	if dto.TargetID != "test-id" {
		t.Errorf("TargetID = %q", dto.TargetID)
	}
	if dto.Action != "TEST_ACTION" {
		t.Errorf("Action = %q", dto.Action)
	}
	if dto.Status != StatusInfo {
		t.Errorf("Status = %q", dto.Status)
	}
	if dto.Message != "test message" {
		t.Errorf("Message = %q", dto.Message)
	}
	if dto.Details != "detail info" {
		t.Errorf("Details = %q", dto.Details)
	}
	if dto.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}
}

func TestListWithLimit(t *testing.T) {
	repo := setupTestRepo(t)
	svc := NewService(repo)

	for i := 0; i < 5; i++ {
		svc.Record("t", "type", "ACTION", StatusInfo, "", "")
	}

	dtos, err := svc.List(Filter{Limit: 3})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(dtos) != 3 {
		t.Fatalf("expected 3 events with limit=3, got %d", len(dtos))
	}
}
