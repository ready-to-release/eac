//go:build L1 && ov

package drawio

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTranslateToContainerPath verifies path translation to container format.
func TestTranslateToContainerPath(t *testing.T) {
	// Use appropriate path format based on OS
	var workspaceRoot string
	if runtime.GOOS == "windows" {
		workspaceRoot = "C:\\workspace"
	} else {
		workspaceRoot = "/workspace"
	}

	tests := []struct {
		name          string
		localPath     string
		workspaceRoot string
		want          string
		wantErr       bool
	}{
		{
			name:          "absolute path within workspace",
			localPath:     filepath.Join(workspaceRoot, "docs", "diagram.drawio.png"),
			workspaceRoot: workspaceRoot,
			want:          "/docs/docs/diagram.drawio.png",
			wantErr:       false,
		},
		{
			name:          "relative path",
			localPath:     filepath.Join("docs", "diagram.drawio.png"),
			workspaceRoot: workspaceRoot,
			want:          "/docs/docs/diagram.drawio.png",
			wantErr:       false,
		},
		{
			name:          "file in root",
			localPath:     "diagram.drawio.png",
			workspaceRoot: workspaceRoot,
			want:          "/docs/diagram.drawio.png",
			wantErr:       false,
		},
		{
			name:          "nested path",
			localPath:     filepath.Join("docs", "assets", "images", "arch.drawio.png"),
			workspaceRoot: workspaceRoot,
			want:          "/docs/docs/assets/images/arch.drawio.png",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TranslateToContainerPath(tt.localPath, tt.workspaceRoot)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestTranslateToContainerPath_OutsideWorkspace verifies error for paths outside workspace.
func TestTranslateToContainerPath_OutsideWorkspace(t *testing.T) {
	var workspaceRoot, outsidePath string
	if runtime.GOOS == "windows" {
		workspaceRoot = "C:\\workspace"
		outsidePath = "C:\\other\\file.png"
	} else {
		workspaceRoot = "/workspace"
		outsidePath = "/other/file.png"
	}

	_, err := TranslateToContainerPath(outsidePath, workspaceRoot)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outside workspace")
}

// TestTranslateToContainerPath_ParentTraversal verifies error for parent directory traversal.
func TestTranslateToContainerPath_ParentTraversal(t *testing.T) {
	var workspaceRoot string
	if runtime.GOOS == "windows" {
		workspaceRoot = "C:\\workspace"
	} else {
		workspaceRoot = "/workspace"
	}

	// Try to traverse up with relative path
	_, err := TranslateToContainerPath(filepath.Join("..", "outside", "file.png"), workspaceRoot)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outside workspace")
}

// TestContainerWorkdir verifies the container workdir constant.
func TestContainerWorkdir(t *testing.T) {
	assert.Equal(t, "/docs", ContainerWorkdir)
}

// TestDrawioImageName verifies the Docker image name constant.
func TestDrawioImageName(t *testing.T) {
	assert.Equal(t, "cli-drawio-cli:latest", DrawioImageName)
}
