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
// The actual XML manipulation is done by the drawio-tool Docker container.
// This package provides Go wrappers that handle Docker invocation.
package drawio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	container "github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces"
	dockerutil "github.com/ready-to-release/eac/go/adapters/docker/util"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

const (
	// DefaultDrawioImageName is the default Docker image for drawio-tool
	// This is used as fallback when tool-config.yml is not loaded.
	DefaultDrawioImageName = "cli-drawio-tool:latest"

	// ContainerWorkdir is where files are mounted in the container
	ContainerWorkdir = "/docs"
)

// ContainerProvider is a function that returns a ContainerPort.
type ContainerProvider func() container.ContainerPort

// defaultContainerProvider is set by the application to provide the container runtime.
var defaultContainerProvider ContainerProvider

// SetContainerProvider sets the container provider for drawio operations.
func SetContainerProvider(provider ContainerProvider) {
	defaultContainerProvider = provider
}

// GetDrawioImageName returns the Docker image for drawio.
// It first checks tool-config.yml, then falls back to the default.
func GetDrawioImageName() string {
	return tool.GetToolImageWithDefault("drawio", DefaultDrawioImageName)
}

var log = logging.C()

// EnsureDrawioImage builds the drawio-tool Docker image if needed.
// Uses Docker's layer cache for efficiency.
func EnsureDrawioImage(workspaceRoot string, logWriter io.Writer) error {
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
		dockerfilePath = hostRepoRoot + "\\containers\\drawio-tool\\Dockerfile"
		contextPath = hostRepoRoot + "\\containers\\drawio-tool"
	} else {
		// Unix host: use forward slash separators
		dockerfilePath = hostRepoRoot + "/containers/drawio-tool/Dockerfile"
		contextPath = hostRepoRoot + "/containers/drawio-tool"
	}

	imageName := GetDrawioImageName()
	if logWriter != nil {
		fmt.Fprintf(logWriter, "Building Docker image: %s\n", imageName)
	}

	// Try to use ContainerPort.Build if available
	if defaultContainerProvider != nil {
		c := defaultContainerProvider()
		if c != nil {
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

	// Fallback to exec.Command
	cmd := exec.Command("docker", "build",
		"-t", imageName,
		"-f", dockerfilePath,
		contextPath)

	if logWriter != nil {
		cmd.Stdout = logWriter
		cmd.Stderr = logWriter
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	return nil
}

// RunDrawioCommand executes a drawio-tool command in the container.
// The workspaceRoot is mounted at /docs in the container.
func RunDrawioCommand(
	workspaceRoot string,
	args []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
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
	if defaultContainerProvider != nil {
		c := defaultContainerProvider()
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

// runWithExec executes the command using exec.Command (fallback).
func runWithExec(workspaceRoot, hostRepoRoot string, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	// Format Docker volume path (handles Windows C:\ to /c/ conversion)
	dockerVolume := dockerutil.FormatDockerVolume(hostRepoRoot)

	// Build Docker command
	dockerArgs := []string{
		"run", "--rm",
		"-v", dockerVolume + ":" + ContainerWorkdir,
		"-w", ContainerWorkdir,
	}

	// Add user spec in DinD mode to avoid permission issues
	if dockerutil.IsDinD() {
		uid := os.Getuid()
		gid := os.Getgid()
		userSpec := fmt.Sprintf("%d:%d", uid, gid)
		dockerArgs = append(dockerArgs, "--user", userSpec)
	}

	// Handle stdin
	if stdin != nil {
		dockerArgs = append(dockerArgs, "-i")
	}

	// Add image and command
	dockerArgs = append(dockerArgs, GetDrawioImageName(), "python", "/app/drawio_cli.py")
	dockerArgs = append(dockerArgs, args...)

	cmd := exec.Command("docker", dockerArgs...)
	cmd.Dir = workspaceRoot
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd.Run()
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
func RunDrawioCommandWithOutput(workspaceRoot string, args []string) (string, error) {
	var stdout, stderr bytes.Buffer
	err := RunDrawioCommand(workspaceRoot, args, nil, &stdout, &stderr)
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

// CheckDockerAvailable verifies Docker is available and the image exists.
func CheckDockerAvailable(workspaceRoot string) error {
	// Check container runtime availability
	if defaultContainerProvider != nil {
		c := defaultContainerProvider()
		if c != nil {
			if !c.IsAvailable() {
				return fmt.Errorf("container runtime not available. Ensure Docker is installed and running")
			}

			// Check if image exists
			imageName := GetDrawioImageName()
			if !c.ImageExists(context.Background(), imageName) {
				// Image doesn't exist, build it
				log.Infof("Building drawio-tool Docker image...")
				if err := EnsureDrawioImage(workspaceRoot, os.Stderr); err != nil {
					return fmt.Errorf("failed to build drawio-tool image: %w", err)
				}
			}
			return nil
		}
	}

	// Fallback: use dockerutil
	if !dockerutil.IsDockerAvailable() {
		return fmt.Errorf("Docker is not available. Ensure Docker is installed and running")
	}

	// Check if image exists, build if not
	cmd := exec.Command("docker", "image", "inspect", GetDrawioImageName())
	if err := cmd.Run(); err != nil {
		// Image doesn't exist, build it
		log.Infof("Building drawio-tool Docker image...")
		if err := EnsureDrawioImage(workspaceRoot, os.Stderr); err != nil {
			return fmt.Errorf("failed to build drawio-tool image: %w", err)
		}
	}

	return nil
}
