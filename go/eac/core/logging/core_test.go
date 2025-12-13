package logging

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestUnifiedLogFile(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()

	cfg := Config{
		Command:           "create",
		WorkspaceRoot:     tmpDir,
		EnableFileLogging: true,
	}

	// Load default logging config
	logCfg := DefaultLoggingConfig()

	// Build the file core
	core, closer, err := buildFileCore(cfg, logCfg)
	if err != nil {
		t.Fatalf("buildFileCore failed: %v", err)
	}
	defer func() {
		if closer != nil {
			closer.Close()
		}
	}()

	if core == nil {
		t.Fatal("expected core to be non-nil")
	}

	// Write something to trigger file creation (lumberjack creates lazily)
	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Message: "test message",
	}
	if err := core.Write(entry, nil); err != nil {
		t.Fatalf("failed to write log entry: %v", err)
	}
	core.Sync()

	// Verify the file was created at out/commands.log
	expectedPath := filepath.Join(tmpDir, "out", "commands.log")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected log file at %s, but it does not exist", expectedPath)
	}
}

func TestTargetFileCore(t *testing.T) {
	tests := []struct {
		name         string
		command      string
		unit         string
		expectTarget bool
		expectedPath string
	}{
		{
			name:         "build with unit creates target log",
			command:      "build",
			unit:         "eac-core",
			expectTarget: true,
			expectedPath: "out/build/eac-core/build.log",
		},
		{
			name:         "test with unit creates target log",
			command:      "test",
			unit:         "component",
			expectTarget: true,
			expectedPath: "out/test/component/test.log",
		},
		{
			name:         "build without unit returns nil",
			command:      "build",
			unit:         "",
			expectTarget: false,
		},
		{
			name:         "create command returns nil (no target configured)",
			command:      "create",
			unit:         "something",
			expectTarget: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := Config{
				Command:           tt.command,
				Unit:              tt.unit,
				WorkspaceRoot:     tmpDir,
				EnableFileLogging: true,
			}

			// Create logging config with targets
			logCfg := LoggingConfig{
				Targets: map[string]TargetConfig{
					"build": {
						Path:      "out/build/{unit}/build.log",
						Levels:    []string{"debug", "info", "warn", "error"},
						Formatter: FormatterJSON,
					},
					"test": {
						Path:      "out/test/{unit}/test.log",
						Levels:    []string{"debug", "info", "warn", "error"},
						Formatter: FormatterJSON,
					},
				},
			}

			core, file, err := buildTargetFileCore(cfg, logCfg)
			if err != nil {
				t.Fatalf("buildTargetFileCore failed: %v", err)
			}

			if tt.expectTarget {
				if core == nil {
					t.Fatal("expected target core to be non-nil")
				}
				defer file.Close()

				// Verify the file was created at expected path
				expectedPath := filepath.Join(tmpDir, tt.expectedPath)
				if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
					t.Errorf("expected target log file at %s, but it does not exist", expectedPath)
				}
			} else {
				if core != nil {
					t.Error("expected target core to be nil")
					if file != nil {
						file.Close()
					}
				}
			}
		})
	}
}
