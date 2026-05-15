-- 001_init.sql
-- Initial schema for Local Service Panel

CREATE TABLE IF NOT EXISTS managed_targets (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  executable_path TEXT,
  working_dir TEXT,
  args_json TEXT,
  start_command TEXT,
  stop_command TEXT,
  auto_start INTEGER NOT NULL DEFAULT 0,
  health_check_json TEXT,
  pid INTEGER,
  last_started_at TEXT,
  last_stopped_at TEXT,
  last_error TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS target_overrides (
  id TEXT PRIMARY KEY,
  target_type TEXT NOT NULL,
  target_key TEXT NOT NULL,
  display_name TEXT,
  favorite INTEGER NOT NULL DEFAULT 0,
  hidden INTEGER NOT NULL DEFAULT 0,
  note TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(target_type, target_key)
);

CREATE TABLE IF NOT EXISTS event_logs (
  id TEXT PRIMARY KEY,
  target_id TEXT,
  target_type TEXT,
  action TEXT NOT NULL,
  status TEXT NOT NULL,
  message TEXT,
  details TEXT,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS process_runtime (
  target_id TEXT PRIMARY KEY,
  pid INTEGER,
  status TEXT NOT NULL,
  started_at TEXT,
  updated_at TEXT NOT NULL
);
