package get

import (
	"os"
	"testing"

	coretesting "github.com/ready-to-release/eac/go/core/testing"
)

func TestGetReleaseNotes(t *testing.T) {
	coretesting.SetupWorkspaceIsolation(t)

	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		expectedExit int
	}{
		{
			name:         "no arguments",
			args:         []string{"eac", "get", "release-notes"},
			wantErr:      true,
			expectedExit: 1,
		},
		{
			name:         "valid module",
			args:         []string{"eac", "get", "release-notes", "ext-eac"},
			wantErr:      false,
			expectedExit: 0,
		},
		{
			name:         "invalid module",
			args:         []string{"eac", "get", "release-notes", "non-existent-module"},
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
			exitCode := GetReleaseNotes()

			if exitCode != tt.expectedExit {
				t.Errorf("GetReleaseNotes() exitCode = %v, want %v", exitCode, tt.expectedExit)
			}
		})
	}
}
