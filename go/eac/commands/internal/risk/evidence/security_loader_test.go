package evidence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindLatestSecurityScan(t *testing.T) {
	tmpDir := t.TempDir()

	// Create out/security/billing/vuln directory structure
	// (using default security path from contract defaults)
	vulnDir := filepath.Join(tmpDir, "out", "security", "billing", "vuln")
	os.MkdirAll(vulnDir, 0755)

	// Create security scan files with timestamps
	scan1 := filepath.Join(vulnDir, "2024-01-01T10-00-00Z.json")
	scan2 := filepath.Join(vulnDir, "2024-01-02T10-00-00Z.json")
	scan3 := filepath.Join(vulnDir, "2024-01-03T10-00-00Z.json")

	os.WriteFile(scan1, []byte(`{"Results":[]}`), 0644)
	os.WriteFile(scan2, []byte(`{"Results":[]}`), 0644)
	os.WriteFile(scan3, []byte(`{"Results":[]}`), 0644)

	tests := []struct {
		name        string
		repoRoot    string
		moduleName  string
		scannerType ScannerType
		wantFile    string
		wantErr     bool
	}{
		{
			name:        "find latest vuln scan",
			repoRoot:    tmpDir,
			moduleName:  "billing",
			scannerType: ScannerVuln,
			wantFile:    scan3,
			wantErr:     false,
		},
		{
			name:        "scanner not found",
			repoRoot:    tmpDir,
			moduleName:  "billing",
			scannerType: ScannerSBOM,
			wantFile:    "",
			wantErr:     true,
		},
		{
			name:        "module not found",
			repoRoot:    tmpDir,
			moduleName:  "nonexistent",
			scannerType: ScannerVuln,
			wantFile:    "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindLatestSecurityScan(tt.repoRoot, tt.moduleName, tt.scannerType)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if got != tt.wantFile {
				t.Errorf("FindLatestSecurityScan() = %s, want %s", got, tt.wantFile)
			}
		})
	}
}

func TestFindSecurityResultsForModule(t *testing.T) {
	tmpDir := t.TempDir()

	// Create security scan files for billing module
	// (using default security path from contract defaults)
	vulnDir := filepath.Join(tmpDir, "out", "security", "billing", "vuln")
	sbomDir := filepath.Join(tmpDir, "out", "security", "billing", "sbom")
	os.MkdirAll(vulnDir, 0755)
	os.MkdirAll(sbomDir, 0755)

	vulnFile := filepath.Join(vulnDir, "2024-01-01T10-00-00Z.json")
	sbomFile := filepath.Join(sbomDir, "2024-01-01T10-00-00Z.json")
	os.WriteFile(vulnFile, []byte(`{"Results":[]}`), 0644)
	os.WriteFile(sbomFile, []byte(`{"Results":[]}`), 0644)

	tests := []struct {
		name       string
		repoRoot   string
		moduleName string
		wantVuln   bool
		wantSBOM   bool
		wantErr    bool
	}{
		{
			name:       "find billing module scans",
			repoRoot:   tmpDir,
			moduleName: "billing",
			wantVuln:   true,
			wantSBOM:   true,
			wantErr:    false,
		},
		{
			name:       "module with no scans",
			repoRoot:   tmpDir,
			moduleName: "nonexistent",
			wantVuln:   false,
			wantSBOM:   false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := FindSecurityResultsForModule(tt.repoRoot, tt.moduleName)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if results == nil {
				t.Fatal("Results is nil")
			}

			hasVuln := results.VulnFile != ""
			if hasVuln != tt.wantVuln {
				t.Errorf("VulnFile present = %v, want %v", hasVuln, tt.wantVuln)
			}

			hasSBOM := results.SBOMFile != ""
			if hasSBOM != tt.wantSBOM {
				t.Errorf("SBOMFile present = %v, want %v", hasSBOM, tt.wantSBOM)
			}
		})
	}
}

func TestLoadSecurityEvidence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid security evidence file
	validEvidence := `{
		"module": "billing",
		"scanner": "trivy",
		"timestamp": "2024-01-01T10:00:00Z",
		"sha256": "abc123",
		"findings": {"Results": []}
	}`
	validFile := filepath.Join(tmpDir, "valid.json")
	os.WriteFile(validFile, []byte(validEvidence), 0644)

	// Create invalid file
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(invalidFile, []byte(`not json`), 0644)

	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "valid evidence",
			filePath: validFile,
			wantErr:  false,
		},
		{
			name:     "invalid JSON",
			filePath: invalidFile,
			wantErr:  true,
		},
		{
			name:     "file not found",
			filePath: filepath.Join(tmpDir, "nonexistent.json"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evidence, err := LoadSecurityEvidence(tt.filePath)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if evidence == nil {
				t.Fatal("Evidence is nil")
			}

			if evidence.Module != "billing" {
				t.Errorf("Module = %s, want billing", evidence.Module)
			}
		})
	}
}

func TestParseVulnerabilitySummary(t *testing.T) {
	tests := []struct {
		name         string
		findings     string
		wantCritical int
		wantHigh     int
		wantMedium   int
		wantLow      int
	}{
		{
			name: "Trivy format with vulnerabilities",
			findings: `{
				"Results": [{
					"Vulnerabilities": [
						{"Severity": "CRITICAL"},
						{"Severity": "HIGH"},
						{"Severity": "HIGH"},
						{"Severity": "MEDIUM"},
						{"Severity": "LOW"}
					]
				}]
			}`,
			wantCritical: 1,
			wantHigh:     2,
			wantMedium:   1,
			wantLow:      1,
		},
		{
			name: "case insensitive",
			findings: `{
				"Results": [{
					"Vulnerabilities": [
						{"Severity": "high"},
						{"Severity": "High"},
						{"Severity": "HIGH"}
					]
				}]
			}`,
			wantCritical: 0,
			wantHigh:     3,
			wantMedium:   0,
			wantLow:      0,
		},
		{
			name:         "empty results",
			findings:     `{"Results": []}`,
			wantCritical: 0,
			wantHigh:     0,
			wantMedium:   0,
			wantLow:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evidence := &SecurityEvidenceFile{
				Module:   "test",
				Scanner:  "trivy",
				Findings: json.RawMessage(tt.findings),
			}

			summary, err := ParseVulnerabilitySummary(evidence)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if summary == nil {
				t.Fatal("Summary is nil")
			}

			if summary.Critical != tt.wantCritical {
				t.Errorf("Critical = %d, want %d", summary.Critical, tt.wantCritical)
			}

			if summary.High != tt.wantHigh {
				t.Errorf("High = %d, want %d", summary.High, tt.wantHigh)
			}

			if summary.Medium != tt.wantMedium {
				t.Errorf("Medium = %d, want %d", summary.Medium, tt.wantMedium)
			}

			if summary.Low != tt.wantLow {
				t.Errorf("Low = %d, want %d", summary.Low, tt.wantLow)
			}

			expectedTotal := tt.wantCritical + tt.wantHigh + tt.wantMedium + tt.wantLow
			if summary.Total != expectedTotal {
				t.Errorf("Total = %d, want %d", summary.Total, expectedTotal)
			}
		})
	}
}

func TestIsEvidenceFresh(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file and modify its timestamp
	file := filepath.Join(tmpDir, "test.json")
	os.WriteFile(file, []byte(`{}`), 0644)

	// Wait a bit to ensure file is not brand new
	time.Sleep(2 * time.Millisecond)

	tests := []struct {
		name   string
		path   string
		maxAge time.Duration
		want   bool
	}{
		{
			name:   "fresh file",
			path:   file,
			maxAge: 24 * time.Hour,
			want:   true,
		},
		{
			name:   "old file",
			path:   file,
			maxAge: 1 * time.Millisecond,
			want:   false,
		},
		{
			name:   "missing file",
			path:   filepath.Join(tmpDir, "nonexistent.json"),
			maxAge: 24 * time.Hour,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEvidenceFresh(tt.path, tt.maxAge)
			if got != tt.want {
				t.Errorf("IsEvidenceFresh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEvidenceAge(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	file := filepath.Join(tmpDir, "test.json")
	os.WriteFile(file, []byte(`{}`), 0644)

	// Get age
	age, err := GetEvidenceAge(file)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if age < 0 {
		t.Error("Age should be positive")
	}

	// Missing file should error
	_, err = GetEvidenceAge(filepath.Join(tmpDir, "nonexistent.json"))
	if err == nil {
		t.Error("Expected error for missing file")
	}
}

func TestCollectAllSecurityEvidence(t *testing.T) {
	results := &SecurityResults{
		ModuleName:  "billing",
		VulnFile:    "/path/to/vuln.json",
		SBOMFile:    "/path/to/sbom.json",
		SecretsFile: "/path/to/secrets.json",
	}

	evidence, err := CollectAllSecurityEvidence(results)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(evidence) != 3 {
		t.Errorf("Expected 3 evidence items, got %d", len(evidence))
	}

	// Verify evidence types
	foundVuln := false
	foundSBOM := false
	foundSecrets := false
	for _, e := range evidence {
		switch e.Type {
		case EvidenceTypeVulnerabilityScan:
			foundVuln = true
		case EvidenceTypeSBOM:
			foundSBOM = true
		case EvidenceTypeSecretsScan:
			foundSecrets = true
		}
	}

	if !foundVuln {
		t.Error("Vulnerability scan evidence not found")
	}
	if !foundSBOM {
		t.Error("SBOM evidence not found")
	}
	if !foundSecrets {
		t.Error("Secrets scan evidence not found")
	}
}

func TestScannerTypes(t *testing.T) {
	scanners := []ScannerType{
		ScannerSBOM,
		ScannerVuln,
		ScannerSecrets,
		ScannerCompliance,
		ScannerIaC,
		ScannerSAST,
		ScannerDAST,
	}

	for _, scanner := range scanners {
		if string(scanner) == "" {
			t.Errorf("Scanner type has empty string value")
		}
	}
}

func TestSecurityResultsFields(t *testing.T) {
	now := time.Now()
	results := &SecurityResults{
		ModuleName:     "billing",
		VulnFile:       "/path/to/vuln.json",
		SBOMFile:       "/path/to/sbom.json",
		SecretsFile:    "/path/to/secrets.json",
		SASTFile:       "/path/to/sast.json",
		ComplianceFile: "/path/to/compliance.json",
		IaCFile:        "/path/to/iac.json",
		ZAPFile:        "/path/to/zap.json",
		LastModified:   now,
	}

	if results.ModuleName != "billing" {
		t.Errorf("ModuleName = %s, want billing", results.ModuleName)
	}

	// Can marshal to JSON
	data, err := json.Marshal(results)
	if err != nil {
		t.Errorf("Failed to marshal: %v", err)
	}
	if len(data) == 0 {
		t.Error("Marshaled data is empty")
	}
}

func TestEvidenceCollection(t *testing.T) {
	now := time.Now()
	collection := &EvidenceCollection{
		Module: "billing",
		TestResults: &TestResults{
			ModuleName:      "billing",
			UnitTestFiles:   []string{"/path/to/unit.json"},
			AcceptanceFiles: []string{"/path/to/acceptance.cucumber.json"},
			LastModified:    now.Add(-1 * time.Hour),
		},
		SecurityResults: &SecurityResults{
			ModuleName:   "billing",
			VulnFile:     "/path/to/vuln.json",
			LastModified: now,
		},
		CollectedAt: now,
	}

	if !collection.HasAnyEvidence() {
		t.Error("Should have evidence")
	}

	if !collection.HasTestEvidence() {
		t.Error("Should have test evidence")
	}

	if !collection.HasSecurityEvidence() {
		t.Error("Should have security evidence")
	}

	latestMod := collection.LatestModTime()
	if latestMod.Before(now.Add(-2 * time.Hour)) {
		t.Error("Latest mod time should be recent")
	}
}

func TestEvidenceAgePolicy(t *testing.T) {
	policy := DefaultEvidenceAgePolicy()

	if policy.MaxAge != 24*time.Hour {
		t.Errorf("MaxAge = %v, want 24h", policy.MaxAge)
	}
}
