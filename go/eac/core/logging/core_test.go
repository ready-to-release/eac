package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogFileNaming(t *testing.T) {
	tests := []struct {
		name             string
		command          string
		pathSegments     []string
		expectedFilename string
		expectedDir      string
	}{
		{
			name:             "simple command",
			command:          "design",
			pathSegments:     []string{},
			expectedFilename: "design.log",
			expectedDir:      "out/design",
		},
		{
			name:             "module-aware command",
			command:          "build",
			pathSegments:     []string{"eac-core"},
			expectedFilename: "build-eac-core.log",
			expectedDir:      "out/build/eac-core",
		},
		{
			name:             "subcommand",
			command:          "templates",
			pathSegments:     []string{"apply"},
			expectedFilename: "templates-apply.log",
			expectedDir:      "out/templates/apply",
		},
		{
			name:             "test suite",
			command:          "test",
			pathSegments:     []string{"component"},
			expectedFilename: "test-component.log",
			expectedDir:      "out/test/component",
		},
		{
			name:             "nested path segments",
			command:          "templates",
			pathSegments:     []string{"apply", "docs"},
			expectedFilename: "templates-apply-docs.log",
			expectedDir:      "out/templates/apply/docs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temp directory for testing
			tmpDir := t.TempDir()

			cfg := Config{
				Command:           tt.command,
				PathSegments:      tt.pathSegments,
				WorkspaceRoot:     tmpDir,
				EnableFileLogging: true,
			}

			// Load default logging config
			logCfg := LoadLoggingConfig(tmpDir)

			// Build the file core
			core, file, err := buildFileCore(cfg, logCfg)
			if err != nil {
				t.Fatalf("buildFileCore failed: %v", err)
			}
			defer func() {
				if file != nil {
					file.Close()
				}
			}()

			if core == nil {
				t.Fatal("expected core to be non-nil")
			}

			// Verify the file was created in the expected location
			expectedPath := filepath.Join(tmpDir, tt.expectedDir, tt.expectedFilename)
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Errorf("expected log file at %s, but it does not exist", expectedPath)
			}

			// Verify the filename is correct
			actualFilename := filepath.Base(file.Name())
			if actualFilename != tt.expectedFilename {
				t.Errorf("expected filename %s, got %s", tt.expectedFilename, actualFilename)
			}

			// Verify the directory is correct
			actualDir := filepath.Dir(file.Name())
			expectedDirAbs := filepath.Join(tmpDir, strings.ReplaceAll(tt.expectedDir, "/", string(filepath.Separator)))
			if actualDir != expectedDirAbs {
				t.Errorf("expected directory %s, got %s", expectedDirAbs, actualDir)
			}
		})
	}
}
