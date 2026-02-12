package riskassess

import (
	"strings"
	"testing"
	"time"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/evidence"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/oscal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetermineControlStatusFromEvidence(t *testing.T) {
	tests := []struct {
		name     string
		testEv   *evidence.ControlTestEvidence
		ec       *evidence.EvidenceCollection
		expected string
	}{
		{
			name:     "nil test evidence returns not-satisfied",
			testEv:   nil,
			ec:       &evidence.EvidenceCollection{},
			expected: oscal.StateNotSatisfied,
		},
		{
			name: "no evidence flag returns not-satisfied",
			testEv: &evidence.ControlTestEvidence{
				HasEvidence: false,
			},
			ec:       &evidence.EvidenceCollection{},
			expected: oscal.StateNotSatisfied,
		},
		{
			name: "failed tests returns not-satisfied",
			testEv: &evidence.ControlTestEvidence{
				HasEvidence: true,
				TotalTests:  3,
				PassedTests: 2,
				FailedTests: 1,
			},
			ec:       &evidence.EvidenceCollection{},
			expected: oscal.StateNotSatisfied,
		},
		{
			name: "critical vulnerabilities returns not-satisfied",
			testEv: &evidence.ControlTestEvidence{
				HasEvidence: true,
				TotalTests:  3,
				PassedTests: 3,
				FailedTests: 0,
			},
			ec: &evidence.EvidenceCollection{
				VulnSummary: &evidence.VulnerabilitySummary{
					Critical: 1,
					High:     0,
					Medium:   0,
					Low:      0,
				},
			},
			expected: oscal.StateNotSatisfied,
		},
		{
			name: "high vulnerabilities returns not-satisfied",
			testEv: &evidence.ControlTestEvidence{
				HasEvidence: true,
				TotalTests:  3,
				PassedTests: 3,
				FailedTests: 0,
			},
			ec: &evidence.EvidenceCollection{
				VulnSummary: &evidence.VulnerabilitySummary{
					Critical: 0,
					High:     2,
					Medium:   0,
					Low:      0,
				},
			},
			expected: oscal.StateNotSatisfied,
		},
		{
			name: "all tests passed no vulns returns satisfied",
			testEv: &evidence.ControlTestEvidence{
				HasEvidence: true,
				TotalTests:  5,
				PassedTests: 5,
				FailedTests: 0,
			},
			ec:       &evidence.EvidenceCollection{},
			expected: oscal.StateSatisfied,
		},
		{
			name: "all tests passed with medium vulns returns satisfied",
			testEv: &evidence.ControlTestEvidence{
				HasEvidence: true,
				TotalTests:  5,
				PassedTests: 5,
				FailedTests: 0,
			},
			ec: &evidence.EvidenceCollection{
				VulnSummary: &evidence.VulnerabilitySummary{
					Critical: 0,
					High:     0,
					Medium:   3,
					Low:      5,
				},
			},
			expected: oscal.StateSatisfied,
		},
		{
			name: "all tests passed with nil vuln summary returns satisfied",
			testEv: &evidence.ControlTestEvidence{
				HasEvidence: true,
				TotalTests:  2,
				PassedTests: 2,
				FailedTests: 0,
			},
			ec: &evidence.EvidenceCollection{
				VulnSummary: nil,
			},
			expected: oscal.StateSatisfied,
		},
		{
			name: "evidence exists but no passed tests returns not-satisfied",
			testEv: &evidence.ControlTestEvidence{
				HasEvidence:  true,
				TotalTests:   2,
				PassedTests:  0,
				FailedTests:  0,
				SkippedTests: 2,
			},
			ec:       &evidence.EvidenceCollection{},
			expected: oscal.StateNotSatisfied,
		},
		{
			name: "both failed tests and critical vulns returns not-satisfied",
			testEv: &evidence.ControlTestEvidence{
				HasEvidence: true,
				TotalTests:  5,
				PassedTests: 3,
				FailedTests: 2,
			},
			ec: &evidence.EvidenceCollection{
				VulnSummary: &evidence.VulnerabilitySummary{
					Critical: 1,
					High:     0,
				},
			},
			expected: oscal.StateNotSatisfied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineControlStatusFromEvidence("ac-1", tt.testEv, tt.ec)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildRemarksFromEvidence(t *testing.T) {
	tests := []struct {
		name           string
		controlID      string
		testEv         *evidence.ControlTestEvidence
		wantContains   []string
		wantNotContain []string
		wantExact      string
	}{
		{
			name:      "nil test evidence",
			controlID: "ac-1",
			testEv:    nil,
			wantExact: "No test coverage found for this control.",
		},
		{
			name:      "no evidence flag",
			controlID: "ac-2",
			testEv: &evidence.ControlTestEvidence{
				HasEvidence: false,
			},
			wantExact: "No test coverage found for this control.",
		},
		{
			name:      "all tests passed",
			controlID: "ac-3",
			testEv: &evidence.ControlTestEvidence{
				HasEvidence: true,
				TotalTests:  2,
				PassedTests: 2,
				Tests: []evidence.TestEvidence{
					{
						ScenarioName: "User can log in",
						FeaturePath:  "features/auth.feature",
						Status:       "passed",
					},
					{
						ScenarioName: "User can log out",
						FeaturePath:  "features/auth.feature",
						Status:       "passed",
					},
				},
			},
			wantContains: []string{
				"Control ac-3 is covered by 2 test(s)",
				"Passed:",
				"User can log in",
				"User can log out",
			},
			wantNotContain: []string{
				"Failed:",
				"Not Run/Skipped:",
			},
		},
		{
			name:      "mixed test results",
			controlID: "au-1",
			testEv: &evidence.ControlTestEvidence{
				HasEvidence: true,
				TotalTests:  3,
				PassedTests: 1,
				FailedTests: 1,
				Tests: []evidence.TestEvidence{
					{
						ScenarioName: "Logs are generated",
						FeaturePath:  "features/audit.feature",
						Status:       "passed",
					},
					{
						ScenarioName: "Logs are tamper-proof",
						FeaturePath:  "features/audit.feature",
						Status:       "failed",
					},
					{
						ScenarioName: "Log rotation works",
						FeaturePath:  "features/audit.feature",
						Status:       "skipped",
					},
				},
			},
			wantContains: []string{
				"Control au-1 is covered by 3 test(s)",
				"Passed:",
				"Logs are generated",
				"Failed:",
				"Logs are tamper-proof",
				"Not Run/Skipped:",
				"Log rotation works",
			},
		},
		{
			name:      "only failed tests",
			controlID: "sc-1",
			testEv: &evidence.ControlTestEvidence{
				HasEvidence: true,
				TotalTests:  1,
				FailedTests: 1,
				Tests: []evidence.TestEvidence{
					{
						ScenarioName: "Encryption at rest",
						FeaturePath:  "features/security.feature",
						Status:       "failed",
					},
				},
			},
			wantContains: []string{
				"Control sc-1 is covered by 1 test(s)",
				"Failed:",
				"Encryption at rest",
			},
			wantNotContain: []string{
				"Passed:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildRemarksFromEvidence(tt.controlID, tt.testEv)

			if tt.wantExact != "" {
				assert.Equal(t, tt.wantExact, result)
				return
			}

			for _, s := range tt.wantContains {
				assert.Contains(t, result, s, "expected remarks to contain %q", s)
			}
			for _, s := range tt.wantNotContain {
				assert.NotContains(t, result, s, "expected remarks not to contain %q", s)
			}
		})
	}
}

func TestCreateObservationsForModule(t *testing.T) {
	config := &AssessConfig{
		WorkspaceRoot: "/workspace",
		OutputDir:     "/workspace/out/risk",
	}

	t.Run("no evidence produces empty observations", func(t *testing.T) {
		ec := &evidence.EvidenceCollection{
			Module: "empty-module",
		}

		obs := createObservationsForModule(config, "empty-module", ec)
		require.NotNil(t, obs)
		assert.Empty(t, *obs)
	})

	t.Run("test manifest data creates test observation", func(t *testing.T) {
		ec := &evidence.EvidenceCollection{
			Module: "test-module",
			TestManifestData: &evidence.TestManifestData{
				TestAgent: "gotest",
				TestID:    "run-001",
				GitCommit: "abc123",
				BuildID:   "build-42",
				Artifacts: []evidence.TestArtifactData{
					{Type: "report", Name: "report.json", Path: "/out/test/report.json"},
					{Type: "coverage", Name: "coverage.out", Path: "/out/test/coverage.out"},
					{Type: "log", Name: "test.log", Path: "/out/test/test.log"},
					{Type: "other", Name: "data.txt", Path: "/out/test/data.txt"},
				},
			},
			TestSummary: &evidence.TestSummary{
				Total:   10,
				Passed:  8,
				Failed:  1,
				Skipped: 1,
			},
		}

		obs := createObservationsForModule(config, "test-module", ec)
		require.NotNil(t, obs)
		require.Len(t, *obs, 1)

		testObs := (*obs)[0]
		assert.Equal(t, "Test Results", testObs.Title)
		assert.Contains(t, testObs.Description, "test-module")
		assert.Equal(t, []string{oscal.MethodTestAutomated}, testObs.Methods)

		// Check relevant evidence (only report, coverage, log types)
		require.NotNil(t, testObs.RelevantEvidence)
		assert.Len(t, *testObs.RelevantEvidence, 3, "should include report, coverage, and log artifacts only")

		// Check properties
		require.NotNil(t, testObs.Props)
		props := *testObs.Props
		propMap := make(map[string]string)
		for _, p := range props {
			propMap[p.Name] = p.Value
		}
		assert.Equal(t, "gotest", propMap["test-agent"])
		assert.Equal(t, "run-001", propMap["test-id"])
		assert.Equal(t, "abc123", propMap["git-commit"])
		assert.Equal(t, "build-42", propMap["build-id"])
		assert.Equal(t, "10", propMap["total-tests"])
		assert.Equal(t, "8", propMap["passed-tests"])
		assert.Equal(t, "1", propMap["failed-tests"])
		assert.Equal(t, "1", propMap["skipped-tests"])
	})

	t.Run("security results create security observation", func(t *testing.T) {
		ec := &evidence.EvidenceCollection{
			Module: "secure-module",
			SecurityResults: &evidence.SecurityResults{
				VulnFile: "/workspace/out/scan/secure-module/vulns.json",
				SBOMFile: "/workspace/out/scan/secure-module/sbom.json",
			},
			VulnSummary: &evidence.VulnerabilitySummary{
				Critical: 0,
				High:     2,
				Medium:   5,
				Low:      10,
			},
			SBOMSummary: &evidence.SBOMSummary{
				TotalComponents: 42,
			},
		}

		obs := createObservationsForModule(config, "secure-module", ec)
		require.NotNil(t, obs)
		require.Len(t, *obs, 1)

		secObs := (*obs)[0]
		assert.Equal(t, "Security Scan Results", secObs.Title)
		assert.Contains(t, secObs.Description, "secure-module")

		// Check relevant evidence references
		require.NotNil(t, secObs.RelevantEvidence)
		assert.Len(t, *secObs.RelevantEvidence, 2, "should include vuln and sbom files")

		// Check vulnerability props
		require.NotNil(t, secObs.Props)
		propMap := make(map[string]string)
		for _, p := range *secObs.Props {
			propMap[p.Name] = p.Value
		}
		assert.Equal(t, "0", propMap["critical-vulns"])
		assert.Equal(t, "2", propMap["high-vulns"])
		assert.Equal(t, "5", propMap["medium-vulns"])
		assert.Equal(t, "10", propMap["low-vulns"])
		assert.Equal(t, "42", propMap["sbom-components"])
	})

	t.Run("both test and security evidence creates two observations", func(t *testing.T) {
		ec := &evidence.EvidenceCollection{
			Module: "full-module",
			TestManifestData: &evidence.TestManifestData{
				TestAgent: "gotest",
				TestID:    "run-002",
			},
			TestSummary: &evidence.TestSummary{
				Total:  5,
				Passed: 5,
			},
			SecurityResults: &evidence.SecurityResults{
				VulnFile: "/workspace/out/scan/full-module/vulns.json",
			},
			VulnSummary: &evidence.VulnerabilitySummary{
				Critical: 0,
				High:     0,
				Medium:   1,
				Low:      2,
			},
		}

		obs := createObservationsForModule(config, "full-module", ec)
		require.NotNil(t, obs)
		assert.Len(t, *obs, 2)
		assert.Equal(t, "Test Results", (*obs)[0].Title)
		assert.Equal(t, "Security Scan Results", (*obs)[1].Title)
	})

	t.Run("test manifest with suites adds suite-level props", func(t *testing.T) {
		ec := &evidence.EvidenceCollection{
			Module: "suite-module",
			TestManifestData: &evidence.TestManifestData{
				TestAgent: "gotest",
				TestID:    "run-003",
				Suites: map[string]evidence.SuiteResultData{
					"unit": {
						Total:  10,
						Passed: 9,
						Failed: 1,
					},
					"integration": {
						Total:   5,
						Passed:  5,
						Failed:  0,
						Skipped: 0,
					},
				},
			},
			TestSummary: &evidence.TestSummary{
				Total:  15,
				Passed: 14,
				Failed: 1,
			},
		}

		obs := createObservationsForModule(config, "suite-module", ec)
		require.NotNil(t, obs)
		require.Len(t, *obs, 1)

		propMap := make(map[string]string)
		for _, p := range *(*obs)[0].Props {
			propMap[p.Name] = p.Value
		}

		assert.Equal(t, "10", propMap["suite-unit-total"])
		assert.Equal(t, "9", propMap["suite-unit-passed"])
		assert.Equal(t, "1", propMap["suite-unit-failed"])
		assert.Equal(t, "5", propMap["suite-integration-total"])
		assert.Equal(t, "5", propMap["suite-integration-passed"])
		assert.Equal(t, "0", propMap["suite-integration-failed"])
	})

	t.Run("test manifest without git commit or build id omits those props", func(t *testing.T) {
		ec := &evidence.EvidenceCollection{
			Module: "minimal-module",
			TestManifestData: &evidence.TestManifestData{
				TestAgent: "pytest",
				TestID:    "run-004",
				GitCommit: "",
				BuildID:   "",
			},
		}

		obs := createObservationsForModule(config, "minimal-module", ec)
		require.NotNil(t, obs)
		require.Len(t, *obs, 1)

		propMap := make(map[string]string)
		for _, p := range *(*obs)[0].Props {
			propMap[p.Name] = p.Value
		}
		assert.Equal(t, "pytest", propMap["test-agent"])
		assert.Equal(t, "run-004", propMap["test-id"])
		_, hasGitCommit := propMap["git-commit"]
		_, hasBuildID := propMap["build-id"]
		assert.False(t, hasGitCommit, "should not include empty git-commit prop")
		assert.False(t, hasBuildID, "should not include empty build-id prop")
	})
}

func TestCreateFindingsForModule(t *testing.T) {
	config := &AssessConfig{
		WorkspaceRoot: "/workspace",
	}

	t.Run("empty control IDs returns empty findings", func(t *testing.T) {
		findings := createFindingsForModule(config, "test-module", []string{}, nil, &evidence.EvidenceCollection{}, nil)
		require.NotNil(t, findings)
		assert.Empty(t, *findings)
	})

	t.Run("controls with no test evidence are not-satisfied", func(t *testing.T) {
		controlIDs := []string{"ac-1", "ac-2"}
		ec := &evidence.EvidenceCollection{}
		controlTestEvidence := map[string]*evidence.ControlTestEvidence{}

		findings := createFindingsForModule(config, "test-module", controlIDs, controlTestEvidence, ec, nil)
		require.NotNil(t, findings)
		assert.Len(t, *findings, 2)

		for _, f := range *findings {
			assert.Equal(t, oscal.StateNotSatisfied, f.Target.Status.State)
			// Check "no tests" property
			require.NotNil(t, f.Props)
			found := false
			for _, p := range *f.Props {
				if p.Name == "test-coverage" && p.Value == "no tests" {
					found = true
				}
			}
			assert.True(t, found, "expected test-coverage property with value 'no tests'")
		}
	})

	t.Run("controls with passing tests are satisfied", func(t *testing.T) {
		controlIDs := []string{"ac-1"}
		ec := &evidence.EvidenceCollection{}
		controlTestEvidence := map[string]*evidence.ControlTestEvidence{
			"ac-1": {
				ControlID:   "ac-1",
				HasEvidence: true,
				TotalTests:  3,
				PassedTests: 3,
				FailedTests: 0,
				Tests: []evidence.TestEvidence{
					{ScenarioName: "Test A", FeaturePath: "a.feature", Status: "passed"},
					{ScenarioName: "Test B", FeaturePath: "b.feature", Status: "passed"},
					{ScenarioName: "Test C", FeaturePath: "c.feature", Status: "passed"},
				},
			},
		}

		findings := createFindingsForModule(config, "test-module", controlIDs, controlTestEvidence, ec, nil)
		require.NotNil(t, findings)
		require.Len(t, *findings, 1)

		f := (*findings)[0]
		assert.Equal(t, oscal.StateSatisfied, f.Target.Status.State)
		assert.Equal(t, "ac-1", f.Target.TargetId)
		assert.Equal(t, oscal.TargetTypeControlID, f.Target.Type)
		assert.Equal(t, "Control AC-1 Assessment", f.Title)
		assert.Contains(t, f.Description, "ac-1")

		// Check test-coverage property
		require.NotNil(t, f.Props)
		propMap := make(map[string]string)
		for _, p := range *f.Props {
			propMap[p.Name] = p.Value
		}
		assert.Contains(t, propMap["test-coverage"], "3 tests")
		assert.Contains(t, propMap["test-coverage"], "3 passed")

		// Check remarks
		assert.Contains(t, f.Remarks, "ac-1")
		assert.Contains(t, f.Remarks, "Test A")
	})

	t.Run("findings link to observations when provided", func(t *testing.T) {
		controlIDs := []string{"ac-1"}
		ec := &evidence.EvidenceCollection{}
		controlTestEvidence := map[string]*evidence.ControlTestEvidence{
			"ac-1": {
				HasEvidence: true,
				TotalTests:  1,
				PassedTests: 1,
				Tests: []evidence.TestEvidence{
					{ScenarioName: "Test A", Status: "passed"},
				},
			},
		}

		observations := &[]oscalTypes.Observation{
			{
				UUID:  "obs-uuid-1",
				Title: "Test Results",
			},
			{
				UUID:  "obs-uuid-2",
				Title: "Security Scan Results",
			},
		}

		findings := createFindingsForModule(config, "test-module", controlIDs, controlTestEvidence, ec, observations)
		require.NotNil(t, findings)
		require.Len(t, *findings, 1)

		f := (*findings)[0]
		require.NotNil(t, f.RelatedObservations)
		assert.Len(t, *f.RelatedObservations, 2)
		assert.Equal(t, "obs-uuid-1", (*f.RelatedObservations)[0].ObservationUuid)
		assert.Equal(t, "obs-uuid-2", (*f.RelatedObservations)[1].ObservationUuid)
	})

	t.Run("mixed control statuses", func(t *testing.T) {
		controlIDs := []string{"ac-1", "ac-2", "ac-3"}
		ec := &evidence.EvidenceCollection{}
		controlTestEvidence := map[string]*evidence.ControlTestEvidence{
			"ac-1": {
				HasEvidence: true,
				TotalTests:  2,
				PassedTests: 2,
				Tests: []evidence.TestEvidence{
					{ScenarioName: "Test A", Status: "passed"},
					{ScenarioName: "Test B", Status: "passed"},
				},
			},
			"ac-2": {
				HasEvidence: true,
				TotalTests:  2,
				PassedTests: 1,
				FailedTests: 1,
				Tests: []evidence.TestEvidence{
					{ScenarioName: "Test C", Status: "passed"},
					{ScenarioName: "Test D", Status: "failed"},
				},
			},
			// ac-3 has no evidence entry
		}

		findings := createFindingsForModule(config, "test-module", controlIDs, controlTestEvidence, ec, nil)
		require.NotNil(t, findings)
		require.Len(t, *findings, 3)

		// Build map by control ID for easier assertions
		findingMap := make(map[string]oscalTypes.Finding)
		for _, f := range *findings {
			findingMap[f.Target.TargetId] = f
		}

		assert.Equal(t, oscal.StateSatisfied, findingMap["ac-1"].Target.Status.State)
		assert.Equal(t, oscal.StateNotSatisfied, findingMap["ac-2"].Target.Status.State)
		assert.Equal(t, oscal.StateNotSatisfied, findingMap["ac-3"].Target.Status.State)
	})

	t.Run("finding title uses uppercase control ID", func(t *testing.T) {
		controlIDs := []string{"au-5"}
		ec := &evidence.EvidenceCollection{}

		findings := createFindingsForModule(config, "test-module", controlIDs, nil, ec, nil)
		require.NotNil(t, findings)
		require.Len(t, *findings, 1)

		assert.Equal(t, "Control AU-5 Assessment", (*findings)[0].Title)
	})

	t.Run("finding with test coverage includes skipped count", func(t *testing.T) {
		controlIDs := []string{"ac-1"}
		ec := &evidence.EvidenceCollection{}
		controlTestEvidence := map[string]*evidence.ControlTestEvidence{
			"ac-1": {
				HasEvidence:  true,
				TotalTests:   5,
				PassedTests:  2,
				FailedTests:  1,
				SkippedTests: 2,
				Tests: []evidence.TestEvidence{
					{ScenarioName: "T1", Status: "passed"},
					{ScenarioName: "T2", Status: "passed"},
					{ScenarioName: "T3", Status: "failed"},
					{ScenarioName: "T4", Status: "skipped"},
					{ScenarioName: "T5", Status: "skipped"},
				},
			},
		}

		findings := createFindingsForModule(config, "test-module", controlIDs, controlTestEvidence, ec, nil)
		require.NotNil(t, findings)
		require.Len(t, *findings, 1)

		propMap := make(map[string]string)
		for _, p := range *(*findings)[0].Props {
			propMap[p.Name] = p.Value
		}
		coverage := propMap["test-coverage"]
		assert.Contains(t, coverage, "5 tests")
		assert.Contains(t, coverage, "2 passed")
		assert.Contains(t, coverage, "1 failed")
	})
}

func TestBuildAssessmentResultsForModule(t *testing.T) {
	t.Run("builds valid assessment results with test evidence", func(t *testing.T) {
		config := &AssessConfig{
			WorkspaceRoot: "/workspace",
			OutputDir:     "/workspace/out/risk/20250101-120000",
		}

		// Build a simple profile with 2 controls
		controlIDs := []string{"ac-1", "ac-2"}
		withIds := make([]string, len(controlIDs))
		copy(withIds, controlIDs)
		includeControls := []oscalTypes.SelectControlById{
			{WithIds: &withIds},
		}
		profile := &oscalTypes.Profile{
			Imports: []oscalTypes.Import{
				{
					Href:            "https://example.com/catalog.json",
					IncludeControls: &includeControls,
				},
			},
		}

		// Build evidence collection with test data that has control tags
		ec := &evidence.EvidenceCollection{
			Module: "my-module",
			TestManifestData: &evidence.TestManifestData{
				TestAgent: "gotest",
				TestID:    "run-001",
				TestTime:  time.Now(),
				Tests: []evidence.TestEntryData{
					{
						Name:     "TestAuth",
						Package:  "pkg/auth",
						Type:     "gotest",
						Suite:    "unit",
						Status:   "passed",
						Tags:     []string{"@control:ac-1"},
						FilePath: "auth_test.go",
					},
				},
			},
			TestSummary: &evidence.TestSummary{
				Total:  1,
				Passed: 1,
			},
		}

		ar, err := buildAssessmentResultsForModule(config, "my-module", profile, ec)
		require.NoError(t, err)
		require.NotNil(t, ar)

		assert.NotEmpty(t, ar.UUID)
		assert.Contains(t, ar.Metadata.Title, "my-module")
		assert.Len(t, ar.Results, 1)

		result := ar.Results[0]
		assert.NotEmpty(t, result.UUID)
		assert.Contains(t, result.Title, "my-module")

		// Check findings exist for both controls
		require.NotNil(t, result.Findings)
		assert.Len(t, *result.Findings, 2)

		// Check observations exist
		require.NotNil(t, result.Observations)
		assert.NotEmpty(t, *result.Observations)

		// Check reviewed controls
		assert.NotEmpty(t, result.ReviewedControls.ControlSelections)

		// Check properties
		require.NotNil(t, result.Props)
		propMap := make(map[string]string)
		for _, p := range *result.Props {
			propMap[p.Name] = p.Value
		}
		assert.Equal(t, "automated", propMap["assessment-type"])
	})

	t.Run("builds results with no test manifest data", func(t *testing.T) {
		config := &AssessConfig{
			WorkspaceRoot: "/workspace",
		}

		controlIDs := []string{"ac-1"}
		withIds := make([]string, len(controlIDs))
		copy(withIds, controlIDs)
		includeControls := []oscalTypes.SelectControlById{
			{WithIds: &withIds},
		}
		profile := &oscalTypes.Profile{
			Imports: []oscalTypes.Import{
				{
					Href:            "https://example.com/catalog.json",
					IncludeControls: &includeControls,
				},
			},
		}

		// Evidence with no test manifest
		ec := &evidence.EvidenceCollection{
			Module: "empty-module",
		}

		ar, err := buildAssessmentResultsForModule(config, "empty-module", profile, ec)
		require.NoError(t, err)
		require.NotNil(t, ar)

		// ac-1 should be not-satisfied since no test evidence
		result := ar.Results[0]
		require.NotNil(t, result.Findings)
		require.Len(t, *result.Findings, 1)
		assert.Equal(t, oscal.StateNotSatisfied, (*result.Findings)[0].Target.Status.State)
		assert.Contains(t, (*result.Findings)[0].Remarks, "No test coverage")
	})
}

func TestBuildRemarksGrouping(t *testing.T) {
	// Ensure the remarks builder correctly groups tests by status
	t.Run("remarks groups tests by status", func(t *testing.T) {
		testEv := &evidence.ControlTestEvidence{
			HasEvidence: true,
			TotalTests:  4,
			PassedTests: 2,
			FailedTests: 1,
			Tests: []evidence.TestEvidence{
				{ScenarioName: "Test Passed 1", FeaturePath: "a.feature", Status: "passed"},
				{ScenarioName: "Test Failed 1", FeaturePath: "b.feature", Status: "failed"},
				{ScenarioName: "Test Passed 2", FeaturePath: "c.feature", Status: "passed"},
				{ScenarioName: "Test Unknown", FeaturePath: "d.feature", Status: "pending"},
			},
		}

		remarks := buildRemarksFromEvidence("ac-1", testEv)

		// Verify order: passed section comes before failed section
		passedIdx := strings.Index(remarks, "Passed:")
		failedIdx := strings.Index(remarks, "Failed:")
		notRunIdx := strings.Index(remarks, "Not Run/Skipped:")

		assert.Greater(t, failedIdx, passedIdx, "Failed section should come after Passed section")
		assert.Greater(t, notRunIdx, failedIdx, "Not Run section should come after Failed section")
	})
}
