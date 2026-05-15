// Package logging provides structured logging for the Agent.
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Level represents log severity.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Logger is a minimal structured logger.
type Logger struct {
	level  Level
	format string
	out    io.Writer
}

// Entry represents a single structured log entry.
type Entry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

// New creates a new Logger.
// format: "json" or "text"
// level: "debug", "info", "warn", "error"
func New(format, level, logDir string) (*Logger, error) {
	l := &Logger{
		level:  Level(level),
		format: format,
		out:    os.Stdout,
	}

	// If logDir is provided, also write to file.
	if logDir != "" {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("create log directory: %w", err)
		}
		logFile := filepath.Join(logDir, "agent.log")
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		l.out = io.MultiWriter(os.Stdout, f)
	}

	return l, nil
}

func (l *Logger) shouldLog(lvl Level) bool {
	levels := map[Level]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}
	return levels[lvl] >= levels[l.level]
}

func (l *Logger) log(lvl Level, format string, args ...interface{}) {
	if !l.shouldLog(lvl) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	now := time.Now().Format(time.RFC3339)

	if l.format == "json" {
		entry := Entry{
			Time:    now,
			Level:   string(lvl),
			Message: msg,
		}
		data, _ := json.Marshal(entry)
		fmt.Fprintln(l.out, string(data))
	} else {
		fmt.Fprintf(l.out, "%s [%s] %s\n", now, string(lvl), msg)
	}
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message.
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}
