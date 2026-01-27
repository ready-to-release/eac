// Package internal provides Trivy security scanner integration via Docker.
package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/ready-to-release/eac/go/eac/commands/internal/dockerutil"
)

// NOTE: trivyImage constant removed - now configured via security-tools.yml
// Docker image versions are loaded from .r2r/eac/security-tools.yml configuration

// Mock support for testing.
var mockTrivyOutput interface{}

// SetMockTrivyOutput sets a mock Trivy response for testing.
func SetMockTrivyOutput(output interface{}) {
	mockTrivyOutput = output
}

// ResetMockTrivyOutput clears the mock Trivy response.
func ResetMockTrivyOutput() {
	mockTrivyOutput = nil
}

// getDefaultMockTrivyOutput returns default mock data for Trivy.
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

// RunTrivySBOM executes Trivy SBOM scanner via Docker.
func RunTrivySBOM(workspaceRoot, moduleRoot, format, trivyImage string) (interface{}, error) {
	// Check for mock output (testing only)
	// Priority: in-process mock > environment variable mock
	if mockTrivyOutput != nil {
		log.Debug("Using mocked Trivy SBOM output (in-process)")
		return mockTrivyOutput, nil
	}
	if os.Getenv("R2R_MOCK_SECURITY") != "" {
		log.Debug("Using mocked Trivy SBOM output (environment)")
		return getDefaultMockTrivyOutput(), nil
	}

	// Create Docker runner
	dockerRunner, err := NewOneOffDockerRunner()
	if err != nil {
		return nil, err
	}
	defer dockerRunner.Close()

	// Pull image if needed
	if err := dockerRunner.CheckAndPullImage(trivyImage); err != nil {
		return nil, err
	}

	log.Debugf("Running Trivy SBOM generator via Docker: moduleRoot=%s format=%s scanners=vuln list-all-pkgs=true", moduleRoot, format)

	// Resolve module root relative to workspace root
	absModuleRoot := filepath.Join(workspaceRoot, moduleRoot)
	log.Debugf("Resolved module path: workspaceRoot=%s moduleRoot=%s absolute=%s", workspaceRoot, moduleRoot, absModuleRoot)

	// Translate path for Docker-in-Docker environments
	hostPath, err := dockerutil.TranslatePathForMount(absModuleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to translate path for mount: %w", err)
	}
	log.Debugf("Translated path for DinD: container=%s host=%s", absModuleRoot, hostPath)

	// Convert to Docker-compatible path format (handles Windows paths)
	dockerPath := dockerutil.FormatDockerVolume(hostPath)
	log.Debugf("Docker bind mount path: host=%s docker=%s", hostPath, dockerPath)

	// Configure container
	config := &container.Config{
		Image: trivyImage,
		Cmd: []string{
			"fs",
			"--scanners", "vuln", // Enable vulnerability/package analysis for SBOM generation
			"--format", format,
			"--list-all-pkgs", // Include all packages in SBOM, not just vulnerable ones
			"--quiet",
			// Skip problematic directories that cause scanning issues
			"--skip-dirs", "/scan/mnt,/scan/.git,/scan/node_modules,/scan/vendor,/scan/out,/scan/bin",
			"/scan",
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/scan:ro", dockerPath)},
	}

	// Run container and capture output
	output, err := dockerRunner.RunContainer(config, hostConfig)
	if err != nil {
		return nil, fmt.Errorf("trivy sbom failed: %w", err)
	}

	// Strip Docker log headers if present
	cleanOutput := stripDockerLogHeaders(output)

	log.Debugf("Trivy SBOM scan completed: outputSize=%d", len(cleanOutput))

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

// RunTrivyVuln executes Trivy vulnerability scanner via Docker.
func RunTrivyVuln(moduleRoot string, severityFilter []Severity, trivyImage string) (interface{}, error) {
	// Check for mock output (testing only)
	// Priority: in-process mock > environment variable mock
	if mockTrivyOutput != nil {
		log.Debug("Using mocked Trivy vulnerability output (in-process)")
		return mockTrivyOutput, nil
	}
	if os.Getenv("R2R_MOCK_SECURITY") != "" {
		log.Debug("Using mocked Trivy vulnerability output (environment)")
		return getDefaultMockTrivyOutput(), nil
	}

	// Create Docker runner
	dockerRunner, err := NewOneOffDockerRunner()
	if err != nil {
		return nil, err
	}
	defer dockerRunner.Close()

	// Pull image if needed
	if err := dockerRunner.CheckAndPullImage(trivyImage); err != nil {
		return nil, err
	}

	log.Debugf("Running Trivy vulnerability scanner via Docker: moduleRoot=%s severityFilter=%v", moduleRoot, severityFilter)

	// Get absolute path for volume mount
	absModuleRoot, err := filepath.Abs(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Translate path for Docker-in-Docker environments
	hostPath, err := dockerutil.TranslatePathForMount(absModuleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to translate path for mount: %w", err)
	}

	// Convert to Docker-compatible path format (handles Windows paths)
	dockerPath := dockerutil.FormatDockerVolume(hostPath)
	log.Debugf("Docker bind mount path: container=%s host=%s docker=%s", absModuleRoot, hostPath, dockerPath)

	// Build command arguments
	cmd := []string{
		"fs",
		"--scanners", "vuln", // Only vulnerability scanning, not secrets
		"--format", "json",
		"--quiet",
		// Skip problematic directories that cause scanning issues
		"--skip-dirs", "/scan/mnt,/scan/.git,/scan/node_modules,/scan/vendor,/scan/out,/scan/bin",
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
		Image: trivyImage,
		Cmd:   cmd,
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/scan:ro", dockerPath)},
	}

	// Run container and capture output
	output, err := dockerRunner.RunContainer(config, hostConfig)
	if err != nil {
		return nil, fmt.Errorf("trivy vulnerability scan failed: %w", err)
	}

	// Strip Docker log headers if present
	cleanOutput := stripDockerLogHeaders(output)

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

	// Normalize Trivy output: ensure Results field exists
	// When no vulnerabilities are found, Trivy omits the Results field entirely.
	// We add an empty array for consistency and clarity.
	if findingsMap, ok := findings.(map[string]interface{}); ok {
		if findingsMap["Results"] == nil {
			findingsMap["Results"] = []interface{}{}
			log.Debug("Normalized Trivy output: added empty Results array (no vulnerabilities found)")
		}
	}

	return findings, nil
}

// RunTrivySecrets executes Trivy secrets scanner via Docker.
func RunTrivySecrets(moduleRoot, trivyImage string) (interface{}, error) {
	// Check for mock output (testing only)
	// Priority: in-process mock > environment variable mock
	if mockTrivyOutput != nil {
		log.Debug("Using mocked Trivy secrets output (in-process)")
		return mockTrivyOutput, nil
	}
	if os.Getenv("R2R_MOCK_SECURITY") != "" {
		log.Debug("Using mocked Trivy secrets output (environment)")
		return getDefaultMockTrivyOutput(), nil
	}

	// Create Docker runner
	dockerRunner, err := NewOneOffDockerRunner()
	if err != nil {
		return nil, err
	}
	defer dockerRunner.Close()

	// Pull image if needed
	if err := dockerRunner.CheckAndPullImage(trivyImage); err != nil {
		return nil, err
	}

	log.Debugf("Running Trivy secrets scanner via Docker: moduleRoot=%s", moduleRoot)

	// Get absolute path for volume mount
	absModuleRoot, err := filepath.Abs(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Translate path for Docker-in-Docker environments
	hostPath, err := dockerutil.TranslatePathForMount(absModuleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to translate path for mount: %w", err)
	}

	// Convert to Docker-compatible path format (handles Windows paths)
	dockerPath := dockerutil.FormatDockerVolume(hostPath)
	log.Debugf("Docker bind mount path: container=%s host=%s docker=%s", absModuleRoot, hostPath, dockerPath)

	// Configure container
	config := &container.Config{
		Image: trivyImage,
		Cmd: []string{
			"fs",
			"--scanners", "secret",
			"--format", "json",
			"--quiet",
			// Skip problematic directories that cause scanning issues
			"--skip-dirs", "/scan/mnt,/scan/.git,/scan/node_modules,/scan/vendor,/scan/out,/scan/bin",
			"/scan",
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/scan:ro", dockerPath)},
	}

	// Run container and capture output
	output, err := dockerRunner.RunContainer(config, hostConfig)
	if err != nil {
		log.Debugf("Trivy secrets scan failed: %v", err)
		return nil, fmt.Errorf("trivy secrets scan failed: %w", err)
	}

	// Strip Docker log headers if present
	cleanOutput := stripDockerLogHeaders(output)

	log.Debugf("Trivy secrets scan completed: outputSize=%d", len(cleanOutput))

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

// RunTrivyCompliance executes Trivy compliance scanner via Docker.
func RunTrivyCompliance(moduleRoot, compliance, trivyImage string) (interface{}, error) {
	// Check for mock output (testing only)
	// Priority: in-process mock > environment variable mock
	if mockTrivyOutput != nil {
		log.Debug("Using mocked Trivy compliance output (in-process)")
		return mockTrivyOutput, nil
	}
	if os.Getenv("R2R_MOCK_SECURITY") != "" {
		log.Debug("Using mocked Trivy compliance output (environment)")
		return getDefaultMockTrivyOutput(), nil
	}

	// Create Docker runner
	dockerRunner, err := NewOneOffDockerRunner()
	if err != nil {
		return nil, err
	}
	defer dockerRunner.Close()

	// Pull image if needed
	if err := dockerRunner.CheckAndPullImage(trivyImage); err != nil {
		return nil, err
	}

	log.Debugf("Running Trivy compliance scanner via Docker: moduleRoot=%s compliance=%s", moduleRoot, compliance)

	// Get absolute path for volume mount
	absModuleRoot, err := filepath.Abs(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Translate path for Docker-in-Docker environments
	hostPath, err := dockerutil.TranslatePathForMount(absModuleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to translate path for mount: %w", err)
	}

	// Convert to Docker-compatible path format (handles Windows paths)
	dockerPath := dockerutil.FormatDockerVolume(hostPath)
	log.Debugf("Docker bind mount path: container=%s host=%s docker=%s", absModuleRoot, hostPath, dockerPath)

	// Configure container
	config := &container.Config{
		Image: trivyImage,
		Cmd: []string{
			"fs",
			"--compliance", compliance,
			"--format", "json",
			"--quiet",
			// Skip problematic directories that cause scanning issues
			"--skip-dirs", "/scan/mnt,/scan/.git,/scan/node_modules,/scan/vendor,/scan/out,/scan/bin",
			"/scan",
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/scan:ro", dockerPath)},
	}

	// Run container and capture output
	output, err := dockerRunner.RunContainer(config, hostConfig)
	if err != nil {
		log.Debugf("Trivy compliance scan failed: %v", err)
		return nil, fmt.Errorf("trivy compliance scan failed: %w", err)
	}

	// Strip Docker log headers if present
	cleanOutput := stripDockerLogHeaders(output)

	log.Debugf("Trivy compliance scan completed: outputSize=%d", len(cleanOutput))

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

// RunTrivyIaC executes Trivy Infrastructure as Code scanner via Docker.
func RunTrivyIaC(moduleRoot, trivyImage string) (interface{}, error) {
	// Check for mock output (testing only)
	// Priority: in-process mock > environment variable mock
	if mockTrivyOutput != nil {
		log.Debug("Using mocked Trivy IaC output (in-process)")
		return mockTrivyOutput, nil
	}
	if os.Getenv("R2R_MOCK_SECURITY") != "" {
		log.Debug("Using mocked Trivy IaC output (environment)")
		return getDefaultMockTrivyOutput(), nil
	}

	// Create Docker runner
	dockerRunner, err := NewOneOffDockerRunner()
	if err != nil {
		return nil, err
	}
	defer dockerRunner.Close()

	// Pull image if needed
	if err := dockerRunner.CheckAndPullImage(trivyImage); err != nil {
		return nil, err
	}

	log.Debugf("Running Trivy IaC scanner via Docker: moduleRoot=%s", moduleRoot)

	// Get absolute path for volume mount
	absModuleRoot, err := filepath.Abs(moduleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Translate path for Docker-in-Docker environments
	hostPath, err := dockerutil.TranslatePathForMount(absModuleRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to translate path for mount: %w", err)
	}

	// Convert to Docker-compatible path format (handles Windows paths)
	dockerPath := dockerutil.FormatDockerVolume(hostPath)
	log.Debugf("Docker bind mount path: container=%s host=%s docker=%s", absModuleRoot, hostPath, dockerPath)

	// Configure container
	config := &container.Config{
		Image: trivyImage,
		Cmd: []string{
			"config",
			"--format", "json",
			"--quiet",
			// Skip problematic directories that cause scanning issues
			"--skip-dirs", "/scan/mnt,/scan/.git,/scan/node_modules,/scan/vendor,/scan/out,/scan/bin",
			"/scan",
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/scan:ro", dockerPath)},
	}

	// Run container and capture output
	output, err := dockerRunner.RunContainer(config, hostConfig)
	if err != nil {
		log.Debugf("Trivy IaC scan failed: %v", err)
		return nil, fmt.Errorf("trivy iac scan failed: %w", err)
	}

	// Strip Docker log headers if present
	cleanOutput := stripDockerLogHeaders(output)

	log.Debugf("Trivy IaC scan completed: outputSize=%d", len(cleanOutput))

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

// joinSeverities joins severity strings with commas.
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
