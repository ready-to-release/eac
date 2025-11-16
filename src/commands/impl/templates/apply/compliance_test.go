package apply

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplatesApplyCompliance_DefaultBehavior(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantSource     string
		wantSourcePath string
		wantDest       string
		wantErr        bool
	}{
		{
			name:           "uses defaults when no flags provided",
			args:           []string{},
			wantSource:     "https://github.com/ready-to-release/eac",
			wantSourcePath: "templates/compliance",
			wantDest:       ".docs/references/compliance",
			wantErr:        false,
		},
		{
			name:           "custom source repository",
			args:           []string{"--source", "https://github.com/custom/repo"},
			wantSource:     "https://github.com/custom/repo",
			wantSourcePath: "templates/compliance",
			wantDest:       ".docs/references/compliance",
			wantErr:        false,
		},
		{
			name:           "custom destination path",
			args:           []string{"--destination", "./custom/path"},
			wantSource:     "https://github.com/ready-to-release/eac",
			wantSourcePath: "templates/compliance",
			wantDest:       "./custom/path",
			wantErr:        false,
		},
		{
			name:           "custom source and destination",
			args:           []string{"--source", "https://github.com/custom/repo", "--destination", "./output"},
			wantSource:     "https://github.com/custom/repo",
			wantSourcePath: "templates/compliance",
			wantDest:       "./output",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parseComplianceFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseComplianceFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if config.Source != tt.wantSource {
				t.Errorf("Source = %v, want %v", config.Source, tt.wantSource)
			}
			if config.SourcePath != tt.wantSourcePath {
				t.Errorf("SourcePath = %v, want %v", config.SourcePath, tt.wantSourcePath)
			}
			if config.Destination != tt.wantDest {
				t.Errorf("Destination = %v, want %v", config.Destination, tt.wantDest)
			}
		})
	}
}

func TestTemplatesApplyCompliance_WithValueReplacement(t *testing.T) {
	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "apply-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create values JSON file
	valuesFile := filepath.Join(tmpDir, "values.json")
	valuesContent := `{"ProjectName": "TestProject", "CompanyName": "ACME"}`
	if err := os.WriteFile(valuesFile, []byte(valuesContent), 0644); err != nil {
		t.Fatalf("Failed to create values file: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		wantValues  bool
		wantErr     bool
		errContains string
	}{
		{
			name:       "no input-json flag",
			args:       []string{},
			wantValues: false,
			wantErr:    false,
		},
		{
			name:       "with input-json flag",
			args:       []string{"--input-json", valuesFile},
			wantValues: true,
			wantErr:    false,
		},
		{
			name:        "input-json file does not exist",
			args:        []string{"--input-json", "nonexistent.json"},
			wantValues:  false,
			wantErr:     true,
			errContains: "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parseComplianceFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseComplianceFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if tt.wantValues && config.ValuesFile == "" {
				t.Error("ValuesFile should be set when input-json is provided")
			}
			if !tt.wantValues && config.ValuesFile != "" {
				t.Error("ValuesFile should be empty when input-json is not provided")
			}
		})
	}
}

func TestTemplatesApplyCompliance_PathResolution(t *testing.T) {
	tests := []struct {
		name         string
		destination  string
		wantAbsolute bool
	}{
		{
			name:         "relative path",
			destination:  "./output",
			wantAbsolute: false,
		},
		{
			name:         "absolute path",
			destination:  filepath.Join(os.TempDir(), "output"),
			wantAbsolute: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isAbs := filepath.IsAbs(tt.destination)
			if isAbs != tt.wantAbsolute {
				t.Errorf("IsAbs() = %v, want %v for path %v", isAbs, tt.wantAbsolute, tt.destination)
			}
		})
	}
}

// Helper function for string contains check
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
