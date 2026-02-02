// Package docker provides ZAP scanner functionality using Docker containers.
package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ready-to-release/eac/go/adapters/docker/util"
)

// ZAP scan types.
const (
	ZAPBaseline = "baseline"
	ZAPFull     = "full"
	ZAPAPI      = "api"
)

// scanContainerCounter ensures unique container names for concurrent scans.
var scanContainerCounter uint64

// RunZAPScan executes OWASP ZAP dynamic security scan via Docker.
func RunZAPScan(targetURL, scanType, workspaceRoot, zapImage string) (interface{}, error) {
	// Check for mock mode via environment variable
	if os.Getenv("R2R_MOCK_SECURITY") != "" || os.Getenv("R2R_MOCK_DOCKER") == "true" {
		return getDefaultMockZAPOutput(), nil
	}

	// Create Docker client
	client, err := NewDockerClient()
	if err != nil {
		return nil, fmt.Errorf("Docker is not available. Please install and start Docker:\n"+
			"  Download from: https://www.docker.com/get-started\n"+
			"  Error: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Check and pull image if needed
	if err := checkAndPullImage(ctx, client, zapImage); err != nil {
		return nil, err
	}

	// Create temporary output directory for ZAP report
	outputDir := filepath.Join(workspaceRoot, "out", "scan", "zap-temp")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create ZAP output directory: %w", err)
	}
	defer os.RemoveAll(outputDir) // Clean up temp files

	// Get absolute path for volume mount
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Translate path for Docker-in-Docker environments
	hostPath, err := util.TranslatePathForMount(absOutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to translate path for mount: %w", err)
	}

	// Convert to Docker-compatible path format for volume mount
	dockerOutputDir := util.FormatDockerVolume(hostPath)

	// Build ZAP scan command based on scan type
	var zapScript string
	switch scanType {
	case ZAPBaseline:
		zapScript = "zap-baseline.py"
	case ZAPFull:
		zapScript = "zap-full-scan.py"
	case ZAPAPI:
		zapScript = "zap-api-scan.py"
	default:
		return nil, fmt.Errorf("invalid scan type: %s (must be baseline, full, or api)", scanType)
	}

	// Create container with unique name
	seq := atomic.AddUint64(&scanContainerCounter, 1)
	containerName := fmt.Sprintf("zap-scan-%d-%d", time.Now().UnixNano(), seq)

	containerConfig := &container.Config{
		Image: zapImage,
		Cmd: []string{
			zapScript,
			"-t", targetURL,
			"-J", "zap-report.json", // JSON output
			"-I", // Don't fail on warnings
		},
		AttachStdout: true,
		AttachStderr: true,
	}

	hostConfig := &container.HostConfig{
		Binds:      []string{fmt.Sprintf("%s:/zap/wrk:rw", dockerOutputDir)},
		AutoRemove: true,
	}

	// Create container
	resp, err := client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZAP container: %w", err)
	}
	containerID := resp.ID

	// Attach to container before starting
	attachResp, err := client.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to ZAP container: %w", err)
	}
	defer attachResp.Close()

	// Start waiting before starting container
	waitChan, errChan := client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	// Start container
	if err := client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start ZAP container: %w", err)
	}

	// Read output
	var stdout, stderr bytes.Buffer
	_, copyErr := stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader)

	// Check container exit status
	select {
	case err := <-errChan:
		if err != nil {
			return nil, fmt.Errorf("ZAP container wait error: %w", err)
		}
	case waitResp := <-waitChan:
		// ZAP may return non-zero exit code even with successful scan
		// Check if report file was created
		reportPath := filepath.Join(absOutputDir, "zap-report.json")
		if _, statErr := os.Stat(reportPath); statErr != nil {
			if waitResp.StatusCode != 0 {
				errMsg := stderr.String()
				if errMsg == "" {
					errMsg = stdout.String()
				}
				if errMsg == "" {
					errMsg = "no error output available"
				}
				return nil, fmt.Errorf("ZAP scan failed with exit code %d: %s", waitResp.StatusCode, errMsg)
			}
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("ZAP scan timed out after 30 minutes")
	}

	// Check if output reading failed
	if copyErr != nil {
		return nil, fmt.Errorf("failed to read ZAP container output: %w", copyErr)
	}

	// Read and parse ZAP report
	reportPath := filepath.Join(absOutputDir, "zap-report.json")
	reportData, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ZAP report: %w", err)
	}

	var findings interface{}
	if err := json.Unmarshal(reportData, &findings); err != nil {
		return nil, fmt.Errorf("failed to parse ZAP report: %w", err)
	}

	return findings, nil
}

// checkAndPullImage checks if an image exists and pulls it if not.
func checkAndPullImage(ctx context.Context, client DockerClient, imageName string) error {
	// Check if image exists locally
	images, err := client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	// Check if our image is in the list
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				return nil
			}
		}
	}

	// Image doesn't exist, pull it
	reader, err := client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Read pull output to completion
	buf := make([]byte, 1024)
	for {
		_, err := reader.Read(buf)
		if err != nil {
			break
		}
	}

	return nil
}

// getDefaultMockZAPOutput returns default mock data for ZAP.
func getDefaultMockZAPOutput() map[string]interface{} {
	return map[string]interface{}{
		"site": []map[string]interface{}{
			{
				"@name":  "http://localhost:8080",
				"alerts": []interface{}{},
			},
		},
	}
}
