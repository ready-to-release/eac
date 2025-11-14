// File: src/core/ai/logging_test.go
package ai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/src/core/ai"
)

func TestFileLogger_LogExecution(t *testing.T) {
	tmpDir := t.TempDir()
	logger := ai.NewFileLogger(tmpDir)

	tests := []struct {
		name      string
		entry     *ai.LogEntry
		wantFields []string
	}{
		{
			name: "successful execution logged",
			entry: &ai.LogEntry{
				Provider:    "claude-cli",
				Input:       "test prompt",
				Response:    "test response",
				Success:     true,
				DidFallback: false,
			},
			wantFields: []string{"timestamp", "provider", "success", "input", "response"},
		},
		{
			name: "failed execution logged",
			entry: &ai.LogEntry{
				Provider:    "claude-api",
				Input:       "test prompt",
				Success:     false,
				Error:       fmt.Errorf("API error"),
				DidFallback: true,
			},
			wantFields: []string{"timestamp", "provider", "success", "error", "didFallback"},
		},
		{
			name: "long input truncated",
			entry: &ai.LogEntry{
				Provider:    "claude-cli",
				Input:       strings.Repeat("a", 300),
				Response:    "response",
				Success:     true,
				DidFallback: false,
			},
			wantFields: []string{"timestamp", "provider", "input"},
		},
		{
			name: "long response truncated",
			entry: &ai.LogEntry{
				Provider:    "claude-cli",
				Input:       "input",
				Response:    strings.Repeat("b", 600),
				Success:     true,
				DidFallback: false,
			},
			wantFields: []string{"timestamp", "provider", "response"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Log execution
			logger.LogExecution(ctx, tt.entry)

			// Read log file
			logPath := filepath.Join(tmpDir, ".r2r", "logs", "ai-executions.jsonl")
			content, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			// Parse last line (latest log entry)
			lines := strings.Split(strings.TrimSpace(string(content)), "\n")
			lastLine := lines[len(lines)-1]

			var logData map[string]interface{}
			if err := json.Unmarshal([]byte(lastLine), &logData); err != nil {
				t.Fatalf("Failed to parse log entry: %v", err)
			}

			// Verify expected fields exist
			for _, field := range tt.wantFields {
				if _, exists := logData[field]; !exists {
					t.Errorf("Log entry missing field: %s", field)
				}
			}

			// Verify provider matches
			if provider, ok := logData["provider"].(string); !ok || provider != tt.entry.Provider {
				t.Errorf("Provider = %v, want %v", logData["provider"], tt.entry.Provider)
			}

			// Verify success status matches
			if success, ok := logData["success"].(bool); !ok || success != tt.entry.Success {
				t.Errorf("Success = %v, want %v", logData["success"], tt.entry.Success)
			}
		})
	}
}

func TestFileLogger_CreateLogDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Log directory doesn't exist yet
	logDir := filepath.Join(tmpDir, ".r2r", "logs")
	if _, err := os.Stat(logDir); !os.IsNotExist(err) {
		t.Fatal("Log directory should not exist yet")
	}

	// Create logger and log entry
	logger := ai.NewFileLogger(tmpDir)
	logger.LogExecution(context.Background(), &ai.LogEntry{
		Provider: "test",
		Input:    "test",
		Success:  true,
	})

	// Verify log directory was created
	if _, err := os.Stat(logDir); err != nil {
		t.Errorf("Log directory was not created: %v", err)
	}

	// Verify log file exists
	logPath := filepath.Join(logDir, "ai-executions.jsonl")
	if _, err := os.Stat(logPath); err != nil {
		t.Errorf("Log file was not created: %v", err)
	}
}

func TestFileLogger_AppendMode(t *testing.T) {
	tmpDir := t.TempDir()
	logger := ai.NewFileLogger(tmpDir)
	ctx := context.Background()

	// Log first entry
	logger.LogExecution(ctx, &ai.LogEntry{
		Provider: "provider1",
		Input:    "input1",
		Success:  true,
	})

	// Log second entry
	logger.LogExecution(ctx, &ai.LogEntry{
		Provider: "provider2",
		Input:    "input2",
		Success:  true,
	})

	// Read log file
	logPath := filepath.Join(tmpDir, ".r2r", "logs", "ai-executions.jsonl")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Verify both entries exist
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines, got %d", len(lines))
	}

	// Parse first entry
	var entry1 map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &entry1); err != nil {
		t.Fatalf("Failed to parse first entry: %v", err)
	}
	if entry1["provider"] != "provider1" {
		t.Errorf("First entry provider = %v, want provider1", entry1["provider"])
	}

	// Parse second entry
	var entry2 map[string]interface{}
	if err := json.Unmarshal([]byte(lines[1]), &entry2); err != nil {
		t.Fatalf("Failed to parse second entry: %v", err)
	}
	if entry2["provider"] != "provider2" {
		t.Errorf("Second entry provider = %v, want provider2", entry2["provider"])
	}
}
