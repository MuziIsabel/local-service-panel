package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAndMigrate(t *testing.T) {
	// Create a temp directory for the test database.
	tmpDir, err := os.MkdirTemp("", "lsp-db-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer database.Close()

	// Verify database is usable by checking a table exists.
	var name string
	err = database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='managed_targets'").Scan(&name)
	if err != nil {
		t.Fatalf("managed_targets table not found: %v", err)
	}
	if name != "managed_targets" {
		t.Errorf("expected table name 'managed_targets', got %q", name)
	}

	// Verify all expected tables exist.
	expectedTables := []string{
		"managed_targets",
		"target_overrides",
		"event_logs",
		"settings",
		"process_runtime",
	}
	for _, table := range expectedTables {
		var found string
		err := database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&found)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestOpenIdempotent(t *testing.T) {
	// Opening the same database twice should succeed (migrations use IF NOT EXISTS).
	tmpDir, err := os.MkdirTemp("", "lsp-db-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	db1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	db1.Close()

	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	db2.Close()
}

func TestDBPath(t *testing.T) {
	got := DBPath("/data")
	want := filepath.Join("/data", "data", "panel.db")
	if got != want {
		t.Errorf("DBPath(/data) = %q, want %q", got, want)
	}
}
