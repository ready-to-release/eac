package reports

import (
	"path/filepath"
	"testing"

	coretesting "github.com/ready-to-release/eac/go/core/testing"
)

func TestGetChangelog(t *testing.T) {
	tests := []struct {
		name           string
		module         string
		wantErr        bool
		errContains    string
		checkResult    bool
		expectedModule string
	}{
		{
			name:           "valid module with changelog",
			module:         "ext-eac",
			wantErr:        false,
			checkResult:    true,
			expectedModule: "ext-eac",
		},
		{
			name:        "non-existent module",
			module:      "non-existent-module",
			wantErr:     true,
			errContains: "module not found",
		},
	}

	workspaceRoot := coretesting.SetupWorkspaceIsolation(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := GetChangelog(workspaceRoot, tt.module)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetChangelog() expected error but got none")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("GetChangelog() error = %v, should contain %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("GetChangelog() unexpected error = %v", err)
				return
			}

			if tt.checkResult {
				if report == nil {
					t.Errorf("GetChangelog() returned nil report")
					return
				}
				if report.Module != tt.expectedModule {
					t.Errorf("GetChangelog() module = %v, want %v", report.Module, tt.expectedModule)
				}
				if report.Changelog == nil {
					t.Errorf("GetChangelog() changelog is nil")
				}
				if report.Path == "" {
					t.Errorf("GetChangelog() path is empty")
				}
				if !filepath.IsAbs(report.Path) {
					t.Errorf("GetChangelog() path is not absolute: %v", report.Path)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
