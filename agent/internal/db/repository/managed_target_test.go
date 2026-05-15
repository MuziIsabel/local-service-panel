package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/local-service-panel/agent/internal/db"
)

func setupTestDB(t *testing.T) *db.DB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "lsp-repo-test-*")
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

func TestManagedTargetCRUD(t *testing.T) {
	database := setupTestDB(t)
	repo := NewManagedTargetRepo(database)

	// Create
	target := &ManagedTarget{
		Name:           "Test App",
		Type:           "custom_app",
		ExecutablePath: "C:\\test\\app.exe",
		WorkingDir:     "C:\\test",
		ArgsJSON:       `["--port","8080"]`,
		AutoStart:      false,
	}
	if err := repo.Create(target); err != nil {
		t.Fatalf("create: %v", err)
	}
	if target.ID == "" {
		t.Fatal("expected ID to be set")
	}
	if target.CreatedAt == "" {
		t.Fatal("expected created_at to be set")
	}

	// GetByID
	got, err := repo.GetByID(target.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Name != "Test App" {
		t.Errorf("name = %q, want %q", got.Name, "Test App")
	}
	if got.Type != "custom_app" {
		t.Errorf("type = %q, want %q", got.Type, "custom_app")
	}
	if got.AutoStart {
		t.Error("auto_start should be false")
	}

	// List
	targets, err := repo.List("custom_app")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("list count = %d, want 1", len(targets))
	}
	if targets[0].ID != target.ID {
		t.Errorf("list id = %q, want %q", targets[0].ID, target.ID)
	}

	// List all
	allTargets, err := repo.List("")
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(allTargets) != 1 {
		t.Fatalf("list all count = %d, want 1", len(allTargets))
	}

	// Delete
	if err := repo.Delete(target.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Verify deleted
	_, err = repo.GetByID(target.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestManagedTargetAutoStart(t *testing.T) {
	database := setupTestDB(t)
	repo := NewManagedTargetRepo(database)

	target := &ManagedTarget{
		Name:           "Auto App",
		Type:           "custom_app",
		ExecutablePath: "C:\\auto\\app.exe",
		AutoStart:      true,
	}
	if err := repo.Create(target); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByID(target.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if !got.AutoStart {
		t.Error("auto_start should be true")
	}
}

func TestManagedTargetUpdate(t *testing.T) {
	database := setupTestDB(t)
	repo := NewManagedTargetRepo(database)

	target := &ManagedTarget{
		Name:           "Original",
		Type:           "custom_app",
		ExecutablePath: "C:\\original\\app.exe",
	}
	if err := repo.Create(target); err != nil {
		t.Fatalf("create: %v", err)
	}

	target.Name = "Updated"
	target.ExecutablePath = "C:\\updated\\app.exe"
	if err := repo.Update(target); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := repo.GetByID(target.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Name != "Updated" {
		t.Errorf("name = %q, want %q", got.Name, "Updated")
	}
	if got.ExecutablePath != "C:\\updated\\app.exe" {
		t.Errorf("executablePath = %q", got.ExecutablePath)
	}
}

func TestManagedTargetListByKeyword(t *testing.T) {
	database := setupTestDB(t)
	repo := NewManagedTargetRepo(database)

	targets := []*ManagedTarget{
		{Name: "My App", Type: "custom_app", ExecutablePath: "C:\\a\\app.exe"},
		{Name: "My Service", Type: "windows_service", ExecutablePath: "C:\\b\\svc.exe"},
		{Name: "Another App", Type: "custom_app", ExecutablePath: "C:\\c\\app.exe"},
	}
	for _, tgt := range targets {
		if err := repo.Create(tgt); err != nil {
			t.Fatalf("create %s: %v", tgt.Name, err)
		}
	}

	// Search by keyword without type filter
	results, err := repo.ListByKeyword("App", "")
	if err != nil {
		t.Fatalf("list by keyword: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("keyword 'App' returned %d results, want 2", len(results))
	}

	// Search by keyword with custom_app type
	results, err = repo.ListByKeyword("App", "custom_app")
	if err != nil {
		t.Fatalf("list by keyword + type: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("keyword 'App' + type custom_app returned %d results, want 2", len(results))
	}

	// Search by keyword with windows_service type
	results, err = repo.ListByKeyword("Service", "windows_service")
	if err != nil {
		t.Fatalf("list by keyword + type: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("keyword 'Service' + type windows_service returned %d results, want 1", len(results))
	}
}
