package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitLogger(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		debug bool
	}{
		{"debug mode enabled", true},
		{"debug mode disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := InitLogger(tt.debug, tmpDir)
			if err != nil {
				t.Fatalf("InitLogger() error = %v", err)
			}
			if logger == nil {
				t.Fatal("InitLogger() returned nil logger")
			}
			defer logger.Sync()

			if logger.IsDebugMode() != tt.debug {
				t.Errorf("IsDebugMode() = %v, want %v", logger.IsDebugMode(), tt.debug)
			}
		})
	}
}

func TestWriteDebugFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("writes file in debug mode", func(t *testing.T) {
		logger, _ := InitLogger(true, tmpDir)
		defer logger.Sync()

		WriteDebugFile(logger, tmpDir, "test.txt", "test content")

		expectedPath := filepath.Join(tmpDir, "out", "logs", "work", "test.txt")
		content, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("Failed to read debug file: %v", err)
		}

		if string(content) != "test content" {
			t.Errorf("File content = %q, want %q", string(content), "test content")
		}
	})

	t.Run("does not write file when not in debug mode", func(t *testing.T) {
		logger, _ := InitLogger(false, tmpDir)
		defer logger.Sync()

		WriteDebugFile(logger, tmpDir, "should-not-exist.txt", "content")

		expectedPath := filepath.Join(tmpDir, "out", "logs", "work", "should-not-exist.txt")
		if _, err := os.Stat(expectedPath); !os.IsNotExist(err) {
			t.Error("File should not exist in non-debug mode")
		}
	})

	t.Run("handles nil logger gracefully", func(t *testing.T) {
		// Should not panic
		WriteDebugFile(nil, tmpDir, "test.txt", "content")
	})
}

func TestWriteDebugFilef(t *testing.T) {
	tmpDir := t.TempDir()
	logger, _ := InitLogger(true, tmpDir)
	defer logger.Sync()

	WriteDebugFilef(logger, tmpDir, "test-%d.txt", "test content", 123)

	expectedPath := filepath.Join(tmpDir, "out", "logs", "work", "test-123.txt")
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read debug file: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("File content = %q, want %q", string(content), "test content")
	}
}
