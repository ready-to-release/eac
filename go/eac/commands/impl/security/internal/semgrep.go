// Package internal provides Semgrep security scanner integration via Docker.
package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"go.uber.org/zap"
)

const (
	// SemgrepImage is the official Semgrep Docker image
	SemgrepImage = "semgrep/semgrep:latest"
)

// Mock support for testing
var mockSemgrepOutput interface{}

// SetMockSemgrepOutput sets a mock Semgrep response for testing
func SetMockSemgrepOutput(output interface{}) {
	mockSemgrepOutput = output
}

// ResetMockSemgrepOutput clears the mock Semgrep response
func ResetMockSemgrepOutput() {
	mockSemgrepOutput = nil
}

// getDefaultMockSemgrepOutput returns default mock data for Semgrep
func getDefaultMockSemgrepOutput() map[string]interface{} {
	return map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"check_id": "test-rule",
				"path":     "test.go",
			},
		},
	}
}

// RunSemgrepSAST executes Semgrep static analysis via Docker
func RunSemgrepSAST(workspaceRoot, moduleRoot string, config string, logger *logging.Logger) (interface{}, error) {
	// Check for mock output (testing only)
	// Priority: in-process mock > environment variable mock
	if mockSemgrepOutput != nil {
		logger.Debug("Using mocked Semgrep output (in-process)")
		return mockSemgrepOutput, nil
	}
	if os.Getenv("R2R_MOCK_SECURITY") != "" {
		logger.Debug("Using mocked Semgrep output (environment)")
		return getDefaultMockSemgrepOutput(), nil
	}

	// Create Docker runner
	dockerRunner, err := NewOneOffDockerRunner(logger)
	if err != nil {
		return nil, err
	}
	defer dockerRunner.Close()

	// Pull image if needed
	if err := dockerRunner.CheckAndPullImage(SemgrepImage); err != nil {
		return nil, err
	}

	logger.Info("Running Semgrep SAST scanner via Docker",
		zap.String("moduleRoot", moduleRoot),
		zap.String("config", config))

	// Resolve module root relative to workspace root
	absModuleRoot := filepath.Join(workspaceRoot, moduleRoot)
	logger.Debug("Resolved module path",
		zap.String("workspaceRoot", workspaceRoot),
		zap.String("moduleRoot", moduleRoot),
		zap.String("absolute", absModuleRoot))

	// Convert to Docker-compatible path format (handles Windows paths)
	dockerPath := ToDockerPath(absModuleRoot)
	logger.Debug("Docker bind mount path",
		zap.String("original", absModuleRoot),
		zap.String("docker", dockerPath))

	// Configure container
	containerConfig := &container.Config{
		Image: SemgrepImage,
		Cmd: []string{
			"semgrep",
			"scan",
			"--config", config,
			"--json",
			"--quiet",
			"/src",
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/src:ro", dockerPath)},
	}

	// Run container and capture output
	output, err := dockerRunner.RunContainer(containerConfig, hostConfig)
	if err != nil {
		logger.Error("Semgrep SAST scan failed", zap.Error(err))
		return nil, fmt.Errorf("semgrep sast failed: %w", err)
	}

	// Strip Docker log headers if present
	cleanOutput := stripDockerLogHeaders(output)

	logger.Debug("Semgrep SAST scan completed",
		zap.Int("outputSize", len(cleanOutput)))

	// Parse JSON output
	var findings interface{}
	if err := json.Unmarshal(cleanOutput, &findings); err != nil {
		return nil, fmt.Errorf("failed to parse semgrep output: %w", err)
	}

	return findings, nil
}
