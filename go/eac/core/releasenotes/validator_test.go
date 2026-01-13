package releasenotes

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

func TestParseContent(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantVersions int
		wantFirstVer string
		wantSections int
		wantErr      bool
	}{
		{
			name: "valid release notes with single version",
			content: `# Release release_notes

## [0.0.7] - 2025-12-11

### Conclusion on Fitness for Intended Use

The changes are fit for use.

### Impact on Business Process

No impact.
`,
			wantVersions: 1,
			wantFirstVer: "0.0.7",
			wantSections: 2,
			wantErr:      false,
		},
		{
			name: "valid release notes with multiple versions",
			content: `# Release release_notes

## [0.0.7] - 2025-12-11

### Conclusion on Fitness for Intended Use

Version 7 content.

## [0.0.6] - 2025-12-10

### Conclusion on Fitness for Intended Use

Version 6 content.
`,
			wantVersions: 2,
			wantFirstVer: "0.0.7",
			wantSections: 1,
			wantErr:      false,
		},
		{
			name: "empty file",
			content: `# Release release_notes
`,
			wantVersions: 0,
			wantErr:      false,
		},
		{
			name: "version with no sections",
			content: `# Release release_notes

## [0.0.7] - 2025-12-11
`,
			wantVersions: 1,
			wantFirstVer: "0.0.7",
			wantSections: 0,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rn, err := ParseContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if len(rn.Versions) != tt.wantVersions {
					t.Errorf("ParseContent() got %d versions, want %d", len(rn.Versions), tt.wantVersions)
				}

				if tt.wantVersions > 0 {
					if rn.Versions[0].Number != tt.wantFirstVer {
						t.Errorf("ParseContent() first version = %s, want %s", rn.Versions[0].Number, tt.wantFirstVer)
					}

					if len(rn.Versions[0].Sections) != tt.wantSections {
						t.Errorf("ParseContent() got %d sections, want %d", len(rn.Versions[0].Sections), tt.wantSections)
					}
				}
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		content string
		version string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid version with content",
			content: `# Release release_notes

## [0.0.7] - 2025-12-11

### Conclusion on Fitness for Intended Use

Content here.
`,
			version: "0.0.7",
			wantErr: false,
		},
		{
			name: "version not found",
			content: `# Release release_notes

## [0.0.6] - 2025-12-10

### Section

Content.
`,
			version: "0.0.7",
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "version exists but no sections",
			content: `# Release release_notes

## [0.0.7] - 2025-12-11
`,
			version: "0.0.7",
			wantErr: true,
			errMsg:  "has no sections",
		},
		{
			name: "version exists but sections are empty",
			content: `# Release release_notes

## [0.0.7] - 2025-12-11

### Section One

### Section Two
`,
			version: "0.0.7",
			wantErr: true,
			errMsg:  "all sections are empty",
		},
		{
			name: "version with TODO placeholders is valid",
			content: `# Release release_notes

## [0.0.7] - 2025-12-11

### Conclusion on Fitness for Intended Use

[TODO: Provide conclusion]
`,
			version: "0.0.7",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rn, err := ParseContent(tt.content)
			if err != nil {
				t.Fatalf("ParseContent() failed: %v", err)
			}

			err = rn.ValidateVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVersion() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateVersion() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	content := `# Release release_notes

## [0.0.7] - 2025-12-11

### Section One

Content for version 7.

## [0.0.6] - 2025-12-10

### Section Two

Content for version 6.
`

	rn, err := ParseContent(content)
	if err != nil {
		t.Fatalf("ParseContent() failed: %v", err)
	}

	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "get existing version 0.0.7",
			version: "0.0.7",
			wantErr: false,
		},
		{
			name:    "get existing version 0.0.6",
			version: "0.0.6",
			wantErr: false,
		},
		{
			name:    "get non-existing version",
			version: "0.0.5",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := rn.GetVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVersion() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && v == nil {
				t.Errorf("GetVersion() returned nil version")
			}

			if !tt.wantErr && v.Number != tt.version {
				t.Errorf("GetVersion() version = %s, want %s", v.Number, tt.version)
			}
		})
	}
}

func TestGenerateTemplate(t *testing.T) {
	// Load config to get workspace root and template paths
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create temp directory
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		module  string
		version string
		date    time.Time
		wantErr bool
	}{
		{
			name:    "generate valid template",
			path:    filepath.Join(tmpDir, "release", "test-module", "RELEASE-NOTES.md"),
			module:  "test-module",
			version: "0.0.7",
			date:    time.Date(2025, 12, 11, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "generate in non-existent directory",
			path:    filepath.Join(tmpDir, "nested", "dir", "RELEASE-NOTES.md"),
			module:  "another-module",
			version: "1.0.0",
			date:    time.Now(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GenerateTemplate(cfg.RepoRoot, cfg.Repository, tt.path, tt.module, tt.version, tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file exists
				if _, err := os.Stat(tt.path); os.IsNotExist(err) {
					t.Errorf("GenerateTemplate() did not create file at %s", tt.path)
					return
				}

				// Read and parse the generated file
				rn, err := Parse(tt.path)
				if err != nil {
					t.Errorf("Parse() generated file failed: %v", err)
					return
				}

				// Verify version exists
				if len(rn.Versions) != 1 {
					t.Errorf("Generated file has %d versions, want 1", len(rn.Versions))
				}

				if rn.Versions[0].Number != tt.version {
					t.Errorf("Generated version = %s, want %s", rn.Versions[0].Number, tt.version)
				}

				// Verify sections exist (template has 3 sections: Summary, Conclusion, Impact)
				if len(rn.Versions[0].Sections) != 3 {
					t.Errorf("Generated file has %d sections, want 3", len(rn.Versions[0].Sections))
				}

				// Verify expected section headers
				expectedSections := []string{"Summary", "Conclusion on Fitness for Intended Use", "Impact on Business Process"}
				for i, expectedHeader := range expectedSections {
					if i < len(rn.Versions[0].Sections) {
						if rn.Versions[0].Sections[i].Header != expectedHeader {
							t.Errorf("Section %d header = %q, want %q", i, rn.Versions[0].Sections[i].Header, expectedHeader)
						}
					}
				}
			}
		})
	}
}

func TestParse_FileOperations(t *testing.T) {
	tmpDir := t.TempDir()

	content := `# Release release_notes

## [0.0.7] - 2025-12-11

### Test Section

Test content.
`

	// Write test file
	testFile := filepath.Join(tmpDir, "RELEASE-NOTES.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test Parse with file
	rn, err := Parse(testFile)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(rn.Versions) != 1 {
		t.Errorf("Parse() got %d versions, want 1", len(rn.Versions))
	}

	// Test Parse with non-existent file
	_, err = Parse(filepath.Join(tmpDir, "non-existent.md"))
	if err == nil {
		t.Error("Parse() should fail for non-existent file")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
