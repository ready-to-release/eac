package scanners

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/impl/scan/internal"
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

func TestTrivySBOMAdapter_RequiresContext(t *testing.T) {
	// Ensure context is nil
	GlobalScanContext = nil

	_, err := trivySBOMAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err == nil {
		t.Error("expected error when GlobalScanContext is nil")
	}
	if err.Error() != "scan context not initialized: trivy sbom requires GlobalScanContext to be set" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestTrivySBOMAdapter_RequiresTrivyImage(t *testing.T) {
	GlobalScanContext = &ScanContext{
		TrivyImage: "", // Empty image
	}
	defer func() { GlobalScanContext = nil }()

	_, err := trivySBOMAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err == nil {
		t.Error("expected error when TrivyImage is empty")
	}
	if err.Error() != "trivy image not configured in scan context" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestTrivySBOMAdapter_WithMockOutput(t *testing.T) {
	// Enable mock mode via environment variable
	os.Setenv("R2R_MOCK_SECURITY", "1")
	defer os.Unsetenv("R2R_MOCK_SECURITY")

	GlobalScanContext = &ScanContext{
		TrivyImage: "aquasec/trivy:latest",
		SBOMFormat: "cyclonedx-json",
	}
	defer func() { GlobalScanContext = nil }()

	var logBuf bytes.Buffer
	findings, err := trivySBOMAdapter("workspace", "module", "output", &logBuf, tool.ScanOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if findings == nil {
		t.Error("expected findings, got nil")
	}

	// Verify mock data structure
	findingsMap, ok := findings.(map[string]interface{})
	if !ok {
		t.Error("expected findings to be map[string]interface{}")
	}
	if _, hasSchema := findingsMap["SchemaVersion"]; !hasSchema {
		t.Error("expected findings to have SchemaVersion")
	}
}

func TestTrivyVulnAdapter_RequiresContext(t *testing.T) {
	GlobalScanContext = nil

	_, err := trivyVulnAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err == nil {
		t.Error("expected error when GlobalScanContext is nil")
	}
}

func TestTrivyVulnAdapter_WithMockOutput(t *testing.T) {
	os.Setenv("R2R_MOCK_SECURITY", "1")
	defer os.Unsetenv("R2R_MOCK_SECURITY")

	GlobalScanContext = &ScanContext{
		TrivyImage:     "aquasec/trivy:latest",
		VulnSeverities: []internal.Severity{internal.SeverityHigh, internal.SeverityCritical},
	}
	defer func() { GlobalScanContext = nil }()

	findings, err := trivyVulnAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if findings == nil {
		t.Error("expected findings, got nil")
	}
}

func TestTrivySecretsAdapter_WithMockOutput(t *testing.T) {
	os.Setenv("R2R_MOCK_SECURITY", "1")
	defer os.Unsetenv("R2R_MOCK_SECURITY")

	GlobalScanContext = &ScanContext{
		TrivyImage: "aquasec/trivy:latest",
	}
	defer func() { GlobalScanContext = nil }()

	findings, err := trivySecretsAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if findings == nil {
		t.Error("expected findings, got nil")
	}
}

func TestTrivyIaCAdapter_WithMockOutput(t *testing.T) {
	os.Setenv("R2R_MOCK_SECURITY", "1")
	defer os.Unsetenv("R2R_MOCK_SECURITY")

	GlobalScanContext = &ScanContext{
		TrivyImage: "aquasec/trivy:latest",
	}
	defer func() { GlobalScanContext = nil }()

	findings, err := trivyIaCAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if findings == nil {
		t.Error("expected findings, got nil")
	}
}

func TestTrivyComplianceAdapter_WithMockOutput(t *testing.T) {
	os.Setenv("R2R_MOCK_SECURITY", "1")
	defer os.Unsetenv("R2R_MOCK_SECURITY")

	GlobalScanContext = &ScanContext{
		TrivyImage:         "aquasec/trivy:latest",
		ComplianceStandard: "docker-cis",
	}
	defer func() { GlobalScanContext = nil }()

	findings, err := trivyComplianceAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if findings == nil {
		t.Error("expected findings, got nil")
	}
}

func TestTrivyComplianceAdapter_DefaultCompliance(t *testing.T) {
	os.Setenv("R2R_MOCK_SECURITY", "1")
	defer os.Unsetenv("R2R_MOCK_SECURITY")

	GlobalScanContext = &ScanContext{
		TrivyImage:         "aquasec/trivy:latest",
		ComplianceStandard: "", // Empty = use default
	}
	defer func() { GlobalScanContext = nil }()

	// Should not error - uses default compliance standard
	findings, err := trivyComplianceAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if findings == nil {
		t.Error("expected findings, got nil")
	}
}

func TestSemgrepSASTAdapter_RequiresContext(t *testing.T) {
	GlobalScanContext = nil

	_, err := semgrepSASTAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err == nil {
		t.Error("expected error when GlobalScanContext is nil")
	}
}

func TestSemgrepSASTAdapter_RequiresSemgrepImage(t *testing.T) {
	GlobalScanContext = &ScanContext{
		SemgrepImage: "", // Empty image
	}
	defer func() { GlobalScanContext = nil }()

	_, err := semgrepSASTAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err == nil {
		t.Error("expected error when SemgrepImage is empty")
	}
}

func TestSemgrepSASTAdapter_WithMockOutput(t *testing.T) {
	os.Setenv("R2R_MOCK_SECURITY", "1")
	defer os.Unsetenv("R2R_MOCK_SECURITY")

	GlobalScanContext = &ScanContext{
		SemgrepImage:  "returntocorp/semgrep:latest",
		SemgrepConfig: "auto",
	}
	defer func() { GlobalScanContext = nil }()

	findings, err := semgrepSASTAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if findings == nil {
		t.Error("expected findings, got nil")
	}
}

func TestSemgrepSASTAdapter_DefaultConfig(t *testing.T) {
	os.Setenv("R2R_MOCK_SECURITY", "1")
	defer os.Unsetenv("R2R_MOCK_SECURITY")

	GlobalScanContext = &ScanContext{
		SemgrepImage:  "returntocorp/semgrep:latest",
		SemgrepConfig: "", // Empty = use default "auto"
	}
	defer func() { GlobalScanContext = nil }()

	findings, err := semgrepSASTAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if findings == nil {
		t.Error("expected findings, got nil")
	}
}

func TestZAPDASTAdapter_RequiresContext(t *testing.T) {
	GlobalScanContext = nil

	_, err := zapDASTAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err == nil {
		t.Error("expected error when GlobalScanContext is nil")
	}
}

func TestZAPDASTAdapter_RequiresZAPImage(t *testing.T) {
	GlobalScanContext = &ScanContext{
		ZAPImage: "", // Empty image
	}
	defer func() { GlobalScanContext = nil }()

	_, err := zapDASTAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err == nil {
		t.Error("expected error when ZAPImage is empty")
	}
}

func TestZAPDASTAdapter_RequiresTargetURL(t *testing.T) {
	GlobalScanContext = &ScanContext{
		ZAPImage:     "owasp/zap2docker-stable:latest",
		ZAPTargetURL: "", // Empty target
	}
	defer func() { GlobalScanContext = nil }()

	_, err := zapDASTAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err == nil {
		t.Error("expected error when ZAPTargetURL is empty")
	}
}

func TestZAPDASTAdapter_WithMockOutput(t *testing.T) {
	os.Setenv("R2R_MOCK_SECURITY", "1")
	defer os.Unsetenv("R2R_MOCK_SECURITY")

	GlobalScanContext = &ScanContext{
		ZAPImage:     "owasp/zap2docker-stable:latest",
		ZAPTargetURL: "http://localhost:8080",
		ZAPScanType:  "baseline",
	}
	defer func() { GlobalScanContext = nil }()

	findings, err := zapDASTAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if findings == nil {
		t.Error("expected findings, got nil")
	}
}

func TestZAPDASTAdapter_DefaultScanType(t *testing.T) {
	os.Setenv("R2R_MOCK_SECURITY", "1")
	defer os.Unsetenv("R2R_MOCK_SECURITY")

	GlobalScanContext = &ScanContext{
		ZAPImage:     "owasp/zap2docker-stable:latest",
		ZAPTargetURL: "http://localhost:8080",
		ZAPScanType:  "", // Empty = use default "baseline"
	}
	defer func() { GlobalScanContext = nil }()

	findings, err := zapDASTAdapter("workspace", "module", "output", nil, tool.ScanOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if findings == nil {
		t.Error("expected findings, got nil")
	}
}

func TestRegistryIntegration(t *testing.T) {
	// Initialize tool system from default config
	repoRoot := findRepoRoot(t)
	registry, _, err := tool.InitializeFromConfig(repoRoot, "")
	if err != nil {
		t.Fatalf("failed to initialize tool system: %v", err)
	}
	tool.SetGlobalRegistry(registry)

	// Verify that native scanners are registered
	scannerTypes := []tool.ScannerType{
		tool.ScannerSBOM,
		tool.ScannerVuln,
		tool.ScannerSecrets,
		tool.ScannerIaC,
		tool.ScannerCompliance,
		tool.ScannerSAST,
		tool.ScannerDAST,
	}

	for _, st := range scannerTypes {
		if !HasScanner(st) {
			t.Errorf("expected scanner %s to be registered", st)
		}

		scanFn := GetScanner(st)
		if scanFn == nil {
			t.Errorf("expected GetScanner(%s) to return non-nil function", st)
		}
	}
}

// findRepoRoot finds the repository root by looking for go.work
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root (go.work)")
		}
		dir = parent
	}
}

func TestParseScannerType(t *testing.T) {
	tests := []struct {
		input    string
		expected ScannerType
		valid    bool
	}{
		{"sbom", ScannerSBOM, true},
		{"vuln", ScannerVuln, true},
		{"secrets", ScannerSecrets, true},
		{"compliance", ScannerCompliance, true},
		{"iac", ScannerIaC, true},
		{"sast", ScannerSAST, true},
		{"zap", ScannerDAST, true},
		{"invalid", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, valid := ParseScannerType(tt.input)
			if valid != tt.valid {
				t.Errorf("ParseScannerType(%q) valid = %v, want %v", tt.input, valid, tt.valid)
			}
			if valid && got != tt.expected {
				t.Errorf("ParseScannerType(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
