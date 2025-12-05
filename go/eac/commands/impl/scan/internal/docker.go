// Package internal provides Docker utilities for security scanners using shared DockerClient.
package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ready-to-release/eac/go/eac/commands/internal/serve"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"go.uber.org/zap"
)

// OneOffDockerRunner wraps serve.DockerClient for running one-off container executions.
// It provides utilities for running security scanners and other short-lived containers
// with automatic cleanup after execution.
type OneOffDockerRunner struct {
	client serve.DockerClient
	logger *logging.Logger
}

// NewOneOffDockerRunner creates a new one-off Docker runner
func NewOneOffDockerRunner(logger *logging.Logger) (*OneOffDockerRunner, error) {
	client, err := serve.NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("Docker is not available. Please install and start Docker:\n"+
			"  Download from: https://www.docker.com/get-started\n"+
			"  Error: %w", err)
	}
	return &OneOffDockerRunner{
		client: client,
		logger: logger,
	}, nil
}

// NewOneOffDockerRunnerWithClient creates a runner with a custom Docker client (for testing)
func NewOneOffDockerRunnerWithClient(client serve.DockerClient, logger *logging.Logger) *OneOffDockerRunner {
	return &OneOffDockerRunner{
		client: client,
		logger: logger,
	}
}

// CheckAndPullImage checks if an image exists and pulls it if not
func (r *OneOffDockerRunner) CheckAndPullImage(imageName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	r.logger.Info("Checking Docker image", zap.String("image", imageName))

	// Check if image exists locally
	images, err := r.client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	// Check if our image is in the list
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				r.logger.Debug("Docker image already exists locally", zap.String("image", imageName))
				return nil
			}
		}
	}

	// Image doesn't exist, pull it
	r.logger.Info("Pulling Docker image (this may take a few minutes)...", zap.String("image", imageName))
	reader, err := r.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Read pull output to completion (required for pull to finish)
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("failed to read pull output: %w", err)
	}

	r.logger.Info("Docker image pulled successfully", zap.String("image", imageName))
	return nil
}

// RunContainer runs a one-off container, captures output, and ensures cleanup
func (r *OneOffDockerRunner) RunContainer(containerConfig *container.Config, hostConfig *container.HostConfig) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Create container with a unique name
	containerName := fmt.Sprintf("security-scan-%d", time.Now().UnixNano())
	r.logger.Debug("Creating container",
		zap.String("name", containerName),
		zap.String("image", containerConfig.Image))

	resp, err := r.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	containerID := resp.ID

	// Ensure container is cleaned up
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		r.gracefulCleanup(cleanupCtx, containerID, containerName)
	}()

	r.logger.Debug("Starting container", zap.String("id", containerID))

	// Start container
	if err := r.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to complete
	waitChan, errChan := r.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	select {
	case err := <-errChan:
		if err != nil {
			return nil, fmt.Errorf("error waiting for container: %w", err)
		}
	case waitResp := <-waitChan:
		r.logger.Debug("Container completed",
			zap.String("id", containerID),
			zap.Int64("exitCode", waitResp.StatusCode))
	case <-ctx.Done():
		return nil, fmt.Errorf("container execution timed out")
	}

	// Get container logs
	r.logger.Debug("Retrieving container logs", zap.String("id", containerID))
	logsReader, err := r.client.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logsReader.Close()

	// Demultiplex Docker logs (stdout and stderr are multiplexed with 8-byte headers)
	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, logsReader); err != nil {
		return nil, fmt.Errorf("failed to demultiplex container logs: %w", err)
	}

	r.logger.Debug("Container output captured",
		zap.Int("stdoutSize", stdout.Len()),
		zap.Int("stderrSize", stderr.Len()))

	// Combine stdout and stderr (stderr usually has diagnostic info, stdout has JSON)
	// Prefer stdout if it has content, otherwise use stderr
	output := stdout.Bytes()
	if len(output) == 0 && stderr.Len() > 0 {
		r.logger.Debug("No stdout, using stderr output")
		output = stderr.Bytes()
	}

	// If we have both stdout and stderr, log stderr for diagnostics
	if stdout.Len() > 0 && stderr.Len() > 0 {
		r.logger.Debug("Container stderr", zap.String("stderr", string(stderr.Bytes())))
	}

	return output, nil
}

// gracefulCleanup stops and removes a container gracefully
func (r *OneOffDockerRunner) gracefulCleanup(ctx context.Context, containerID, containerName string) {
	r.logger.Debug("Cleaning up container", zap.String("id", containerID))

	// Stop container with timeout
	timeout := 10
	if err := r.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		r.logger.Debug("Container stop completed or already stopped",
			zap.String("id", containerID),
			zap.Error(err))
	}

	// Remove container
	if err := r.client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
		r.logger.Warn("Failed to remove container",
			zap.String("id", containerID),
			zap.Error(err))
	} else {
		r.logger.Debug("Container cleaned up successfully", zap.String("name", containerName))
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
