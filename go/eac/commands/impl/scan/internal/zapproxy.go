// Package internal provides OWASP ZAP security scanner integration.
package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/ready-to-release/eac/go/eac/commands/internal/dockerutil"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

// NOTE: ZAPImage constant removed - now configured via security-tools.yml
// Docker image versions are loaded from .r2r/eac/security-tools.yml configuration

const (
	// ZAP scan types.
	ZAPBaseline = "baseline"
	ZAPFull     = "full"
	ZAPAPI      = "api"
)

// Mock support for testing.
var mockZAPOutput interface{}

// SetMockZAPOutput sets a mock ZAP response for testing.
func SetMockZAPOutput(output interface{}) {
	mockZAPOutput = output
}

// ResetMockZAPOutput clears the mock ZAP response.
func ResetMockZAPOutput() {
	mockZAPOutput = nil
}

// getDefaultMockZAPOutput returns default mock data for ZAP.
func getDefaultMockZAPOutput() map[string]interface{} {
	return map[string]interface{}{
		"site": []map[string]interface{}{
			{
				"@name":  "http://localhost:8080",
				"alerts": []string{},
			},
		},
	}
}

// RunZAPScan executes OWASP ZAP dynamic security scan via Docker.
func RunZAPScan(targetURL, scanType, workspaceRoot, zapImage string) (interface{}, error) {
	// Check for mock output (testing only)
	// Priority: in-process mock > environment variable mock
	if mockZAPOutput != nil {
		log.Debug("Using mocked ZAP output (in-process)")
		return mockZAPOutput, nil
	}

	// Check for environment-based mocking (for subprocess tests)
	if os.Getenv("R2R_MOCK_SECURITY") != "" {
		log.Debug("Using mocked ZAP output (environment)")
		return getDefaultMockZAPOutput(), nil
	}

	// Create Docker runner
	dockerRunner, err := NewOneOffDockerRunner()
	if err != nil {
		return nil, err
	}
	defer dockerRunner.Close()

	// Pull image if needed
	if err := dockerRunner.CheckAndPullImage(zapImage); err != nil {
		return nil, err
	}

	log.Infof("Running OWASP ZAP scan: targetURL=%s scanType=%s", targetURL, scanType)

	// Load config for path resolution
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create temporary output directory for ZAP report
	// Scan path already includes "out" prefix (e.g., "out/scan")
	outputDir := filepath.Join(workspaceRoot, cfg.Repository.Paths.Out.Scan, "zap-temp")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create ZAP output directory: %w", err)
	}
	defer os.RemoveAll(outputDir) // Clean up temp files

	reportPath := filepath.Join(outputDir, "zap-report.json")

	// Get absolute path for volume mount
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Translate path for Docker-in-Docker environments
	hostPath, err := dockerutil.TranslatePathForMount(absOutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to translate path for mount: %w", err)
	}

	// Convert to Docker-compatible path format for volume mount
	dockerOutputDir := dockerutil.FormatDockerVolume(hostPath)
	log.Debugf("Docker bind mount path: container=%s host=%s docker=%s", absOutputDir, hostPath, dockerOutputDir)

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

	// Configure container
	containerConfig := &container.Config{
		Image: zapImage,
		Cmd: []string{
			zapScript,
			"-t", targetURL,
			"-J", "zap-report.json", // JSON output
			"-I", // Don't fail on warnings
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/zap/wrk:rw", dockerOutputDir)},
	}

	// Run container and capture output
	output, err := dockerRunner.RunContainer(containerConfig, hostConfig)
	if err != nil {
		// ZAP may return non-zero exit code even with successful scan
		// Check if report file was created
		if _, statErr := os.Stat(reportPath); statErr != nil {
			log.Errorf("ZAP scan failed: %v", err)
			return nil, fmt.Errorf("zap scan failed: %w", err)
		}
		log.Debugf("ZAP scan completed with warnings: output=%s", string(output))
	} else {
		log.Debugf("ZAP scan completed: outputSize=%d", len(output))
	}

	// Read and parse ZAP report
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
