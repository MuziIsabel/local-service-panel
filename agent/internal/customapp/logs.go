package customapp

import (
	"bufio"
	"os"
	"path/filepath"
)

// LogsResponse is the response for GET /api/custom-apps/{id}/logs.
type LogsResponse struct {
	Stdout []string `json:"stdout"`
	Stderr []string `json:"stderr"`
}

// ReadLogs reads the last N lines from both stdout.log and stderr.log.
func (pm *ProcessManager) ReadLogs(appID string, lines int) (*LogsResponse, error) {
	appLogDir := filepath.Join(pm.logsDir, appID)

	stdoutLines, _ := tailFile(filepath.Join(appLogDir, "stdout.log"), lines)
	stderrLines, _ := tailFile(filepath.Join(appLogDir, "stderr.log"), lines)

	return &LogsResponse{
		Stdout: stdoutLines,
		Stderr: stderrLines,
	}, nil
}

// tailFile reads the last N lines from a file.
// Returns an empty slice if the file doesn't exist or can't be read.
func tailFile(path string, n int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer f.Close()

	// If n <= 0, set a reasonable default
	if n <= 0 {
		n = 200
	}

	// Read all lines
	var allLines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Return last n lines
	if len(allLines) <= n {
		return allLines, nil
	}
	return allLines[len(allLines)-n:], nil
}
