package reports

import (
	"os"
	"strings"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name               string
		module             string
		versionStr         string
		expectedVersion    string
		expectedUnreleased bool
		expectedLatest     bool
		wantErr            bool
		errContains        string
	}{
		{
			name:               "unreleased - explicit",
			module:             "ext-eac",
			versionStr:         "unreleased",
			expectedVersion:    "Unreleased",
			expectedUnreleased: true,
			expectedLatest:     false,
			wantErr:            false,
		},
		{
			name:               "unreleased - implicit (empty)",
			module:             "ext-eac",
			versionStr:         "",
			expectedVersion:    "Unreleased",
			expectedUnreleased: true,
			expectedLatest:     false,
			wantErr:            false,
		},
		{
			name:               "latest version",
			module:             "ext-eac",
			versionStr:         "latest",
			expectedVersion:    "", // Actual version from changelog
			expectedUnreleased: false,
			expectedLatest:     true,
			wantErr:            false,
		},
		{
			name:        "invalid module",
			module:      "nonexistent-module-xyz",
			versionStr:  "latest",
			wantErr:     true,
			errContains: "module not found",
		},
		{
			name:        "version not found",
			module:      "ext-eac",
			versionStr:  "99.99.99",
			wantErr:     true,
			errContains: "version not found",
		},
	}

	// Get repository root
	workspaceRoot := os.Getenv("WORKSPACE_ROOT")
	if workspaceRoot == "" {
		t.Skip("WORKSPACE_ROOT not set, skipping integration test")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ResolveVersion(workspaceRoot, tt.module, tt.versionStr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveVersion() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ResolveVersion() error = %v, should contain %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveVersion() unexpected error = %v", err)
				return
			}

			if info.Module != tt.module {
				t.Errorf("Module = %v, want %v", info.Module, tt.module)
			}

			if tt.expectedVersion != "" && info.VersionNumber != tt.expectedVersion {
				t.Errorf("VersionNumber = %v, want %v", info.VersionNumber, tt.expectedVersion)
			}

			if info.IsUnreleased != tt.expectedUnreleased {
				t.Errorf("IsUnreleased = %v, want %v", info.IsUnreleased, tt.expectedUnreleased)
			}

			if info.IsLatest != tt.expectedLatest {
				t.Errorf("IsLatest = %v, want %v", info.IsLatest, tt.expectedLatest)
			}

			// Verify git tags are properly formatted when set
			if info.GitTag != "" && !strings.Contains(info.GitTag, "/") {
				t.Errorf("GitTag = %v, should contain module/version format", info.GitTag)
			}
		})
	}
}
