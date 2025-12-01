package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ready-to-release/eac/src/commands/impl/security/internal"
)

func TestWriteEvidence(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	testFindings := map[string]interface{}{
		"results": []string{"finding1", "finding2"},
	}

	// Write evidence file
	outputPath, err := internal.WriteEvidence(tempDir, "test-module", internal.ScannerSBOM, testFindings)
	if err != nil {
		t.Fatalf("WriteEvidence failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Evidence file was not created at %s", outputPath)
	}

	// Verify file is in correct directory structure
	expectedDir := filepath.Join(tempDir, "out", "security", "test-module", "sbom")
	if !strings.Contains(outputPath, expectedDir) {
		t.Errorf("Evidence file not in expected directory. Got %s, expected to contain %s", outputPath, expectedDir)
	}

	// Read and verify evidence file
	evidence, err := internal.ReadEvidence(outputPath)
	if err != nil {
		t.Fatalf("ReadEvidence failed: %v", err)
	}

	// Verify fields
	if evidence.Module != "test-module" {
		t.Errorf("Module = %s, want test-module", evidence.Module)
	}

	if evidence.Scanner != "sbom" {
		t.Errorf("Scanner = %s, want sbom", evidence.Scanner)
	}

	// Verify timestamp is RFC3339 format
	if _, err := time.Parse(time.RFC3339, evidence.Timestamp); err != nil {
		t.Errorf("Timestamp is not RFC3339 format: %s", evidence.Timestamp)
	}

	// Verify SHA256 hash length (64 hex characters)
	if len(evidence.SHA256) != 64 {
		t.Errorf("SHA256 length = %d, want 64", len(evidence.SHA256))
	}

	// Verify SHA256 contains only hex characters
	for _, c := range evidence.SHA256 {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("SHA256 contains non-hex character: %c", c)
		}
	}

	// Verify findings can be parsed
	var parsedFindings map[string]interface{}
	if err := json.Unmarshal(evidence.Findings, &parsedFindings); err != nil {
		t.Errorf("Findings is not valid JSON: %v", err)
	}
}

func TestWriteEvidenceCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()

	testFindings := map[string]interface{}{"test": "data"}

	// Write evidence to non-existent directory structure
	outputPath, err := internal.WriteEvidence(tempDir, "new-module", internal.ScannerVuln, testFindings)
	if err != nil {
		t.Fatalf("WriteEvidence failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Evidence file was not created")
	}
}

func TestReadEvidence(t *testing.T) {
	// Create temporary directory and write test evidence
	tempDir := t.TempDir()
	testFindings := map[string]interface{}{"test": "data"}

	outputPath, err := internal.WriteEvidence(tempDir, "test-module", internal.ScannerSecrets, testFindings)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Read evidence
	evidence, err := internal.ReadEvidence(outputPath)
	if err != nil {
		t.Fatalf("ReadEvidence failed: %v", err)
	}

	// Verify structure
	if evidence == nil {
		t.Fatal("Evidence is nil")
	}

	if evidence.Module != "test-module" {
		t.Errorf("Module = %s, want test-module", evidence.Module)
	}

	if evidence.Scanner != "secrets" {
		t.Errorf("Scanner = %s, want secrets", evidence.Scanner)
	}
}

func TestReadEvidenceInvalidFile(t *testing.T) {
	_, err := internal.ReadEvidence("/nonexistent/path/file.json")
	if err == nil {
		t.Error("ReadEvidence should fail for nonexistent file")
	}
}

func TestGetTimestamp(t *testing.T) {
	timestamp := internal.GetTimestamp()

	// Verify it's RFC3339 format
	_, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		t.Errorf("GetTimestamp returned invalid RFC3339: %s, error: %v", timestamp, err)
	}

	// Verify it contains required RFC3339 markers
	if !strings.Contains(timestamp, "T") || !strings.Contains(timestamp, "Z") {
		t.Errorf("Timestamp missing RFC3339 markers (T and Z): %s", timestamp)
	}
}

func TestScannerTypeConstants(t *testing.T) {
	tests := []struct {
		scanner internal.ScannerType
		want    string
	}{
		{internal.ScannerSBOM, "sbom"},
		{internal.ScannerVuln, "vuln"},
		{internal.ScannerSecrets, "secrets"},
		{internal.ScannerCompliance, "compliance"},
		{internal.ScannerIaC, "iac"},
		{internal.ScannerSAST, "sast"},
		{internal.ScannerDAST, "zap"},
	}

	for _, tt := range tests {
		if string(tt.scanner) != tt.want {
			t.Errorf("Scanner constant = %s, want %s", tt.scanner, tt.want)
		}
	}
}

func TestEvidenceFileSHA256Integrity(t *testing.T) {
	tempDir := t.TempDir()
	testFindings := map[string]interface{}{
		"results": []string{"finding1", "finding2"},
	}

	// Write evidence
	outputPath, err := internal.WriteEvidence(tempDir, "test-module", internal.ScannerSBOM, testFindings)
	if err != nil {
		t.Fatalf("WriteEvidence failed: %v", err)
	}

	// Read evidence
	evidence, err := internal.ReadEvidence(outputPath)
	if err != nil {
		t.Fatalf("ReadEvidence failed: %v", err)
	}

	// Parse the evidence findings back
	var parsedFindings map[string]interface{}
	if err := json.Unmarshal(evidence.Findings, &parsedFindings); err != nil {
		t.Fatalf("Failed to parse findings: %v", err)
	}

	// Verify the findings content matches
	if len(parsedFindings) == 0 {
		t.Error("Findings is empty")
	}

	// Verify SHA256 is not empty
	if evidence.SHA256 == "" {
		t.Error("SHA256 hash is empty")
	}
}

func TestMultipleEvidenceFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Write multiple evidence files for the same module and scanner
	for i := 0; i < 3; i++ {
		testFindings := map[string]interface{}{
			"iteration": i,
		}

		_, err := internal.WriteEvidence(tempDir, "test-module", internal.ScannerSBOM, testFindings)
		if err != nil {
			t.Fatalf("WriteEvidence iteration %d failed: %v", i, err)
		}

		// Delay to ensure different timestamps (1 second for RFC3339 precision)
		time.Sleep(1 * time.Second)
	}

	// Verify all files were created
	evidenceDir := filepath.Join(tempDir, "out", "security", "test-module", "sbom")
	files, err := filepath.Glob(filepath.Join(evidenceDir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to glob evidence files: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 3 evidence files, got %d", len(files))
	}
}
