package tool

import (
	"os"
	"testing"
)

// TestExecutionContext_IsDinD tests the IsDinD method.
func TestExecutionContext_IsDinD(t *testing.T) {
	tests := []struct {
		name              string
		hostWorkspaceRoot string
		want              bool
	}{
		{
			name:              "returns false when HostWorkspaceRoot is empty",
			hostWorkspaceRoot: "",
			want:              false,
		},
		{
			name:              "returns true when HostWorkspaceRoot is set (Linux path)",
			hostWorkspaceRoot: "/home/user/project",
			want:              true,
		},
		{
			name:              "returns true when HostWorkspaceRoot is set (Windows path)",
			hostWorkspaceRoot: `C:\projects\eac`,
			want:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ExecutionContext{
				HostWorkspaceRoot: tt.hostWorkspaceRoot,
			}

			got := ctx.IsDinD()
			if got != tt.want {
				t.Errorf("IsDinD() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestExecutionContext_TranslatePathForMount tests path translation for DinD.
func TestExecutionContext_TranslatePathForMount(t *testing.T) {
	tests := []struct {
		name              string
		hostWorkspaceRoot string
		containerRepoRoot string
		containerPath     string
		want              string
	}{
		// Not in DinD mode - path unchanged
		{
			name:              "non-DinD mode returns path unchanged",
			hostWorkspaceRoot: "",
			containerRepoRoot: "",
			containerPath:     "/some/path",
			want:              "/some/path",
		},

		// DinD mode with Linux host
		{
			name:              "DinD Linux host translates container root",
			hostWorkspaceRoot: "/home/user/project",
			containerRepoRoot: "/var/task",
			containerPath:     "/var/task",
			want:              "/home/user/project",
		},
		{
			name:              "DinD Linux host translates subpath",
			hostWorkspaceRoot: "/home/user/project",
			containerRepoRoot: "/var/task",
			containerPath:     "/var/task/src/main.go",
			want:              "/home/user/project/src/main.go",
		},
		{
			name:              "DinD Linux host translates nested path",
			hostWorkspaceRoot: "/home/user/project",
			containerRepoRoot: "/var/task",
			containerPath:     "/var/task/go/eac/core/tool",
			want:              "/home/user/project/go/eac/core/tool",
		},

		// DinD mode with Windows host
		{
			name:              "DinD Windows host translates container root",
			hostWorkspaceRoot: `C:\projects\eac`,
			containerRepoRoot: "/var/task",
			containerPath:     "/var/task",
			want:              `C:\projects\eac`,
		},
		{
			name:              "DinD Windows host translates subpath",
			hostWorkspaceRoot: `C:\projects\eac`,
			containerRepoRoot: "/var/task",
			containerPath:     "/var/task/src/main.go",
			want:              `C:\projects\eac\src\main.go`,
		},
		{
			name:              "DinD Windows host translates nested path",
			hostWorkspaceRoot: `C:\projects\eac`,
			containerRepoRoot: "/var/task",
			containerPath:     "/var/task/go/eac/core/tool",
			want:              `C:\projects\eac\go\eac\core\tool`,
		},
		{
			name:              "DinD Windows host with different drive letter",
			hostWorkspaceRoot: `D:\work\repo`,
			containerRepoRoot: "/var/task",
			containerPath:     "/var/task/docs",
			want:              `D:\work\repo\docs`,
		},

		// Path not under container root - unchanged
		{
			name:              "path not under container root returns unchanged",
			hostWorkspaceRoot: "/home/user/project",
			containerRepoRoot: "/var/task",
			containerPath:     "/tmp/some/other/path",
			want:              "/tmp/some/other/path",
		},
		{
			name:              "path above container root returns unchanged",
			hostWorkspaceRoot: "/home/user/project",
			containerRepoRoot: "/var/task",
			containerPath:     "/var",
			want:              "/var",
		},
		{
			name:              "sibling path returns unchanged",
			hostWorkspaceRoot: "/home/user/project",
			containerRepoRoot: "/var/task",
			containerPath:     "/var/taskdata",
			want:              "/var/taskdata",
		},

		// Custom container root
		{
			name:              "custom container root /workspace",
			hostWorkspaceRoot: "/home/user/project",
			containerRepoRoot: "/workspace",
			containerPath:     "/workspace/src",
			want:              "/home/user/project/src",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ExecutionContext{
				HostWorkspaceRoot: tt.hostWorkspaceRoot,
				ContainerRepoRoot: tt.containerRepoRoot,
			}

			got := ctx.TranslatePathForMount(tt.containerPath)
			if got != tt.want {
				t.Errorf("TranslatePathForMount(%q) = %q, want %q", tt.containerPath, got, tt.want)
			}
		})
	}
}

// TestJoinHostPath tests the joinHostPath helper function.
func TestJoinHostPath(t *testing.T) {
	tests := []struct {
		name     string
		hostRoot string
		relPath  string
		want     string
	}{
		// Unix host paths - uses forward slashes
		{
			name:     "Unix path with relative file",
			hostRoot: "/home/user/project",
			relPath:  "src/main.go",
			want:     "/home/user/project/src/main.go",
		},
		{
			name:     "Unix path with single directory",
			hostRoot: "/home/user/project",
			relPath:  "docs",
			want:     "/home/user/project/docs",
		},
		{
			name:     "Unix path with dot relPath",
			hostRoot: "/home/user/project",
			relPath:  ".",
			want:     "/home/user/project",
		},
		{
			name:     "Unix path with deeply nested",
			hostRoot: "/opt/workspace",
			relPath:  "go/eac/core/tool",
			want:     "/opt/workspace/go/eac/core/tool",
		},

		// Windows host paths - uses backslashes
		{
			name:     "Windows path with relative file",
			hostRoot: `C:\projects\eac`,
			relPath:  "src/main.go",
			want:     `C:\projects\eac\src\main.go`,
		},
		{
			name:     "Windows path with single directory",
			hostRoot: `C:\projects\eac`,
			relPath:  "docs",
			want:     `C:\projects\eac\docs`,
		},
		{
			name:     "Windows path with dot relPath",
			hostRoot: `C:\projects\eac`,
			relPath:  ".",
			want:     `C:\projects\eac`,
		},
		{
			name:     "Windows path with deeply nested",
			hostRoot: `D:\work\repo`,
			relPath:  "go/eac/core/tool",
			want:     `D:\work\repo\go\eac\core\tool`,
		},
		{
			name:     "Windows path preserves drive letter case",
			hostRoot: `c:\users\dev`,
			relPath:  "src",
			want:     `c:\users\dev\src`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinHostPath(tt.hostRoot, tt.relPath)
			if got != tt.want {
				t.Errorf("joinHostPath(%q, %q) = %q, want %q", tt.hostRoot, tt.relPath, got, tt.want)
			}
		})
	}
}

// TestFormatDockerVolume tests the formatDockerVolume helper function.
func TestFormatDockerVolume(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		// Windows paths converted to /c/path format
		{
			name: "Windows C drive",
			path: `C:\projects\eac`,
			want: "/c/projects/eac",
		},
		{
			name: "Windows D drive",
			path: `D:\work\repo\src`,
			want: "/d/work/repo/src",
		},
		{
			name: "Windows lowercase drive letter",
			path: `c:\users\dev`,
			want: "/c/users/dev",
		},
		{
			name: "Windows path with spaces",
			path: `C:\Program Files\app`,
			want: "/c/Program Files/app",
		},
		{
			name: "Windows root drive only",
			path: `C:\`,
			want: "/c/",
		},

		// Unix paths unchanged (only backslash to forward slash)
		{
			name: "Unix path unchanged",
			path: "/home/user/project",
			want: "/home/user/project",
		},
		{
			name: "Unix path with spaces",
			path: "/home/user/my project",
			want: "/home/user/my project",
		},

		// Paths with mixed separators (already partially unix-style)
		{
			name: "already unix-style path",
			path: "/c/projects/eac",
			want: "/c/projects/eac",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDockerVolume(tt.path)
			if got != tt.want {
				t.Errorf("formatDockerVolume(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// TestPopulateDinDContext tests environment variable reading for DinD context.
func TestPopulateDinDContext(t *testing.T) {
	tests := []struct {
		name                  string
		envHostRepoRoot       string
		envContainerRepoRoot  string
		wantHostWorkspaceRoot string
		wantContainerRepoRoot string
	}{
		{
			name:                  "no env vars - context unchanged",
			envHostRepoRoot:       "",
			envContainerRepoRoot:  "",
			wantHostWorkspaceRoot: "",
			wantContainerRepoRoot: "",
		},
		{
			name:                  "only host root set - defaults container root to /var/task",
			envHostRepoRoot:       "/home/user/project",
			envContainerRepoRoot:  "",
			wantHostWorkspaceRoot: "/home/user/project",
			wantContainerRepoRoot: DefaultContainerRoot,
		},
		{
			name:                  "both env vars set",
			envHostRepoRoot:       `C:\projects\eac`,
			envContainerRepoRoot:  "/workspace",
			wantHostWorkspaceRoot: `C:\projects\eac`,
			wantContainerRepoRoot: "/workspace",
		},
		{
			name:                  "host root set with custom container root",
			envHostRepoRoot:       "/opt/ci/workspace",
			envContainerRepoRoot:  "/app",
			wantHostWorkspaceRoot: "/opt/ci/workspace",
			wantContainerRepoRoot: "/app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			oldHost := os.Getenv(EnvHostRepoRoot)
			oldContainer := os.Getenv(EnvContainerRepoRoot)
			defer func() {
				if oldHost != "" {
					os.Setenv(EnvHostRepoRoot, oldHost)
				} else {
					os.Unsetenv(EnvHostRepoRoot)
				}
				if oldContainer != "" {
					os.Setenv(EnvContainerRepoRoot, oldContainer)
				} else {
					os.Unsetenv(EnvContainerRepoRoot)
				}
			}()

			// Set test environment
			if tt.envHostRepoRoot != "" {
				os.Setenv(EnvHostRepoRoot, tt.envHostRepoRoot)
			} else {
				os.Unsetenv(EnvHostRepoRoot)
			}
			if tt.envContainerRepoRoot != "" {
				os.Setenv(EnvContainerRepoRoot, tt.envContainerRepoRoot)
			} else {
				os.Unsetenv(EnvContainerRepoRoot)
			}

			// Create executor and context
			e := NewExecutor()
			defer e.Close()

			execCtx := &ExecutionContext{
				WorkspaceRoot: "/var/task", // Typical container workspace
			}

			// Call populateDinDContext
			e.populateDinDContext(execCtx)

			// Verify results
			if execCtx.HostWorkspaceRoot != tt.wantHostWorkspaceRoot {
				t.Errorf("HostWorkspaceRoot = %q, want %q", execCtx.HostWorkspaceRoot, tt.wantHostWorkspaceRoot)
			}
			if execCtx.ContainerRepoRoot != tt.wantContainerRepoRoot {
				t.Errorf("ContainerRepoRoot = %q, want %q", execCtx.ContainerRepoRoot, tt.wantContainerRepoRoot)
			}
		})
	}
}

// TestPopulateDinDContext_PreservesExistingFields verifies populateDinDContext
// does not overwrite existing fields when env vars are not set.
func TestPopulateDinDContext_PreservesExistingFields(t *testing.T) {
	// Save and restore environment
	oldHost := os.Getenv(EnvHostRepoRoot)
	defer func() {
		if oldHost != "" {
			os.Setenv(EnvHostRepoRoot, oldHost)
		} else {
			os.Unsetenv(EnvHostRepoRoot)
		}
	}()

	// Ensure env var is not set
	os.Unsetenv(EnvHostRepoRoot)

	e := NewExecutor()
	defer e.Close()

	execCtx := &ExecutionContext{
		WorkspaceRoot:     "/var/task",
		ModuleRoot:        "go/eac/core",
		OutputDir:         "/var/task/out",
		HostWorkspaceRoot: "", // Should remain empty
		ContainerRepoRoot: "", // Should remain empty
	}

	e.populateDinDContext(execCtx)

	// DinD fields should remain empty when env not set
	if execCtx.HostWorkspaceRoot != "" {
		t.Errorf("HostWorkspaceRoot should be empty, got %q", execCtx.HostWorkspaceRoot)
	}
	if execCtx.ContainerRepoRoot != "" {
		t.Errorf("ContainerRepoRoot should be empty, got %q", execCtx.ContainerRepoRoot)
	}

	// Other fields should be unchanged
	if execCtx.WorkspaceRoot != "/var/task" {
		t.Errorf("WorkspaceRoot changed unexpectedly: %q", execCtx.WorkspaceRoot)
	}
	if execCtx.ModuleRoot != "go/eac/core" {
		t.Errorf("ModuleRoot changed unexpectedly: %q", execCtx.ModuleRoot)
	}
}

// TestDinDConstants verifies the expected constant values.
func TestDinDConstants(t *testing.T) {
	if EnvHostRepoRoot != "R2R_HOST_REPOROOT" {
		t.Errorf("EnvHostRepoRoot = %q, want %q", EnvHostRepoRoot, "R2R_HOST_REPOROOT")
	}
	if EnvContainerRepoRoot != "R2R_CONTAINER_REPOROOT" {
		t.Errorf("EnvContainerRepoRoot = %q, want %q", EnvContainerRepoRoot, "R2R_CONTAINER_REPOROOT")
	}
	if DefaultContainerRoot != "/var/task" {
		t.Errorf("DefaultContainerRoot = %q, want %q", DefaultContainerRoot, "/var/task")
	}
}

// TestTranslatePathForMount_EdgeCases tests edge cases in path translation.
func TestTranslatePathForMount_EdgeCases(t *testing.T) {
	tests := []struct {
		name              string
		hostWorkspaceRoot string
		containerRepoRoot string
		containerPath     string
		want              string
	}{
		{
			name:              "empty container path",
			hostWorkspaceRoot: "/home/user/project",
			containerRepoRoot: "/var/task",
			containerPath:     "",
			want:              "",
		},
		{
			name:              "path is exactly container root with trailing slash",
			hostWorkspaceRoot: "/home/user/project",
			containerRepoRoot: "/var/task",
			containerPath:     "/var/task/",
			want:              "/home/user/project",
		},
		{
			name:              "container root with trailing slash in config",
			hostWorkspaceRoot: "/home/user/project",
			containerRepoRoot: "/var/task/",
			containerPath:     "/var/task/src",
			want:              "/home/user/project/src",
		},
		{
			name:              "Windows UNC-style path unchanged when not under root",
			hostWorkspaceRoot: `C:\projects\eac`,
			containerRepoRoot: "/var/task",
			containerPath:     "//server/share/file",
			want:              "//server/share/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ExecutionContext{
				HostWorkspaceRoot: tt.hostWorkspaceRoot,
				ContainerRepoRoot: tt.containerRepoRoot,
			}

			got := ctx.TranslatePathForMount(tt.containerPath)
			if got != tt.want {
				t.Errorf("TranslatePathForMount(%q) = %q, want %q", tt.containerPath, got, tt.want)
			}
		})
	}
}
