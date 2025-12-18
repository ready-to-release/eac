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

// containerCounter ensures unique container names even with parallel execution
var containerCounter uint64

// Package-level logger for internal scanner operations
var log = logging.C()

// OneOffDockerRunner wraps serve.DockerClient for running one-off container executions.
// It provides utilities for running security scanners and other short-lived containers
// with automatic cleanup after execution.
type OneOffDockerRunner struct {
	client serve.DockerClient
}

// NewOneOffDockerRunner creates a new one-off Docker runner
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

// NewOneOffDockerRunnerWithClient creates a runner with a custom Docker client (for testing)
func NewOneOffDockerRunnerWithClient(client serve.DockerClient) *OneOffDockerRunner {
	return &OneOffDockerRunner{
		client: client,
	}
}

// CheckAndPullImage checks if an image exists and pulls it if not
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

// RunContainer runs a one-off container, captures output, and ensures cleanup
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

	log.Debugf("Starting container: id=%s", containerID)

	// Start container
	if err := r.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Read output from attached stream (demultiplex stdout/stderr)
	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader); err != nil {
		return nil, fmt.Errorf("failed to read container output: %w", err)
	}

	// Wait for container to complete
	waitChan, errChan := r.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	select {
	case err := <-errChan:
		if err != nil {
			// Container may already be removed by AutoRemove, which is fine
			log.Debugf("Container wait: %v (may be auto-removed)", err)
		}
	case waitResp := <-waitChan:
		log.Debugf("Container completed: id=%s exitCode=%d", containerID, waitResp.StatusCode)
	case <-ctx.Done():
		return nil, fmt.Errorf("container execution timed out")
	}

	log.Debugf("Container output captured: stdoutSize=%d stderrSize=%d", stdout.Len(), stderr.Len())

	// Combine stdout and stderr (stderr usually has diagnostic info, stdout has JSON)
	// Prefer stdout if it has content, otherwise use stderr
	output := stdout.Bytes()
	if len(output) == 0 && stderr.Len() > 0 {
		log.Debug("No stdout, using stderr output")
		output = stderr.Bytes()
	}

	// If we have both stdout and stderr, log stderr for diagnostics
	if stdout.Len() > 0 && stderr.Len() > 0 {
		log.Debugf("Container stderr: %s", string(stderr.Bytes()))
	}

	return output, nil
}

// gracefulCleanup stops and removes a container gracefully
func (r *OneOffDockerRunner) gracefulCleanup(ctx context.Context, containerID, containerName string) {
	log.Debugf("Cleaning up container: id=%s", containerID)

	// Stop container with timeout
	timeout := 10
	if err := r.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		log.Debugf("Container stop completed or already stopped: id=%s err=%v", containerID, err)
	}

	// Remove container
	if err := r.client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
		log.Warnf("Failed to remove container: id=%s err=%v", containerID, err)
	} else {
		log.Debugf("Container cleaned up successfully: name=%s", containerName)
	}
}

// Close closes the Docker client connection
func (r *OneOffDockerRunner) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// stripDockerLogHeaders removes Docker log stream headers from output and extracts JSON.
// This function finds the first JSON object or array and extracts just the JSON portion,
// handling cases where there's text before or after the JSON.
func stripDockerLogHeaders(output []byte) []byte {
	if len(output) == 0 {
		return output
	}

	// Find first JSON character
	start := -1
	for i, b := range output {
		if b == '{' || b == '[' {
			start = i
			break
		}
	}

	if start == -1 {
		// No JSON found, return as-is (might be error message)
		return bytes.TrimSpace(output)
	}

	// Try to find the end of the JSON object/array
	// Simple approach: find matching closing brace/bracket
	depth := 0
	inString := false
	escape := false
	startChar := output[start]
	endChar := byte('}')
	if startChar == '[' {
		endChar = ']'
	}

	for i := start; i < len(output); i++ {
		b := output[i]

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

		if b == startChar {
			depth++
		} else if b == endChar {
			depth--
			if depth == 0 {
				// Found the end of JSON
				return output[start : i+1]
			}
		}
	}

	// If we couldn't find proper end, return from start to end
	return output[start:]
}
