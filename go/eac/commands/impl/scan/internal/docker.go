// Package internal provides Docker utilities for security scanners using shared DockerClient.
package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ready-to-release/eac/go/eac/commands/internal/serve"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// containerCounter ensures unique container names even with parallel execution.
var containerCounter uint64

// Package-level logger for internal scanner operations.
var log = logging.C()

// OneOffDockerRunner wraps serve.DockerClient for running one-off container executions.
// It provides utilities for running security scanners and other short-lived containers
// with automatic cleanup after execution.
type OneOffDockerRunner struct {
	client serve.DockerClient
}

// NewOneOffDockerRunner creates a new one-off Docker runner.
func NewOneOffDockerRunner() (*OneOffDockerRunner, error) {
	client, err := serve.NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("Docker is not available. Please install and start Docker:\n"+
			"  Download from: https://www.docker.com/get-started\n"+
			"  Error: %w", err)
	}
	return &OneOffDockerRunner{
		client: client,
	}, nil
}

// NewOneOffDockerRunnerWithClient creates a runner with a custom Docker client (for testing).
func NewOneOffDockerRunnerWithClient(client serve.DockerClient) *OneOffDockerRunner {
	return &OneOffDockerRunner{
		client: client,
	}
}

// CheckAndPullImage checks if an image exists and pulls it if not.
func (r *OneOffDockerRunner) CheckAndPullImage(imageName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Debugf("Checking Docker image: %s", imageName)

	// Check if image exists locally
	images, err := r.client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	// Check if our image is in the list
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				log.Debugf("Docker image already exists locally: %s", imageName)
				return nil
			}
		}
	}

	// Image doesn't exist, pull it
	log.Debugf("Pulling Docker image (this may take a few minutes)... %s", imageName)
	reader, err := r.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Read pull output to completion (required for pull to finish)
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("failed to read pull output: %w", err)
	}

	log.Debugf("Docker image pulled successfully: %s", imageName)
	return nil
}

// RunContainer runs a one-off container, captures output, and ensures cleanup.
func (r *OneOffDockerRunner) RunContainer(containerConfig *container.Config, hostConfig *container.HostConfig) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Ensure AutoRemove is set so Docker cleans up container on exit
	if hostConfig == nil {
		hostConfig = &container.HostConfig{}
	}
	hostConfig.AutoRemove = true

	// Ensure container has stdout/stderr attached for capturing output
	containerConfig.AttachStdout = true
	containerConfig.AttachStderr = true

	// Create container with a unique name (timestamp + atomic counter for parallel safety)
	seq := atomic.AddUint64(&containerCounter, 1)
	containerName := fmt.Sprintf("security-scan-%d-%d", time.Now().UnixNano(), seq)
	log.Debugf("Creating container: name=%s image=%s autoRemove=true", containerName, containerConfig.Image)

	resp, err := r.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	containerID := resp.ID

	// Attach to container BEFORE starting to capture all output
	// This is required when using AutoRemove since container is deleted on exit
	log.Debugf("Attaching to container: id=%s", containerID)
	attachResp, err := r.client.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	// Start waiting BEFORE starting container to avoid missing exit event
	// This is the recommended pattern from Docker SDK documentation
	waitChan, errChan := r.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	log.Debugf("Starting container: id=%s", containerID)

	// Start container
	if err := r.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Read output synchronously - this blocks until container closes its output streams
	// No goroutine needed - StdCopy will return when the container exits and closes streams
	var stdout, stderr bytes.Buffer
	_, copyErr := stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader)

	log.Debugf("Container output read complete: stdoutSize=%d stderrSize=%d", stdout.Len(), stderr.Len())

	// Check container exit status
	select {
	case err := <-errChan:
		if err != nil {
			return nil, fmt.Errorf("container wait error: %w", err)
		}
	case waitResp := <-waitChan:
		log.Debugf("Container completed: id=%s exitCode=%d", containerID, waitResp.StatusCode)
		if waitResp.StatusCode != 0 {
			// Container failed - return stderr for debugging
			errMsg := stderr.String()
			if errMsg == "" {
				errMsg = "no error output available"
			}
			return nil, fmt.Errorf("container exited with code %d: %s", waitResp.StatusCode, errMsg)
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("container execution timed out after 30 minutes")
	}

	// Check if output reading failed
	if copyErr != nil {
		return nil, fmt.Errorf("failed to read container output: %w", copyErr)
	}

	// Prefer stdout, fallback to stderr for response
	output := stdout.Bytes()
	if len(output) == 0 && stderr.Len() > 0 {
		log.Debug("No stdout, using stderr output")
		output = stderr.Bytes()
	}

	return stripDockerLogHeaders(output), nil
}

// Close closes the Docker client connection.
func (r *OneOffDockerRunner) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// stripDockerLogHeaders removes Docker log stream headers and extracts JSON.
// Trivy and other scanners may prefix JSON output with log lines like:
// "2026-01-06T15:31:35Z	INFO	[vuln] Vulnerability scanning is enabled"
// This function finds the first valid JSON object/array and returns it.
func stripDockerLogHeaders(output []byte) []byte {
	if len(output) == 0 {
		return output
	}

	// Find first potential JSON start
	start := -1
	for i, b := range output {
		if b == '{' || b == '[' {
			start = i
			break
		}
	}

	if start == -1 {
		// No JSON found - return trimmed output (likely error message)
		return bytes.TrimSpace(output)
	}

	// Extract from potential start to end
	candidate := output[start:]

	// Try progressive substrings to find valid JSON
	// This handles cases where there's trailing text after JSON
	for end := len(candidate); end > 0; end-- {
		substring := candidate[:end]

		// Quick validation: must end with } or ]
		lastChar := substring[len(substring)-1]
		if lastChar != '}' && lastChar != ']' {
			continue
		}

		// Check if it's valid JSON by attempting to find matching braces
		if isBalancedJSON(substring) {
			return substring
		}
	}

	// Fallback: return everything from first brace
	return candidate
}

// isBalancedJSON checks if JSON braces/brackets are properly balanced
// This is a simple validation - not as thorough as json.Valid but faster.
func isBalancedJSON(data []byte) bool {
	depth := 0
	inString := false
	escape := false
	startChar := data[0]

	for _, b := range data {
		if escape {
			escape = false
			continue
		}

		if b == '\\' {
			escape = true
			continue
		}

		if b == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		switch b {
		case '{', '[':
			depth++
		case '}', ']':
			depth--
			if depth == 0 {
				// Found complete JSON
				return true
			}
			if depth < 0 {
				// Unbalanced - more closing than opening
				return false
			}
		}
	}

	// Must end with depth 0
	return depth == 0 && data[0] == startChar
}
