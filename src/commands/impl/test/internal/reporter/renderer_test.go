package reporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRenderer_Render(t *testing.T) {
	tests := []struct {
		name         string
		templateData TestSuiteReportData
		wantContains []string
		wantErr      bool
	}{
		{
			name: "renders basic suite report",
			templateData: BuildReportData(
				"Integration Tests", "integration",
				time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
				time.Date(2025, 1, 1, 10, 5, 30, 0, time.UTC),
				100, 90, 10,
				80, 75, 5, 0,
				10, 9, 1,
				nil,
				"/tmp/test-results",
			),
			wantContains: []string{
				"# Test Suite Report: Integration Tests",
				"**Suite**: integration",
				"**Duration**: 330.0s",
				"Tests Discovered   | 100",
				"Production Tests   | 90",
				"Framework Tests Excluded | 10",
				"Tests Total**    | **80**",
				"Tests Passed       | 75 ✅",
				"Tests Failed       | 5",
				"Test Pass Rate     | 93.8%",
				"Packages Passed    | 9 ✅",
				"Packages Failed    | 1",
			},
			wantErr: false,
		},
		{
			name: "renders suite with modules",
			templateData: TestSuiteReportData{
				SuiteName:     "Unit Tests",
				SuiteMoniker:  "unit",
				StartTime:     "2025-01-01 10:00:00",
				EndTime:       "2025-01-01 10:02:00",
				Duration:      120.0,
				Status:        "✅ PASSED",
				GeneratedAt:   "2025-01-01 10:02:00",
				TestsTotal:    50,
				TestsPassed:   50,
				TestPassRate:  100.0,
				PackagesTotal: 5,
				PackagesPassed: 5,
				PackagePassRate: 100.0,
				Modules: []ModuleReportData{
					{
						ModuleName:    "src-commands",
						TestResultDir: "src-commands",
						Features: []FeatureReportData{
							{
								Name:          "Template Management",
								URI:           "../../../specs/src-commands/templates/specification.feature",
								NormalizedURI: "specs/src-commands/templates/specification.feature",
								Tags:          "@smoke @templates",
								TestCount:     10,
								PassedCount:   10,
								Status:        "🟢 Passed",
							},
						},
						UnitTests: []UnitTestReportData{
							{
								Package:   "impl/templates",
								TestCount: 15,
								Status:    "🟢 Passed",
							},
						},
						TotalScenarios: 10,
						TotalUnitTests: 15,
					},
				},
				ResultsDirectory: "/tmp/test-results",
			},
			wantContains: []string{
				"## Module: src-commands",
				"[Test results](./src-commands)",
				"### Specifications (BDD Tests)",
				"[Template Management](../../specs/src-commands/templates/specification.feature)",
				"@smoke @templates",
				"### Unit Tests (Implementation Validation)",
				"impl/templates",
				"**Total: 10 scenarios + 15 unit tests = 25 tests**",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test output
			tempDir, err := os.MkdirTemp("", "reporter-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Use the actual template file from templates/test-reports
			// Get repository root to find template
			workingDir, _ := os.Getwd()
			// Navigate up from src/commands/impl/test/internal/reporter to root
			repoRoot := filepath.Join(workingDir, "..", "..", "..", "..", "..", "..")
			templatePath := filepath.Join(repoRoot, "templates", "test-reports", "suite-summary.md")

			outputPath := filepath.Join(tempDir, "test-suite-summary.md")

			// Create renderer and render
			renderer := NewRenderer(templatePath, outputPath, tt.templateData)
			err = renderer.Render()

			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Read rendered output
			output, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			outputStr := string(output)

			// Check for expected content
			for _, want := range tt.wantContains {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Rendered output missing expected content:\nWant: %s\n\nGot:\n%s", want, outputStr)
				}
			}
		})
	}
}

func TestNormalizeSpecPath(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want string
	}{
		{
			name: "removes leading ../",
			uri:  "../../../specs/src-commands/templates/specification.feature",
			want: "specs/src-commands/templates/specification.feature",
		},
		{
			name: "adds specs/ prefix if missing",
			uri:  "src-commands/templates/specification.feature",
			want: "specs/src-commands/templates/specification.feature",
		},
		{
			name: "leaves correct path unchanged",
			uri:  "specs/src-cli/init/specification.feature",
			want: "specs/src-cli/init/specification.feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeSpecPath(tt.uri)
			if got != tt.want {
				t.Errorf("NormalizeSpecPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildReportData(t *testing.T) {
	startTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 1, 1, 10, 5, 30, 0, time.UTC)

	tests := []struct {
		name                   string
		testsTotal             int
		testsPassed            int
		packagesTotal          int
		packagesPassed         int
		wantStatus             string
		wantTestPassRate       float64
		wantPackagePassRate    float64
		wantDuration           float64
	}{
		{
			name:                "all tests passed",
			testsTotal:          100,
			testsPassed:         100,
			packagesTotal:       10,
			packagesPassed:      10,
			wantStatus:          "✅ PASSED",
			wantTestPassRate:    100.0,
			wantPackagePassRate: 100.0,
			wantDuration:        330.0,
		},
		{
			name:                "some tests failed",
			testsTotal:          100,
			testsPassed:         95,
			packagesTotal:       10,
			packagesPassed:      9,
			wantStatus:          "❌ FAILED",
			wantTestPassRate:    95.0,
			wantPackagePassRate: 90.0,
			wantDuration:        330.0,
		},
		{
			name:                "no tests",
			testsTotal:          0,
			testsPassed:         0,
			packagesTotal:       0,
			packagesPassed:      0,
			wantStatus:          "✅ PASSED",
			wantTestPassRate:    0.0,
			wantPackagePassRate: 0.0,
			wantDuration:        330.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := BuildReportData(
				"Test Suite", "test",
				startTime, endTime,
				tt.testsTotal, tt.testsTotal, 0,
				tt.testsTotal, tt.testsPassed, tt.testsTotal-tt.testsPassed, 0,
				tt.packagesTotal, tt.packagesPassed, tt.packagesTotal-tt.packagesPassed,
				nil,
				"/tmp/results",
			)

			if data.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", data.Status, tt.wantStatus)
			}
			if data.TestPassRate != tt.wantTestPassRate {
				t.Errorf("TestPassRate = %v, want %v", data.TestPassRate, tt.wantTestPassRate)
			}
			if data.PackagePassRate != tt.wantPackagePassRate {
				t.Errorf("PackagePassRate = %v, want %v", data.PackagePassRate, tt.wantPackagePassRate)
			}
			if data.Duration != tt.wantDuration {
				t.Errorf("Duration = %v, want %v", data.Duration, tt.wantDuration)
			}
		})
	}
}
