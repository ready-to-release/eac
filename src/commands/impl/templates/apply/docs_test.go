package apply

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplatesApplyDocs_DefaultBehavior(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantSource  string
		wantDest    string
		wantLocal   bool
		wantErr     bool
	}{
		{
			name:       "uses defaults when no flags provided",
			args:       []string{},
			wantSource: "", // Empty means use default GitHub repo
			wantDest:   ".docs/references/docs",
			wantLocal:  false,
			wantErr:    false,
		},
		{
			name:       "custom local source path",
			args:       []string{"--source", "./custom/templates"},
			wantSource: "./custom/templates",
			wantDest:   ".docs/references/docs",
			wantLocal:  true,
			wantErr:    false,
		},
		{
			name:       "custom destination path",
			args:       []string{"--destination", "./custom/path"},
			wantSource: "", // Empty means use default GitHub repo
			wantDest:   "./custom/path",
			wantLocal:  false,
			wantErr:    false,
		},
		{
			name:       "custom local source and destination",
			args:       []string{"--source", "./local/templates", "--destination", "./output"},
			wantSource: "./local/templates",
			wantDest:   "./output",
			wantLocal:  true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parseDocsFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDocsFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// For default (no custom source), Source should be empty
			// For local source, we expect the path to be resolved against working directory
			if config.Source != tt.wantSource && !tt.wantLocal {
				t.Errorf("Source = %v, want %v", config.Source, tt.wantSource)
			}

			if config.Destination != tt.wantDest {
				t.Errorf("Destination = %v, want %v", config.Destination, tt.wantDest)
			}

			if config.isLocalSrc != tt.wantLocal {
				t.Errorf("isLocalSrc = %v, want %v", config.isLocalSrc, tt.wantLocal)
			}
		})
	}
}

func TestTemplatesApplyDocs_WithValueReplacement(t *testing.T) {
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
			config, err := parseDocsFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDocsFlags() error = %v, wantErr %v", err, tt.wantErr)
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

func TestTemplatesApplyDocs_PathResolution(t *testing.T) {
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
