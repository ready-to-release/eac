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
// The actual XML manipulation is done by the drawio-cli Docker container.
// This package provides Go wrappers that handle Docker invocation.
package drawio

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	dockerutil "github.com/ready-to-release/eac/go/eac/adapters/docker/util"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

const (
	// DefaultDrawioImageName is the default Docker image for drawio-cli
	// This is used as fallback when tool-config.yml is not loaded.
	DefaultDrawioImageName = "cli-drawio-cli:latest"

	// ContainerWorkdir is where files are mounted in the container
	ContainerWorkdir = "/docs"
)

// GetDrawioImageName returns the Docker image for drawio.
// It first checks tool-config.yml, then falls back to the default.
func GetDrawioImageName() string {
	return tool.GetToolImageWithDefault("drawio", DefaultDrawioImageName)
}

var log = logging.C()

// EnsureDrawioImage builds the drawio-cli Docker image if needed.
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
		dockerfilePath = hostRepoRoot + "\\containers\\drawio-cli\\Dockerfile"
		contextPath = hostRepoRoot + "\\containers\\drawio-cli"
	} else {
		// Unix host: use forward slash separators
		dockerfilePath = hostRepoRoot + "/containers/drawio-cli/Dockerfile"
		contextPath = hostRepoRoot + "/containers/drawio-cli"
	}

	imageName := GetDrawioImageName()
	if logWriter != nil {
		fmt.Fprintf(logWriter, "Building Docker image: %s\n", imageName)
	}

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

// RunDrawioCommand executes a drawio-cli command in the container.
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
	if !dockerutil.IsDockerAvailable() {
		return fmt.Errorf("Docker is not available. Ensure Docker is installed and running")
	}

	// Check if image exists, build if not
	cmd := exec.Command("docker", "image", "inspect", GetDrawioImageName())
	if err := cmd.Run(); err != nil {
		// Image doesn't exist, build it
		log.Infof("Building drawio-cli Docker image...")
		if err := EnsureDrawioImage(workspaceRoot, os.Stderr); err != nil {
			return fmt.Errorf("failed to build drawio-cli image: %w", err)
		}
	}

	return nil
}
