package show

import (
	"os"
	"testing"

	coretesting "github.com/ready-to-release/eac/go/eac/core/testing"
)

func TestShowChangelog(t *testing.T) {
	coretesting.SetupWorkspaceIsolation(t)

	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		expectedExit int
	}{
		{
			name:         "no arguments",
			args:         []string{"eac", "show", "changelog"},
			wantErr:      true,
			expectedExit: 1,
		},
		{
			name:         "valid module",
			args:         []string{"eac", "show", "changelog", "ext-eac"},
			wantErr:      false,
			expectedExit: 0,
		},
		{
			name:         "invalid module",
			args:         []string{"eac", "show", "changelog", "non-existent-module"},
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
			exitCode := ShowChangelog()

			if exitCode != tt.expectedExit {
				t.Errorf("ShowChangelog() exitCode = %v, want %v", exitCode, tt.expectedExit)
			}
		})
	}
}
