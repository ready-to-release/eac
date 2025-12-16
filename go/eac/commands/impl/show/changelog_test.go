package show

import (
	"os"
	"testing"
)

func TestShowChangelog(t *testing.T) {
	// Get repository root
	workspaceRoot := os.Getenv("WORKSPACE_ROOT")
	if workspaceRoot == "" {
		t.Skip("WORKSPACE_ROOT not set, skipping integration test")
	}

	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		expectedExit int
	}{
		{
			name:         "no arguments",
			args:         []string{"show", "changelog"},
			wantErr:      true,
			expectedExit: 1,
		},
		{
			name:         "valid module",
			args:         []string{"show", "changelog", "ext-eac"},
			wantErr:      false,
			expectedExit: 0,
		},
		{
			name:         "invalid module",
			args:         []string{"show", "changelog", "non-existent-module"},
			wantErr:      true,
			expectedExit: 1,
		},
		{
			name:         "valid module with version",
			args:         []string{"show", "changelog", "ext-eac", "Unreleased"},
			wantErr:      false,
			expectedExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore os.Args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = tt.args
			exitCode := ShowChangelog()

			if exitCode != tt.expectedExit {
				t.Errorf("ShowChangelog() exitCode = %v, want %v", exitCode, tt.expectedExit)
			}
		})
	}
}
