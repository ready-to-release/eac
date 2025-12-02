// Package internal provides Trivy security scanner integration via Docker.
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
	// TrivyImage is the official Trivy Docker image
	TrivyImage = "ghcr.io/aquasecurity/trivy:latest"
)

// Mock support for testing
var mockTrivyOutput interface{}

// SetMockTrivyOutput sets a mock Trivy response for testing
func SetMockTrivyOutput(output interface{}) {
	mockTrivyOutput = output
}

// ResetMockTrivyOutput clears the mock Trivy response
func ResetMockTrivyOutput() {
	mockTrivyOutput = nil
}

// getDefaultMockTrivyOutput returns default mock data for Trivy
func getDefaultMockTrivyOutput() map[string]interface{} {
	return map[string]interface{}{
		"SchemaVersion": 2,
		"ArtifactName":  "test-module",
		"ArtifactType":  "filesystem",
		"Results": []map[string]interface{}{
			{
				"Target": "test-module",
				"Type":   "test",
			},
		},
	}
}

// RunTrivySBOM executes Trivy SBOM scanner via Docker
func RunTrivySBOM(workspaceRoot, moduleRoot, format string, logger *logging.Logger) (interface{}, error) {
	// Check for mock output (testing only)
	// Priority: in-process mock > environment variable mock
	if mockTrivyOutput != nil {
		logger.Debug("Using mocked Trivy SBOM output (in-process)")
		return mockTrivyOutput, nil
	}
	if os.Getenv("R2R_MOCK_SECURITY") != "" {
		logger.Debug("Using mocked Trivy SBOM output (environment)")
		return getDefaultMockTrivyOutput(), nil
	}

	// Create Docker runner
	dockerRunner, err := NewOneOffDockerRunner(logger)
	if err != nil {
		return nil, err
	}
	defer dockerRunner.Close()

	// Pull image if needed
	if err := dockerRunner.CheckAndPullImage(TrivyImage); err != nil {
		return nil, err
	}

	logger.Info("Running Trivy SBOM scanner via Docker",
		zap.String("moduleRoot", moduleRoot),
		zap.String("format", format))

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
	config := &container.Config{
		Image: TrivyImage,
		Cmd: []string{
			"fs",
			"--format", format,
			"--quiet",
			"/scan",
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/scan:ro", dockerPath)},
	}

	// Run container and capture output
	output, err := dockerRunner.RunContainer(config, hostConfig)
	if err != nil {
		logger.Error("Trivy SBOM scan failed", zap.Error(err))
		return nil, fmt.Errorf("trivy sbom failed: %w", err)
	}

	// Strip Docker log headers if present
	cleanOutput := stripDockerLogHeaders(output)

	logger.Debug("Trivy SBOM scan completed",
		zap.Int("outputSize", len(cleanOutput)))

	// Parse JSON output
	var findings interface{}
	if err := json.Unmarshal(cleanOutput, &findings); err != nil {
		// If JSON parsing fails, the output might be an error message from Trivy
		// Return the actual Trivy error for better debugging
		outputPreview := string(cleanOutput)
		if len(outputPreview) > 200 {
			outputPreview = outputPreview[:200] + "..."
		}
		return nil, fmt.Errorf("trivy returned non-JSON output (likely an error): %s", outputPreview)
	}

	return findings, nil
}

// RunTrivyVuln executes Trivy vulnerability scanner via Docker
func RunTrivyVuln(moduleRoot string, severityFilter []Severity, logger *logging.Logger) (interface{}, error) {
	// Check for mock output (testing only)
	// Priority: in-process mock > environment variable mock
	if mockTrivyOutput != nil {
		logger.Debug("Using mocked Trivy vulnerability output (in-process)")
		return mockTrivyOutput, nil
	}
	if os.Getenv("R2R_MOCK_SECURITY") != "" {
		logger.Debug("Using mocked Trivy vulnerability output (environment)")
		return getDefaultMockTrivyOutput(), nil
	}

	// Create Docker runner
	dockerRunner, err := NewOneOffDockerRunner(logger)
	if err != nil {
		return nil, err
	}
	defer dockerRunner.Close()

	// Pull image if needed
	if err := dockerRunner.CheckAndPullImage(TrivyImage); err != nil {
		return nil, err
	}

	logger.Info("Running Trivy vulnerability scanner via Docker",
		zap.String("moduleRoot", moduleRoot),
		zap.Any("severityFilter", severityFilter))

	// Get absolute path for volume mount
	absModuleRoot, err := filepath.Abs(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Convert to Docker-compatible path format (handles Windows paths)
	dockerPath := ToDockerPath(absModuleRoot)
	logger.Debug("Docker bind mount path",
		zap.String("original", absModuleRoot),
		zap.String("docker", dockerPath))

	// Build command arguments
	cmd := []string{
		"fs",
		"--format", "json",
		"--quiet",
	}

	// Add severity filter if provided
	if len(severityFilter) > 0 {
		severities := make([]string, len(severityFilter))
		for i, sev := range severityFilter {
			severities[i] = string(sev)
		}
		cmd = append(cmd, "--severity", joinSeverities(severities))
	}

	cmd = append(cmd, "/scan")

	// Configure container
	config := &container.Config{
		Image: TrivyImage,
		Cmd:   cmd,
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/scan:ro", dockerPath)},
	}

	// Run container and capture output
	output, err := dockerRunner.RunContainer(config, hostConfig)
	if err != nil {
		logger.Error("Trivy vulnerability scan failed", zap.Error(err))
		return nil, fmt.Errorf("trivy vulnerability scan failed: %w", err)
	}

	// Strip Docker log headers if present
	cleanOutput := stripDockerLogHeaders(output)

	logger.Debug("Trivy vulnerability scan completed",
		zap.Int("outputSize", len(cleanOutput)))

	// Parse JSON output
	var findings interface{}
	if err := json.Unmarshal(cleanOutput, &findings); err != nil {
		// If JSON parsing fails, the output might be an error message from Trivy
		// Return the actual Trivy error for better debugging
		outputPreview := string(cleanOutput)
		if len(outputPreview) > 200 {
			outputPreview = outputPreview[:200] + "..."
		}
		return nil, fmt.Errorf("trivy returned non-JSON output (likely an error): %s", outputPreview)
	}

	return findings, nil
}

// RunTrivySecrets executes Trivy secrets scanner via Docker
func RunTrivySecrets(moduleRoot string, logger *logging.Logger) (interface{}, error) {
	// Check for mock output (testing only)
	// Priority: in-process mock > environment variable mock
	if mockTrivyOutput != nil {
		logger.Debug("Using mocked Trivy secrets output (in-process)")
		return mockTrivyOutput, nil
	}
	if os.Getenv("R2R_MOCK_SECURITY") != "" {
		logger.Debug("Using mocked Trivy secrets output (environment)")
		return getDefaultMockTrivyOutput(), nil
	}

	// Create Docker runner
	dockerRunner, err := NewOneOffDockerRunner(logger)
	if err != nil {
		return nil, err
	}
	defer dockerRunner.Close()

	// Pull image if needed
	if err := dockerRunner.CheckAndPullImage(TrivyImage); err != nil {
		return nil, err
	}

	logger.Info("Running Trivy secrets scanner via Docker",
		zap.String("moduleRoot", moduleRoot))

	// Get absolute path for volume mount
	absModuleRoot, err := filepath.Abs(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Convert to Docker-compatible path format (handles Windows paths)
	dockerPath := ToDockerPath(absModuleRoot)
	logger.Debug("Docker bind mount path",
		zap.String("original", absModuleRoot),
		zap.String("docker", dockerPath))

	// Configure container
	config := &container.Config{
		Image: TrivyImage,
		Cmd: []string{
			"fs",
			"--scanners", "secret",
			"--format", "json",
			"--quiet",
			"/scan",
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/scan:ro", dockerPath)},
	}

	// Run container and capture output
	output, err := dockerRunner.RunContainer(config, hostConfig)
	if err != nil {
		logger.Error("Trivy secrets scan failed", zap.Error(err))
		return nil, fmt.Errorf("trivy secrets scan failed: %w", err)
	}

	// Strip Docker log headers if present
	cleanOutput := stripDockerLogHeaders(output)

	logger.Debug("Trivy secrets scan completed",
		zap.Int("outputSize", len(cleanOutput)))

	// Parse JSON output
	var findings interface{}
	if err := json.Unmarshal(cleanOutput, &findings); err != nil {
		// If JSON parsing fails, the output might be an error message from Trivy
		// Return the actual Trivy error for better debugging
		outputPreview := string(cleanOutput)
		if len(outputPreview) > 200 {
			outputPreview = outputPreview[:200] + "..."
		}
		return nil, fmt.Errorf("trivy returned non-JSON output (likely an error): %s", outputPreview)
	}

	return findings, nil
}

// RunTrivyCompliance executes Trivy compliance scanner via Docker
func RunTrivyCompliance(moduleRoot, compliance string, logger *logging.Logger) (interface{}, error) {
	// Check for mock output (testing only)
	// Priority: in-process mock > environment variable mock
	if mockTrivyOutput != nil {
		logger.Debug("Using mocked Trivy compliance output (in-process)")
		return mockTrivyOutput, nil
	}
	if os.Getenv("R2R_MOCK_SECURITY") != "" {
		logger.Debug("Using mocked Trivy compliance output (environment)")
		return getDefaultMockTrivyOutput(), nil
	}

	// Create Docker runner
	dockerRunner, err := NewOneOffDockerRunner(logger)
	if err != nil {
		return nil, err
	}
	defer dockerRunner.Close()

	// Pull image if needed
	if err := dockerRunner.CheckAndPullImage(TrivyImage); err != nil {
		return nil, err
	}

	logger.Info("Running Trivy compliance scanner via Docker",
		zap.String("moduleRoot", moduleRoot),
		zap.String("compliance", compliance))

	// Get absolute path for volume mount
	absModuleRoot, err := filepath.Abs(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Convert to Docker-compatible path format (handles Windows paths)
	dockerPath := ToDockerPath(absModuleRoot)
	logger.Debug("Docker bind mount path",
		zap.String("original", absModuleRoot),
		zap.String("docker", dockerPath))

	// Configure container
	config := &container.Config{
		Image: TrivyImage,
		Cmd: []string{
			"fs",
			"--compliance", compliance,
			"--format", "json",
			"--quiet",
			"/scan",
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/scan:ro", dockerPath)},
	}

	// Run container and capture output
	output, err := dockerRunner.RunContainer(config, hostConfig)
	if err != nil {
		logger.Error("Trivy compliance scan failed", zap.Error(err))
		return nil, fmt.Errorf("trivy compliance scan failed: %w", err)
	}

	// Strip Docker log headers if present
	cleanOutput := stripDockerLogHeaders(output)

	logger.Debug("Trivy compliance scan completed",
		zap.Int("outputSize", len(cleanOutput)))

	// Parse JSON output
	var findings interface{}
	if err := json.Unmarshal(cleanOutput, &findings); err != nil {
		// If JSON parsing fails, the output might be an error message from Trivy
		// Return the actual Trivy error for better debugging
		outputPreview := string(cleanOutput)
		if len(outputPreview) > 200 {
			outputPreview = outputPreview[:200] + "..."
		}
		return nil, fmt.Errorf("trivy returned non-JSON output (likely an error): %s", outputPreview)
	}

	return findings, nil
}

// RunTrivyIaC executes Trivy Infrastructure as Code scanner via Docker
func RunTrivyIaC(moduleRoot string, logger *logging.Logger) (interface{}, error) {
	// Check for mock output (testing only)
	// Priority: in-process mock > environment variable mock
	if mockTrivyOutput != nil {
		logger.Debug("Using mocked Trivy IaC output (in-process)")
		return mockTrivyOutput, nil
	}
	if os.Getenv("R2R_MOCK_SECURITY") != "" {
		logger.Debug("Using mocked Trivy IaC output (environment)")
		return getDefaultMockTrivyOutput(), nil
	}

	// Create Docker runner
	dockerRunner, err := NewOneOffDockerRunner(logger)
	if err != nil {
		return nil, err
	}
	defer dockerRunner.Close()

	// Pull image if needed
	if err := dockerRunner.CheckAndPullImage(TrivyImage); err != nil {
		return nil, err
	}

	logger.Info("Running Trivy IaC scanner via Docker",
		zap.String("moduleRoot", moduleRoot))

	// Get absolute path for volume mount
	absModuleRoot, err := filepath.Abs(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Convert to Docker-compatible path format (handles Windows paths)
	dockerPath := ToDockerPath(absModuleRoot)
	logger.Debug("Docker bind mount path",
		zap.String("original", absModuleRoot),
		zap.String("docker", dockerPath))

	// Configure container
	config := &container.Config{
		Image: TrivyImage,
		Cmd: []string{
			"config",
			"--format", "json",
			"--quiet",
			"/scan",
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/scan:ro", dockerPath)},
	}

	// Run container and capture output
	output, err := dockerRunner.RunContainer(config, hostConfig)
	if err != nil {
		logger.Error("Trivy IaC scan failed", zap.Error(err))
		return nil, fmt.Errorf("trivy iac scan failed: %w", err)
	}

	// Strip Docker log headers if present
	cleanOutput := stripDockerLogHeaders(output)

	logger.Debug("Trivy IaC scan completed",
		zap.Int("outputSize", len(cleanOutput)))

	// Parse JSON output
	var findings interface{}
	if err := json.Unmarshal(cleanOutput, &findings); err != nil {
		// If JSON parsing fails, the output might be an error message from Trivy
		// Return the actual Trivy error for better debugging
		outputPreview := string(cleanOutput)
		if len(outputPreview) > 200 {
			outputPreview = outputPreview[:200] + "..."
		}
		return nil, fmt.Errorf("trivy returned non-JSON output (likely an error): %s", outputPreview)
	}

	return findings, nil
}

// joinSeverities joins severity strings with commas
func joinSeverities(severities []string) string {
	result := ""
	for i, sev := range severities {
		if i > 0 {
			result += ","
		}
		result += sev
	}
	return result
}
