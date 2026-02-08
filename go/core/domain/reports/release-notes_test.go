package reports

import (
	"path/filepath"
	"testing"

	coretesting "github.com/ready-to-release/eac/go/core/testing"
)

func TestGetReleaseNotes(t *testing.T) {
	tests := []struct {
		name           string
		module         string
		wantErr        bool
		errContains    string
		checkResult    bool
		expectedModule string
	}{
		{
			name:           "valid module with release notes",
			module:         "eac-ext",
			wantErr:        false,
			checkResult:    true,
			expectedModule: "eac-ext",
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
			report, err := GetReleaseNotes(workspaceRoot, tt.module)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetReleaseNotes() expected error but got none")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("GetReleaseNotes() error = %v, should contain %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("GetReleaseNotes() unexpected error = %v", err)
				return
			}

			if tt.checkResult {
				if report == nil {
					t.Errorf("GetReleaseNotes() returned nil report")
					return
				}
				if report.Module != tt.expectedModule {
					t.Errorf("GetReleaseNotes() module = %v, want %v", report.Module, tt.expectedModule)
				}
				if report.ReleaseNotes == nil {
					t.Errorf("GetReleaseNotes() release notes is nil")
				}
				if report.Path == "" {
					t.Errorf("GetReleaseNotes() path is empty")
				}
				if !filepath.IsAbs(report.Path) {
					t.Errorf("GetReleaseNotes() path is not absolute: %v", report.Path)
				}
			}
		})
	}
}
