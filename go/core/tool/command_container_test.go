package tool

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildCommand_ContainerTool verifies BuildCommand creates proper docker run commands.
func TestBuildCommand_ContainerTool(t *testing.T) {
	td := &ToolDefinition{
		ID:    "test-tool",
		Type:  ToolTypeContainer,
		Image: "golang",
		Tag:   "1.21",
		Command: []string{"go", "test", "./..."},
		WorkDir: "/workspace",
		Mounts: []MountConfig{
			{Source: "/host/path", Target: "/container/path"},
		},
		Env: map[string]string{
			"GO111MODULE": "on",
		},
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: "/workspace/root",
		Placeholders:  map[string]string{},
	}

	cmd := BuildCommand(context.Background(), td, execCtx)
	require.NotNil(t, cmd)

	// Verify docker command (path may be resolved to full path on Windows)
	assert.Contains(t, cmd.Path, "docker")
	args := cmd.Args[1:] // Skip "docker" itself

	// Check run and rm flags
	assert.Contains(t, args, "run")
	assert.Contains(t, args, "--rm")
	assert.Contains(t, args, "-i")

	// Check workdir
	workdirIdx := indexOf(args, "-w")
	require.NotEqual(t, -1, workdirIdx)
	assert.Equal(t, "/workspace", args[workdirIdx+1])

	// Check image reference
	assert.Contains(t, args, "golang:1.21")

	// Check command is at the end
	lastArgs := args[len(args)-3:]
	assert.Equal(t, []string{"go", "test", "./..."}, lastArgs)
}

// TestBuildCommand_ContainerWithMounts verifies mount formatting.
func TestBuildCommand_ContainerWithMounts(t *testing.T) {
	td := &ToolDefinition{
		ID:    "test-tool",
		Type:  ToolTypeContainer,
		Image: "alpine",
		Tag:   "3.18",
		Mounts: []MountConfig{
			// Use absolute paths to avoid workspace prefix join
			{Source: "/host/src", Target: "/container/src", ReadOnly: false},
			{Source: "/host/cache", Target: "/container/cache", ReadOnly: true},
		},
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: "/workspace",
	}

	cmd := BuildCommand(context.Background(), td, execCtx)
	require.NotNil(t, cmd)

	args := strings.Join(cmd.Args, " ")

	// Check mounts - sources are absolute so should be preserved as-is
	assert.Contains(t, args, "/host/src:/container/src")
	assert.Contains(t, args, "/host/cache:/container/cache:ro")
}

// TestBuildCommand_ContainerWithEnv verifies environment variable handling.
func TestBuildCommand_ContainerWithEnv(t *testing.T) {
	td := &ToolDefinition{
		ID:    "test-tool",
		Type:  ToolTypeContainer,
		Image: "node",
		Tag:   "18",
		Env: map[string]string{
			"NODE_ENV": "production",
		},
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: "/workspace",
		EnvOverrides: map[string]string{
			"CI": "true",
		},
	}

	cmd := BuildCommand(context.Background(), td, execCtx)
	require.NotNil(t, cmd)

	args := strings.Join(cmd.Args, " ")

	// Check env vars
	assert.Contains(t, args, "-e NODE_ENV=production")
	assert.Contains(t, args, "-e CI=true")
}

// TestBuildCommand_ContainerWithResources verifies resource limit handling.
func TestBuildCommand_ContainerWithResources(t *testing.T) {
	td := &ToolDefinition{
		ID:    "test-tool",
		Type:  ToolTypeContainer,
		Image: "golang",
		Tag:   "1.21",
		Resources: &ToolResources{
			CPUs:   2,
			Memory: "4g",
		},
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: "/workspace",
	}

	cmd := BuildCommand(context.Background(), td, execCtx)
	require.NotNil(t, cmd)

	args := strings.Join(cmd.Args, " ")

	// Check resource limits
	assert.Contains(t, args, "--memory 4g")
	assert.Contains(t, args, "--cpus 2.00")
}

// TestBuildCommand_ContainerWithNetwork verifies network mode handling.
func TestBuildCommand_ContainerWithNetwork(t *testing.T) {
	td := &ToolDefinition{
		ID:      "test-tool",
		Type:    ToolTypeContainer,
		Image:   "nginx",
		Tag:     "1.25",
		Network: "host",
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: "/workspace",
	}

	cmd := BuildCommand(context.Background(), td, execCtx)
	require.NotNil(t, cmd)

	args := cmd.Args
	networkIdx := indexOf(args, "--network")
	require.NotEqual(t, -1, networkIdx)
	assert.Equal(t, "host", args[networkIdx+1])
}

// TestBuildCommand_ContainerWithPrivileged verifies privileged mode.
func TestBuildCommand_ContainerWithPrivileged(t *testing.T) {
	td := &ToolDefinition{
		ID:         "test-tool",
		Type:       ToolTypeContainer,
		Image:      "docker",
		Tag:        "dind",
		Privileged: true,
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: "/workspace",
	}

	cmd := BuildCommand(context.Background(), td, execCtx)
	require.NotNil(t, cmd)

	assert.Contains(t, cmd.Args, "--privileged")
}

// TestBuildCommand_ContainerLocalPath verifies local container image tags.
func TestBuildCommand_ContainerLocalPath(t *testing.T) {
	td := &ToolDefinition{
		ID:        "test-tool",
		Type:      ToolTypeContainer,
		LocalPath: "containers/my-tool",
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: "/workspace",
	}

	cmd := BuildCommand(context.Background(), td, execCtx)
	require.NotNil(t, cmd)

	// Should use local image tag (directory name + :local)
	assert.Contains(t, cmd.Args, "my-tool:local")
}

// TestBuildCommand_ContainerWithArgsOverrides verifies args override handling.
func TestBuildCommand_ContainerWithArgsOverrides(t *testing.T) {
	td := &ToolDefinition{
		ID:      "test-tool",
		Type:    ToolTypeContainer,
		Image:   "golang",
		Tag:     "1.21",
		Command: []string{"go", "test"},
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: "/workspace",
		ArgsOverrides: []string{"-v", "-count=1"},
	}

	cmd := BuildCommand(context.Background(), td, execCtx)
	require.NotNil(t, cmd)

	// ArgsOverrides should be appended after command
	lastArgs := cmd.Args[len(cmd.Args)-4:]
	assert.Equal(t, []string{"go", "test", "-v", "-count=1"}, lastArgs)
}

// TestBuildCommand_NilInputs verifies nil handling.
func TestBuildCommand_NilInputs(t *testing.T) {
	td := &ToolDefinition{
		ID:    "test-tool",
		Type:  ToolTypeContainer,
		Image: "alpine",
		Tag:   "latest",
	}
	execCtx := &ExecutionContext{}

	// nil tool
	assert.Nil(t, BuildCommand(context.Background(), nil, execCtx))

	// nil execCtx
	assert.Nil(t, BuildCommand(context.Background(), td, nil))

	// both nil
	assert.Nil(t, BuildCommand(context.Background(), nil, nil))
}

// TestBuildCommand_NoImage verifies missing image returns nil.
func TestBuildCommand_NoImage(t *testing.T) {
	td := &ToolDefinition{
		ID:   "test-tool",
		Type: ToolTypeContainer,
		// No Image, Tag, or LocalPath
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: "/workspace",
	}

	cmd := BuildCommand(context.Background(), td, execCtx)
	assert.Nil(t, cmd, "should return nil when no image reference is available")
}

// TestBuildCommand_SystemToolUnchanged verifies system tool behavior is preserved.
func TestBuildCommand_SystemToolUnchanged(t *testing.T) {
	td := &ToolDefinition{
		ID:     "go",
		Type:   ToolTypeSystem,
		Binary: "go",
		Args:   []string{"build"},
	}

	execCtx := &ExecutionContext{
		WorkspaceRoot: "/workspace",
		ModuleRoot:    "/workspace/module",
	}

	cmd := BuildCommand(context.Background(), td, execCtx)
	require.NotNil(t, cmd)

	// Should be direct go command, not docker
	assert.True(t, strings.HasSuffix(cmd.Path, "go") || strings.HasSuffix(cmd.Path, "go.exe"))
	assert.Equal(t, []string{"go", "build"}, cmd.Args)
}

// indexOf finds the index of a string in a slice, or -1 if not found.
func indexOf(slice []string, target string) int {
	for i, s := range slice {
		if s == target {
			return i
		}
	}
	return -1
}
