// File: src/core/ai/logging.go
// Intent: Log AI execution details to .r2r/logs/ai-executions.jsonl for debugging and audit
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Single LogExecution() method appends one line to JSONL file
//   - Each log entry is self-contained JSON object
//   - Timestamp, provider, success/failure clearly recorded
//
// Easy to change:
//   - JSONL format allows easy parsing and analysis
//   - Log path can be configured
//   - No dependencies on external logging frameworks
//   - Can add new fields without breaking existing logs
//
// Hard to break:
//   - Creates log directory if it doesn't exist
//   - File opened in append mode (no data loss)
//   - JSON marshaling catches serialization errors
//   - Errors logged to stderr if logging fails

package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileLogger logs AI executions to a JSONL file
type FileLogger struct {
	logPath string
}

// NewFileLogger creates a logger that writes to the specified path
func NewFileLogger(workspaceRoot string) *FileLogger {
	logPath := filepath.Join(workspaceRoot, ".r2r", "logs", "ai-executions.jsonl")
	return &FileLogger{logPath: logPath}
}

// LogExecution appends an execution log entry to the JSONL file
func (l *FileLogger) LogExecution(ctx context.Context, entry *LogEntry) {
	// Create logs directory if it doesn't exist
	logDir := filepath.Dir(l.logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to create log directory: %v\n", err)
		return
	}

	// Build log entry
	logData := map[string]interface{}{
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"provider":    entry.Provider,
		"success":     entry.Success,
		"didFallback": entry.DidFallback,
	}

	// Add input (truncated if too long)
	if len(entry.Input) > 200 {
		logData["input"] = entry.Input[:200] + "..."
	} else {
		logData["input"] = entry.Input
	}

	// Add response (truncated if too long)
	if entry.Success && entry.Response != "" {
		if len(entry.Response) > 500 {
			logData["response"] = entry.Response[:500] + "..."
		} else {
			logData["response"] = entry.Response
		}
	}

	// Add error if present
	if entry.Error != nil {
		logData["error"] = entry.Error.Error()
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(logData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to marshal log entry: %v\n", err)
		return
	}

	// Append to file
	f, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to open log file: %v\n", err)
		return
	}
	defer f.Close()

	// Write JSON line
	if _, err := f.Write(append(jsonBytes, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to write log entry: %v\n", err)
		return
	}
}
