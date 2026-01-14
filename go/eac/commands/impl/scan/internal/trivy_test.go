//go:build L0 && ov
// +build L0,ov

package internal

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/logging"
)

func TestMockTrivyOutput(t *testing.T) {
	// Setup mock
	mockData := map[string]interface{}{
		"test":     "mock sbom data",
		"packages": []string{"pkg1", "pkg2"},
	}
	SetMockTrivyOutput(mockData)
	defer ResetMockTrivyOutput()

	// Create logger
	tmpDir := t.TempDir()
	logger, err := logging.NewDefault("trivy-test", tmpDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Test SBOM with mock
	result, err := RunTrivySBOM(tmpDir, "/fake/path", "cyclonedx", "test-image:latest", logger)
	if err != nil {
		t.Errorf("RunTrivySBOM() with mock error = %v", err)
	}
	if result == nil {
		t.Error("RunTrivySBOM() with mock returned nil")
	}

	// Verify we got the mock data back
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}
	if resultMap["test"] != "mock sbom data" {
		t.Errorf("Mock data not returned correctly")
	}
}

func TestResetMockTrivyOutput(t *testing.T) {
	// Set mock
	SetMockTrivyOutput(map[string]interface{}{"test": "data"})

	// Reset
	ResetMockTrivyOutput()

	tmpDir := t.TempDir()
	logger, err := logging.NewDefault("trivy-test", tmpDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Should try to run real trivy (may succeed or fail depending on Docker availability)
	result, err := RunTrivySBOM(tmpDir, "/fake/path", "cyclonedx", "test-image:latest", logger)
	// If Docker is available, trivy might run successfully (though may error on invalid path)
	// If Docker is not available, we expect an error
	// The key is that we're not using mock data anymore
	if err == nil {
		// Trivy ran successfully - verify we didn't get mock data
		if resultMap, ok := result.(map[string]interface{}); ok {
			if testVal, exists := resultMap["test"]; exists && testVal == "data" {
				t.Error("Got mock data even though mock was reset")
			}
		}
		t.Log("Trivy available and ran successfully")
	} else {
		// Expected error when Docker/Trivy not available
		t.Logf("Expected error occurred: %v", err)
	}
}

func TestRunTrivyVulnWithMock(t *testing.T) {
	mockData := map[string]interface{}{
		"vulnerabilities": []map[string]string{
			{"cve": "CVE-2024-1234", "severity": "HIGH"},
		},
	}
	SetMockTrivyOutput(mockData)
	defer ResetMockTrivyOutput()

	tmpDir := t.TempDir()
	logger, err := logging.NewDefault("trivy-test", tmpDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	severityFilter := []Severity{SeverityHigh, SeverityCritical}
	result, err := RunTrivyVuln("/fake/path", severityFilter, "test-image:latest", logger)
	if err != nil {
		t.Errorf("RunTrivyVuln() with mock error = %v", err)
	}
	if result == nil {
		t.Error("RunTrivyVuln() with mock returned nil")
	}
}

func TestRunTrivySecretsWithMock(t *testing.T) {
	mockData := map[string]interface{}{
		"secrets": []map[string]string{
			{"type": "AWS_ACCESS_KEY", "file": "config.yaml"},
		},
	}
	SetMockTrivyOutput(mockData)
	defer ResetMockTrivyOutput()

	tmpDir := t.TempDir()
	logger, err := logging.NewDefault("trivy-test", tmpDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	result, err := RunTrivySecrets("/fake/path", "test-image:latest", logger)
	if err != nil {
		t.Errorf("RunTrivySecrets() with mock error = %v", err)
	}
	if result == nil {
		t.Error("RunTrivySecrets() with mock returned nil")
	}
}

func TestRunTrivyComplianceWithMock(t *testing.T) {
	mockData := map[string]interface{}{
		"compliance": "k8s-cis",
		"results":    []string{"check1", "check2"},
	}
	SetMockTrivyOutput(mockData)
	defer ResetMockTrivyOutput()

	tmpDir := t.TempDir()
	logger, err := logging.NewDefault("trivy-test", tmpDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	result, err := RunTrivyCompliance("/fake/path", "k8s-cis", "test-image:latest", logger)
	if err != nil {
		t.Errorf("RunTrivyCompliance() with mock error = %v", err)
	}
	if result == nil {
		t.Error("RunTrivyCompliance() with mock returned nil")
	}
}

func TestRunTrivyIaCWithMock(t *testing.T) {
	mockData := map[string]interface{}{
		"misconfigurations": []map[string]string{
			{"id": "AVD-AWS-0001", "severity": "MEDIUM"},
		},
	}
	SetMockTrivyOutput(mockData)
	defer ResetMockTrivyOutput()

	tmpDir := t.TempDir()
	logger, err := logging.NewDefault("trivy-test", tmpDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	result, err := RunTrivyIaC("/fake/path", "test-image:latest", logger)
	if err != nil {
		t.Errorf("RunTrivyIaC() with mock error = %v", err)
	}
	if result == nil {
		t.Error("RunTrivyIaC() with mock returned nil")
	}
}

func TestJoinSeverities(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"Single severity", []string{"HIGH"}, "HIGH"},
		{"Multiple severities", []string{"HIGH", "CRITICAL"}, "HIGH,CRITICAL"},
		{"Three severities", []string{"LOW", "MEDIUM", "HIGH"}, "LOW,MEDIUM,HIGH"},
		{"Empty", []string{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinSeverities(tt.input)
			if result != tt.expected {
				t.Errorf("joinSeverities(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
