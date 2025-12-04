package evidence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindLatestTestRun(t *testing.T) {
	tmpDir := t.TempDir()

	// Create out/test directory structure
	testDir := filepath.Join(tmpDir, "out", "test")
	os.MkdirAll(testDir, 0755)

	// Create test run directories with timestamps
	run1 := filepath.Join(testDir, "2024-01-01T10-00-00")
	run2 := filepath.Join(testDir, "2024-01-02T10-00-00")
	run3 := filepath.Join(testDir, "2024-01-03T10-00-00")

	os.MkdirAll(run1, 0755)
	os.MkdirAll(run2, 0755)
	os.MkdirAll(run3, 0755)

	tests := []struct {
		name     string
		repoRoot string
		wantDir  string
		wantErr  bool
	}{
		{
			name:     "find latest run",
			repoRoot: tmpDir,
			wantDir:  run3,
			wantErr:  false,
		},
		{
			name:     "no test runs",
			repoRoot: t.TempDir(),
			wantDir:  "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindLatestTestRun(tt.repoRoot)

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

			if got != tt.wantDir {
				t.Errorf("FindLatestTestRun() = %s, want %s", got, tt.wantDir)
			}
		})
	}
}

func TestFindTestResultsForModule(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test run directory structure
	testRun := filepath.Join(tmpDir, "out", "test", "2024-01-01T10-00-00")
	billingDir := filepath.Join(testRun, "billing")
	os.MkdirAll(billingDir, 0755)

	// Create test result files
	unitTestFile := filepath.Join(billingDir, "unit-test.json")
	acceptanceFile := filepath.Join(billingDir, "acceptance.cucumber.json")
	os.WriteFile(unitTestFile, []byte(`{}`), 0644)
	os.WriteFile(acceptanceFile, []byte(`[]`), 0644)

	tests := []struct {
		name           string
		repoRoot       string
		moduleName     string
		wantUnitFiles  int
		wantAcceptance int
		wantErr        bool
	}{
		{
			name:           "find billing module results",
			repoRoot:       tmpDir,
			moduleName:     "billing",
			wantUnitFiles:  1,
			wantAcceptance: 1,
			wantErr:        false,
		},
		{
			name:       "module not found",
			repoRoot:   tmpDir,
			moduleName: "nonexistent",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := FindTestResultsForModule(tt.repoRoot, tt.moduleName)

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

			if len(results.UnitTestFiles) != tt.wantUnitFiles {
				t.Errorf("UnitTestFiles count = %d, want %d", len(results.UnitTestFiles), tt.wantUnitFiles)
			}

			if len(results.AcceptanceFiles) != tt.wantAcceptance {
				t.Errorf("AcceptanceFiles count = %d, want %d", len(results.AcceptanceFiles), tt.wantAcceptance)
			}
		})
	}
}

func TestParseCucumberResults(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid cucumber file
	validCucumber := `[{
		"uri": "specs/test.feature",
		"name": "Test Feature",
		"elements": [{
			"type": "scenario",
			"name": "Test scenario",
			"steps": [{"result": {"status": "passed"}}]
		}]
	}]`
	validFile := filepath.Join(tmpDir, "valid.cucumber.json")
	os.WriteFile(validFile, []byte(validCucumber), 0644)

	// Create invalid file
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(invalidFile, []byte(`not json`), 0644)

	tests := []struct {
		name         string
		filePath     string
		wantFeatures int
		wantErr      bool
	}{
		{
			name:         "valid cucumber results",
			filePath:     validFile,
			wantFeatures: 1,
			wantErr:      false,
		},
		{
			name:         "invalid JSON",
			filePath:     invalidFile,
			wantFeatures: 0,
			wantErr:      true,
		},
		{
			name:         "file not found",
			filePath:     filepath.Join(tmpDir, "nonexistent.json"),
			wantFeatures: 0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := ParseCucumberResults(tt.filePath)

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

			if len(results) != tt.wantFeatures {
				t.Errorf("Feature count = %d, want %d", len(results), tt.wantFeatures)
			}
		})
	}
}

func TestExtractControlTags(t *testing.T) {
	tests := []struct {
		name      string
		features  []CucumberFeature
		wantCount int
	}{
		{
			name: "extract control tags",
			features: []CucumberFeature{
				{
					URI: "specs/billing.feature",
					Elements: []CucumberElement{
						{
							Type: "scenario",
							Name: "Process payment",
							Tags: []CucumberTag{
								{Name: "@control:ac-2"},
								{Name: "@control:ia-2"},
								{Name: "@smoke"},
							},
						},
					},
				},
			},
			wantCount: 2,
		},
		{
			name: "no control tags",
			features: []CucumberFeature{
				{
					Elements: []CucumberElement{
						{
							Type: "scenario",
							Tags: []CucumberTag{
								{Name: "@smoke"},
								{Name: "@integration"},
							},
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name:      "empty features",
			features:  []CucumberFeature{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controlMap := ExtractControlTags(tt.features)

			if len(controlMap) != tt.wantCount {
				t.Errorf("Control tag count = %d, want %d", len(controlMap), tt.wantCount)
			}

			// Verify format
			for controlID, locations := range controlMap {
				if controlID == "" {
					t.Error("Found empty control ID")
				}
				if len(locations) == 0 {
					t.Errorf("Control %s has no locations", controlID)
				}
			}
		})
	}
}

func TestCalculateTestSummary(t *testing.T) {
	tests := []struct {
		name    string
		feature CucumberFeature
		want    TestSummary
	}{
		{
			name: "all passed",
			feature: CucumberFeature{
				Elements: []CucumberElement{
					{
						Type: "scenario",
						Steps: []CucumberStep{
							{Result: CucumberResult{Status: "passed"}},
							{Result: CucumberResult{Status: "passed"}},
						},
					},
				},
			},
			want: TestSummary{Total: 1, Passed: 1, Failed: 0, Skipped: 0},
		},
		{
			name: "one failed",
			feature: CucumberFeature{
				Elements: []CucumberElement{
					{
						Type: "scenario",
						Steps: []CucumberStep{
							{Result: CucumberResult{Status: "passed"}},
							{Result: CucumberResult{Status: "failed"}},
						},
					},
				},
			},
			want: TestSummary{Total: 1, Passed: 0, Failed: 1, Skipped: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := []CucumberFeature{tt.feature}
			got := CalculateTestSummary(features)

			if got.Total != tt.want.Total {
				t.Errorf("Total = %d, want %d", got.Total, tt.want.Total)
			}
			if got.Passed != tt.want.Passed {
				t.Errorf("Passed = %d, want %d", got.Passed, tt.want.Passed)
			}
			if got.Failed != tt.want.Failed {
				t.Errorf("Failed = %d, want %d", got.Failed, tt.want.Failed)
			}
		})
	}
}

func TestCollectAllTestEvidence(t *testing.T) {
	results := &TestResults{
		ModuleName: "billing",
		UnitTestFiles: []string{
			"/path/to/unit-test.json",
		},
		AcceptanceFiles: []string{
			"/path/to/acceptance.cucumber.json",
		},
	}

	evidence, err := CollectAllTestEvidence(results)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(evidence) != 2 {
		t.Errorf("Expected 2 evidence items, got %d", len(evidence))
	}

	// Verify evidence types
	foundUnit := false
	foundAcceptance := false
	for _, e := range evidence {
		if e.Type == EvidenceTypeUnitTest {
			foundUnit = true
		}
		if e.Type == EvidenceTypeAcceptanceTest {
			foundAcceptance = true
		}
	}

	if !foundUnit {
		t.Error("Unit test evidence not found")
	}
	if !foundAcceptance {
		t.Error("Acceptance test evidence not found")
	}
}

func TestTestResultsFields(t *testing.T) {
	now := time.Now()
	results := &TestResults{
		ModuleName:       "billing",
		UnitTestFiles:    []string{"/path/to/unit.json"},
		AcceptanceFiles:  []string{"/path/to/acceptance.cucumber.json"},
		TestRunDirectory: "/path/to/test-run",
		LastModified:     now,
	}

	if results.ModuleName != "billing" {
		t.Errorf("ModuleName = %s, want billing", results.ModuleName)
	}

	if len(results.UnitTestFiles) != 1 {
		t.Errorf("UnitTestFiles count = %d, want 1", len(results.UnitTestFiles))
	}

	if len(results.AcceptanceFiles) != 1 {
		t.Errorf("AcceptanceFiles count = %d, want 1", len(results.AcceptanceFiles))
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
