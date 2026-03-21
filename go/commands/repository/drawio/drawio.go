// Package drawio provides commands for editing DrawIO diagram files.
//
// DrawIO diagrams are stored as .drawio.png files - PNG images with embedded
// XML metadata. This package provides commands to extract, decode, encode,
// and embed the XML content, enabling LLM-powered diagram editing.
//
// Architecture:
//   - decode: Extract and decode XML from PNG to human-readable format
//   - encode: Encode human-readable XML back to DrawIO format
//   - embed: Write encoded XML into PNG file
//   - create: Create new .drawio.png with blank or provided content
//   - info: Show diagram metadata
//
// The actual XML manipulation is done by the drawio-oci Docker container.
// This package provides Go wrappers that handle Docker invocation.
package drawio

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	container "github.com/ready-to-release/eac/contracts/container-runtime/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/docker"
	dockerutil "github.com/ready-to-release/eac/go/adapters/docker/util"
	"github.com/ready-to-release/eac/go/core/iobuffer"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

const (
	// DefaultDrawioImageName is the default Docker image for drawio-oci
	// This is used as fallback when tool-config.yml is not loaded.
	DefaultDrawioImageName = "cli-drawio-oci:latest"

	// ContainerWorkdir is where files are mounted in the container
	ContainerWorkdir = "/docs"

	// MaxDockerOutputSize limits Docker command output to prevent memory exhaustion.
	MaxDockerOutputSize = 10 * 1024 * 1024 // 10MB
)

// ContainerProvider is a function that returns a ContainerPort.
type ContainerProvider func() container.ContainerPort

// GetDrawioImageName returns the Docker image for drawio.
// It first checks tool-config.yml, then falls back to the default.
func GetDrawioImageName() string {
	return tool.GetToolImageWithDefault("drawio", DefaultDrawioImageName)
}

var log = logging.C()

// EnsureDrawioImage builds the drawio-oci Docker image if needed.
// Uses Docker's layer cache for efficiency.
// The containerProvider parameter is optional (may be nil); when nil, falls back to docker.RunContainer.
func EnsureDrawioImage(workspaceRoot string, logWriter io.Writer, containerProvider ContainerProvider) error {
	// Get host repo root for Docker build context
	// In DinD mode, Docker daemon runs on the host, so it needs host paths
	hostRepoRoot, err := dockerutil.GetHostRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get host repo root: %w", err)
	}

	// Build paths for Dockerfile and context
	// These need to be host paths since Docker daemon runs on the host
	var dockerfilePath, contextPath string

	// Check if host path is Windows (has drive letter like C:)
	isWindowsHost := len(hostRepoRoot) >= 2 && hostRepoRoot[1] == ':'

	if isWindowsHost {
		// Windows host: use backslash separators
		dockerfilePath = hostRepoRoot + "\\containers\\drawio-oci\\Dockerfile"
		contextPath = hostRepoRoot + "\\containers\\drawio-oci"
	} else {
		// Unix host: use forward slash separators
		dockerfilePath = hostRepoRoot + "/containers/drawio-oci/Dockerfile"
		contextPath = hostRepoRoot + "/containers/drawio-oci"
	}

	imageName := GetDrawioImageName()

	// Try to use ContainerPort.Build if available
	if containerProvider != nil {
		c := containerProvider()
		if c != nil {
			if logWriter != nil {
				fmt.Fprintf(logWriter, "Building Docker image: %s\n", imageName)
			}
			config := &container.BuildConfig{
				ContextPath: contextPath,
				Dockerfile:  dockerfilePath,
				ImageTag:    imageName,
				Load:        true,
				LogWriter:   logWriter,
			}
			return c.Build(context.Background(), config)
		}
	}

	// Fallback: use EnsureServeImage which handles staleness checks + SDK-based builds
	output := io.Writer(io.Discard)
	if logWriter != nil {
		output = logWriter
	}

	serveConfig := &docker.ServeConfig{
		Image: imageName,
		BuildInfo: &docker.BuildInfo{
			Dockerfile:  dockerfilePath,
			ContextPath: contextPath,
		},
		Output: output,
	}

	return docker.EnsureServeImage(context.Background(), serveConfig)
}

// RunDrawioCommand executes a drawio-oci command in the container.
// The workspaceRoot is mounted at /docs in the container.
// The containerProvider parameter is optional (may be nil); when nil, falls back to docker.RunContainer.
func RunDrawioCommand(
	workspaceRoot string,
	args []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	containerProvider ContainerProvider,
) error {
	// Get host repo root for Docker volume mount
	// In DinD mode, Docker daemon runs on the host, so it needs host paths
	hostRepoRoot, err := dockerutil.GetHostRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get host repo root: %w", err)
	}
	if hostRepoRoot == "" {
		// Not in DinD mode - use the passed workspace root directly
		hostRepoRoot = workspaceRoot
	}

	// Try to use ContainerPort.Execute if available
	if containerProvider != nil {
		c := containerProvider()
		if c != nil && c.IsAvailable() {
			return runWithContainerPort(c, hostRepoRoot, args, stdout)
		}
	}

	// Fallback to exec.Command for backward compatibility
	return runWithExec(workspaceRoot, hostRepoRoot, args, stdin, stdout, stderr)
}

// runWithContainerPort executes the command using ContainerPort interface.
func runWithContainerPort(c container.ContainerPort, hostRepoRoot string, args []string, logWriter io.Writer) error {
	config := &container.ContainerConfig{
		Image:      GetDrawioImageName(),
		Command:    append([]string{"python", "/app/drawio_cli.py"}, args...),
		WorkingDir: ContainerWorkdir,
		Mounts: []container.MountConfig{
			{
				Source: hostRepoRoot,
				Target: ContainerWorkdir,
			},
		},
		LogWriter: logWriter,
	}

	// Add user spec in DinD mode to avoid permission issues
	if dockerutil.IsDinD() {
		uid := os.Getuid()
		gid := os.Getgid()
		config.User = fmt.Sprintf("%d:%d", uid, gid)
	}

	result, err := c.Execute(context.Background(), config)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("drawio command failed with exit code %d", result.ExitCode)
	}

	return nil
}

// runWithExec executes the command using docker.RunContainer (fallback when no ContainerPort).
func runWithExec(_ string, hostRepoRoot string, args []string, _ io.Reader, stdout, stderr io.Writer) error {
	command := append([]string{"python", "/app/drawio_cli.py"}, args...)

	runConfig := &docker.RunConfig{
		Image:      GetDrawioImageName(),
		Command:    command,
		WorkingDir: ContainerWorkdir,
		Mounts: []docker.MountConfig{
			{Source: hostRepoRoot, Target: ContainerWorkdir},
		},
		ContainerName: fmt.Sprintf("drawio-run-%d", time.Now().UnixNano()),
	}

	result, err := docker.RunContainer(context.Background(), runConfig)
	if err != nil {
		return err
	}

	// Write captured output to the provided writers
	if stdout != nil && result.Stdout != "" {
		_, _ = io.WriteString(stdout, result.Stdout)
	}
	if stderr != nil && result.Stderr != "" {
		_, _ = io.WriteString(stderr, result.Stderr)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("drawio command failed with exit code %d", result.ExitCode)
	}

	return nil
}

// TranslateToContainerPath converts a local path to a container path.
// Paths must be relative to workspaceRoot or absolute within it.
func TranslateToContainerPath(localPath, workspaceRoot string) (string, error) {
	// Make path absolute if relative
	if !filepath.IsAbs(localPath) {
		localPath = filepath.Join(workspaceRoot, localPath)
	}

	// Get relative path from workspace root
	relPath, err := filepath.Rel(workspaceRoot, localPath)
	if err != nil {
		return "", fmt.Errorf("path not within workspace: %w", err)
	}

	// Check for path traversal
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("path outside workspace: %s", localPath)
	}

	// Convert to container path (forward slashes)
	containerPath := ContainerWorkdir + "/" + strings.ReplaceAll(relPath, "\\", "/")
	return containerPath, nil
}

// GetRepoRoot returns the repository root, handling errors.
func GetRepoRoot() (string, error) {
	return repository.GetRepositoryRoot("")
}

// RunDrawioCommandWithOutput runs a command and returns stdout as string.
// Uses a limited buffer to prevent memory exhaustion from runaway Docker output.
// The containerProvider parameter is optional (may be nil); when nil, falls back to docker.RunContainer.
func RunDrawioCommandWithOutput(workspaceRoot string, args []string, containerProvider ContainerProvider) (string, error) {
	stdout := iobuffer.NewLimitedBuffer(MaxDockerOutputSize)
	stderr := iobuffer.NewLimitedBuffer(MaxDockerOutputSize)
	err := RunDrawioCommand(workspaceRoot, args, nil, stdout, stderr, containerProvider)
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

// CheckDockerAvailable verifies Docker is available and the image exists.
// The containerProvider parameter is optional (may be nil); when nil, falls back to docker.RunContainer.
func CheckDockerAvailable(workspaceRoot string, containerProvider ContainerProvider) error {
	// Check container runtime availability
	if containerProvider != nil {
		c := containerProvider()
		if c != nil {
			if !c.IsAvailable() {
				return fmt.Errorf("container runtime not available. Ensure Docker is installed and running")
			}

			// Check if image exists
			imageName := GetDrawioImageName()
			if !c.ImageExists(context.Background(), imageName) {
				// Image doesn't exist, build it
				log.Infof("Building drawio-oci Docker image...")
				if err := EnsureDrawioImage(workspaceRoot, os.Stderr, containerProvider); err != nil {
					return fmt.Errorf("failed to build drawio-oci image: %w", err)
				}
			}
			return nil
		}
	}

	// Fallback: use dockerutil for availability check, EnsureDrawioImage for build
	if !dockerutil.IsDockerAvailable() {
		return fmt.Errorf("Docker is not available. Ensure Docker is installed and running")
	}

	// EnsureDrawioImage handles staleness checks and only builds if needed
	return EnsureDrawioImage(workspaceRoot, nil, containerProvider)
}
