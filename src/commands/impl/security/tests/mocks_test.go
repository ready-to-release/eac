package tests

import (
	"os"
	"testing"

	"github.com/ready-to-release/eac/src/commands/impl/security/internal"
	"github.com/ready-to-release/eac/src/core/logging"
)

func TestTrivyMockOutput(t *testing.T) {
	// Set mock output
	mockData := map[string]interface{}{
		"test": "data",
	}
	internal.SetMockTrivyOutput(mockData)
	defer internal.ResetMockTrivyOutput()

	// Create logger for test
	tempDir := t.TempDir()
	logger, err := logging.NewDefault("test", tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Run scanner - should return mock data
	result, err := internal.RunTrivySBOM(tempDir, "json", logger)
	if err != nil {
		t.Fatalf("RunTrivySBOM failed: %v", err)
	}

	// Verify mock data was returned
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map")
	}

	if resultMap["test"] != "data" {
		t.Errorf("Mock data not returned correctly")
	}
}

func TestSemgrepMockOutput(t *testing.T) {
	mockData := map[string]interface{}{
		"results": []string{"test"},
	}
	internal.SetMockSemgrepOutput(mockData)
	defer internal.ResetMockSemgrepOutput()

	tempDir := t.TempDir()
	logger, err := logging.NewDefault("test", tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	result, err := internal.RunSemgrepSAST(tempDir, "auto", logger)
	if err != nil {
		t.Fatalf("RunSemgrepSAST failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map")
	}

	if resultMap["results"] == nil {
		t.Errorf("Mock data not returned correctly")
	}
}

func TestZAPMockOutput(t *testing.T) {
	mockData := map[string]interface{}{
		"site": []string{"test"},
	}
	internal.SetMockZAPOutput(mockData)
	defer internal.ResetMockZAPOutput()

	tempDir := t.TempDir()
	logger, err := logging.NewDefault("test", tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	result, err := internal.RunZAPScan("http://localhost:8080", "baseline", tempDir, logger)
	if err != nil {
		t.Fatalf("RunZAPScan failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map")
	}

	if resultMap["site"] == nil {
		t.Errorf("Mock data not returned correctly")
	}
}

func TestResetMocks(t *testing.T) {
	// Set all mocks
	internal.SetMockTrivyOutput(map[string]interface{}{"test": "trivy"})
	internal.SetMockSemgrepOutput(map[string]interface{}{"test": "semgrep"})
	internal.SetMockZAPOutput(map[string]interface{}{"test": "zap"})

	// Reset all
	internal.ResetMockTrivyOutput()
	internal.ResetMockSemgrepOutput()
	internal.ResetMockZAPOutput()

	// Verify mocks are cleared by checking that scanners would try to run real tools
	// (we can't actually test this without having the tools installed, but we can verify
	// the reset functions don't panic)
}

func TestMockPriority(t *testing.T) {
	// Test that in-process mocks take priority over environment mocking
	os.Setenv("SECURITY_MOCK_TOOLS", "true")
	defer os.Unsetenv("SECURITY_MOCK_TOOLS")

	mockData := map[string]interface{}{
		"priority": "in-process",
	}
	internal.SetMockTrivyOutput(mockData)
	defer internal.ResetMockTrivyOutput()

	tempDir := t.TempDir()
	logger, err := logging.NewDefault("test", tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	result, err := internal.RunTrivySBOM(tempDir, "json", logger)
	if err != nil {
		t.Fatalf("RunTrivySBOM failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map")
	}

	// Should get in-process mock data, not environment default
	if resultMap["priority"] != "in-process" {
		t.Error("In-process mock should take priority over environment mock")
	}
}

func TestAllTrivyScannersMockable(t *testing.T) {
	mockData := map[string]interface{}{"test": "mock"}
	internal.SetMockTrivyOutput(mockData)
	defer internal.ResetMockTrivyOutput()

	tempDir := t.TempDir()
	logger, err := logging.NewDefault("test", tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	tests := []struct {
		name string
		fn   func() (interface{}, error)
	}{
		{"SBOM", func() (interface{}, error) {
			return internal.RunTrivySBOM(tempDir, "json", logger)
		}},
		{"Vuln", func() (interface{}, error) {
			return internal.RunTrivyVuln(tempDir, nil, logger)
		}},
		{"Secrets", func() (interface{}, error) {
			return internal.RunTrivySecrets(tempDir, logger)
		}},
		{"Compliance", func() (interface{}, error) {
			return internal.RunTrivyCompliance(tempDir, "k8s-cis", logger)
		}},
		{"IaC", func() (interface{}, error) {
			return internal.RunTrivyIaC(tempDir, logger)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn()
			if err != nil {
				t.Errorf("%s failed: %v", tt.name, err)
			}
			if result == nil {
				t.Errorf("%s returned nil result", tt.name)
			}
		})
	}
}
