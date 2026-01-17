// Package internal provides Semgrep security scanner integration via Docker.
package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/ready-to-release/eac/go/eac/commands/internal/dockerutil"
)

// NOTE: SemgrepImage constant removed - now configured via security-tools.yml
// Docker image versions are loaded from .r2r/eac/security-tools.yml configuration

// Mock support for testing.
var mockSemgrepOutput interface{}

// SetMockSemgrepOutput sets a mock Semgrep response for testing.
func SetMockSemgrepOutput(output interface{}) {
	mockSemgrepOutput = output
}

// ResetMockSemgrepOutput clears the mock Semgrep response.
func ResetMockSemgrepOutput() {
	mockSemgrepOutput = nil
}

// getDefaultMockSemgrepOutput returns default mock data for Semgrep.
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

// RunSemgrepSAST executes Semgrep static analysis via Docker.
func RunSemgrepSAST(workspaceRoot, moduleRoot, config, semgrepImage string) (interface{}, error) {
	// Check for mock output (testing only)
	// Priority: in-process mock > environment variable mock
	if mockSemgrepOutput != nil {
		log.Debug("Using mocked Semgrep output (in-process)")
		return mockSemgrepOutput, nil
	}
	if os.Getenv("R2R_MOCK_SECURITY") != "" {
		log.Debug("Using mocked Semgrep output (environment)")
		return getDefaultMockSemgrepOutput(), nil
	}

	// Create Docker runner
	dockerRunner, err := NewOneOffDockerRunner()
	if err != nil {
		return nil, err
	}
	defer dockerRunner.Close()

	// Pull image if needed
	if err := dockerRunner.CheckAndPullImage(semgrepImage); err != nil {
		return nil, err
	}

	log.Infof("Running Semgrep SAST scanner via Docker: moduleRoot=%s config=%s", moduleRoot, config)

	// Resolve module root relative to workspace root
	absModuleRoot := filepath.Join(workspaceRoot, moduleRoot)
	log.Debugf("Resolved module path: workspaceRoot=%s moduleRoot=%s absolute=%s", workspaceRoot, moduleRoot, absModuleRoot)

	// Translate path for Docker-in-Docker environments
	hostPath, err := dockerutil.TranslatePathForMount(absModuleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to translate path for mount: %w", err)
	}

	// Convert to Docker-compatible path format (handles Windows paths)
	dockerPath := dockerutil.FormatDockerVolume(hostPath)
	log.Debugf("Docker bind mount path: container=%s host=%s docker=%s", absModuleRoot, hostPath, dockerPath)

	// Configure container
	containerConfig := &container.Config{
		Image: semgrepImage,
		Cmd: []string{
			"semgrep",
			"scan",
			"--config", config,
			"--json",
			// Exclude problematic directories that cause scanning issues
			"--exclude", "mnt",
			"--exclude", ".git",
			"--exclude", "node_modules",
			"--exclude", "vendor",
			"--exclude", "out",
			"--exclude", "bin",
			"/src",
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/src:ro", dockerPath)},
	}

	// Run container and capture output
	output, err := dockerRunner.RunContainer(containerConfig, hostConfig)
	if err != nil {
		log.Errorf("Semgrep SAST scan failed: %v", err)
		return nil, fmt.Errorf("semgrep sast failed: %w", err)
	}

	// Strip Docker log headers if present
	cleanOutput := stripDockerLogHeaders(output)

	log.Debugf("Semgrep SAST scan completed: outputSize=%d", len(cleanOutput))

	// Parse JSON output
	var findings interface{}
	if err := json.Unmarshal(cleanOutput, &findings); err != nil {
		return nil, fmt.Errorf("failed to parse semgrep output: %w", err)
	}

	return findings, nil
}
