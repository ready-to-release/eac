package get

import (
	"os"
	"testing"
)

func TestGetChangelog(t *testing.T) {
	// Get repository root
	workspaceRoot := os.Getenv("WORKSPACE_ROOT")
	if workspaceRoot == "" {
		t.Skip("WORKSPACE_ROOT not set, skipping integration test")
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		expectedExit int
	}{
		{
			name:         "no arguments",
			args:         []string{"get", "changelog"},
			wantErr:      true,
			expectedExit: 1,
		},
		{
			name:         "valid module",
			args:         []string{"get", "changelog", "ext-eac"},
			wantErr:      false,
			expectedExit: 0,
		},
		{
			name:         "invalid module",
			args:         []string{"get", "changelog", "non-existent-module"},
			wantErr:      true,
			expectedExit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore os.Args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = tt.args
			exitCode := GetChangelog()

			if exitCode != tt.expectedExit {
				t.Errorf("GetChangelog() exitCode = %v, want %v", exitCode, tt.expectedExit)
			}
		})
	}
}
