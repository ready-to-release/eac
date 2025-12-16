package evidence

import (
	"testing"
)

func TestGetControlTestEvidenceFromManifest(t *testing.T) {
	tests := []struct {
		name          string
		tests         []TestEntryData
		testSuite     string
		wantControls  int
		wantEvidence  map[string]int // controlID -> test count
	}{
		{
			name: "extract control tags from manifest tests",
			tests: []TestEntryData{
				{
					Name:     "Test user login",
					Package:  "auth",
					Type:     "godog",
					Suite:    "acceptance",
					Status:   "passed",
					Tags:     []string{"@control:ac-2", "@smoke"},
					FilePath: "specs/auth.feature",
				},
				{
					Name:     "Test user logout",
					Package:  "auth",
					Type:     "godog",
					Suite:    "acceptance",
					Status:   "passed",
					Tags:     []string{"@control:ac-2", "@control:ia-2"},
					FilePath: "specs/auth.feature",
				},
			},
			testSuite:    "",
			wantControls: 2,
			wantEvidence: map[string]int{
				"ac-2": 2,
				"ia-2": 1,
			},
		},
		{
			name: "filter by suite",
			tests: []TestEntryData{
				{
					Name:     "Unit test",
					Package:  "auth",
					Type:     "gotest",
					Suite:    "unit",
					Status:   "passed",
					Tags:     []string{"@control:ac-1"},
					FilePath: "auth_test.go",
				},
				{
					Name:     "Acceptance test",
					Package:  "auth",
					Type:     "godog",
					Suite:    "acceptance",
					Status:   "passed",
					Tags:     []string{"@control:ac-2"},
					FilePath: "specs/auth.feature",
				},
			},
			testSuite:    "acceptance",
			wantControls: 1,
			wantEvidence: map[string]int{
				"ac-2": 1,
			},
		},
		{
			name: "multiple controls in one tag",
			tests: []TestEntryData{
				{
					Name:     "Test authentication",
					Package:  "auth",
					Type:     "godog",
					Suite:    "acceptance",
					Status:   "passed",
					Tags:     []string{"@controls:ac-2,ia-2,au-3"},
					FilePath: "specs/auth.feature",
				},
			},
			testSuite:    "",
			wantControls: 3,
			wantEvidence: map[string]int{
				"ac-2": 1,
				"ia-2": 1,
				"au-3": 1,
			},
		},
		{
			name: "tests with different statuses",
			tests: []TestEntryData{
				{
					Name:     "Test 1",
					Package:  "auth",
					Type:     "godog",
					Suite:    "acceptance",
					Status:   "passed",
					Tags:     []string{"@control:ac-2"},
					FilePath: "specs/auth.feature",
				},
				{
					Name:     "Test 2",
					Package:  "auth",
					Type:     "godog",
					Suite:    "acceptance",
					Status:   "failed",
					Tags:     []string{"@control:ac-2"},
					FilePath: "specs/auth.feature",
				},
				{
					Name:     "Test 3",
					Package:  "auth",
					Type:     "godog",
					Suite:    "acceptance",
					Status:   "skipped",
					Tags:     []string{"@control:ac-2"},
					FilePath: "specs/auth.feature",
				},
			},
			testSuite:    "",
			wantControls: 1,
			wantEvidence: map[string]int{
				"ac-2": 3,
			},
		},
		{
			name: "no control tags",
			tests: []TestEntryData{
				{
					Name:     "Test without controls",
					Package:  "auth",
					Type:     "godog",
					Suite:    "acceptance",
					Status:   "passed",
					Tags:     []string{"@smoke", "@integration"},
					FilePath: "specs/auth.feature",
				},
			},
			testSuite:    "",
			wantControls: 0,
			wantEvidence: map[string]int{},
		},
		{
			name:         "empty tests",
			tests:        []TestEntryData{},
			testSuite:    "",
			wantControls: 0,
			wantEvidence: map[string]int{},
		},
		{
			name: "composite suite - unit+integration",
			tests: []TestEntryData{
				{
					Name:     "Unit test",
					Package:  "auth",
					Type:     "gotest",
					Suite:    "unit",
					Status:   "passed",
					Tags:     []string{"@control:ac-1"},
					FilePath: "auth_test.go",
				},
				{
					Name:     "Integration test",
					Package:  "auth",
					Type:     "godog",
					Suite:    "integration",
					Status:   "passed",
					Tags:     []string{"@control:ac-2"},
					FilePath: "specs/auth.feature",
				},
				{
					Name:     "Acceptance test",
					Package:  "auth",
					Type:     "godog",
					Suite:    "acceptance",
					Status:   "passed",
					Tags:     []string{"@control:ac-3"},
					FilePath: "specs/auth_acceptance.feature",
				},
			},
			testSuite:    "unit+integration",
			wantControls: 2,
			wantEvidence: map[string]int{
				"ac-1": 1,
				"ac-2": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evidenceMap := GetControlTestEvidenceFromManifest(tt.tests, tt.testSuite)

			if len(evidenceMap) != tt.wantControls {
				t.Errorf("Control count = %d, want %d", len(evidenceMap), tt.wantControls)
			}

			// Verify test counts for each control
			for controlID, wantCount := range tt.wantEvidence {
				evidence, exists := evidenceMap[controlID]
				if !exists {
					t.Errorf("Control %s not found in evidence map", controlID)
					continue
				}

				if len(evidence.Tests) != wantCount {
					t.Errorf("Control %s: test count = %d, want %d", controlID, len(evidence.Tests), wantCount)
				}

				if evidence.ControlID != controlID {
					t.Errorf("Control %s: ControlID field = %s, want %s", controlID, evidence.ControlID, controlID)
				}

				if evidence.TotalTests != wantCount {
					t.Errorf("Control %s: TotalTests = %d, want %d", controlID, evidence.TotalTests, wantCount)
				}

				if !evidence.HasEvidence {
					t.Errorf("Control %s: HasEvidence should be true", controlID)
				}
			}
		})
	}
}

func TestGetControlTestEvidenceFromManifest_Aggregation(t *testing.T) {
	tests := []TestEntryData{
		{
			Name:     "Test 1",
			Package:  "auth",
			Type:     "godog",
			Suite:    "acceptance",
			Status:   "passed",
			Tags:     []string{"@control:ac-2"},
			FilePath: "specs/auth.feature",
		},
		{
			Name:     "Test 2",
			Package:  "auth",
			Type:     "godog",
			Suite:    "acceptance",
			Status:   "passed",
			Tags:     []string{"@control:ac-2"},
			FilePath: "specs/auth.feature",
		},
		{
			Name:     "Test 3",
			Package:  "auth",
			Type:     "godog",
			Suite:    "acceptance",
			Status:   "failed",
			Tags:     []string{"@control:ac-2"},
			FilePath: "specs/auth.feature",
		},
		{
			Name:     "Test 4",
			Package:  "auth",
			Type:     "godog",
			Suite:    "acceptance",
			Status:   "skipped",
			Tags:     []string{"@control:ac-2"},
			FilePath: "specs/auth.feature",
		},
	}

	evidenceMap := GetControlTestEvidenceFromManifest(tests, "")

	evidence := evidenceMap["ac-2"]
	if evidence == nil {
		t.Fatal("Evidence for ac-2 not found")
	}

	if evidence.TotalTests != 4 {
		t.Errorf("TotalTests = %d, want 4", evidence.TotalTests)
	}

	if evidence.PassedTests != 2 {
		t.Errorf("PassedTests = %d, want 2", evidence.PassedTests)
	}

	if evidence.FailedTests != 1 {
		t.Errorf("FailedTests = %d, want 1", evidence.FailedTests)
	}

	if evidence.SkippedTests != 1 {
		t.Errorf("SkippedTests = %d, want 1", evidence.SkippedTests)
	}
}

func TestExtractControlIDsFromTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		wantIDs  []string
	}{
		{
			name:    "single control",
			tag:     "@control:ac-2",
			wantIDs: []string{"ac-2"},
		},
		{
			name:    "single control with enhancement",
			tag:     "@control:ac-2(1)",
			wantIDs: []string{"ac-2(1)"},
		},
		{
			name:    "multiple controls",
			tag:     "@controls:ac-2,ia-2,au-3",
			wantIDs: []string{"ac-2", "ia-2", "au-3"},
		},
		{
			name:    "multiple controls with enhancements",
			tag:     "@controls:ac-2(1),ia-2(3),au-3",
			wantIDs: []string{"ac-2(1)", "ia-2(3)", "au-3"},
		},
		{
			name:    "not a control tag",
			tag:     "@smoke",
			wantIDs: []string{},
		},
		{
			name:    "empty tag",
			tag:     "",
			wantIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIDs := extractControlIDsFromTag(tt.tag)

			if len(gotIDs) != len(tt.wantIDs) {
				t.Errorf("extractControlIDsFromTag() returned %d IDs, want %d", len(gotIDs), len(tt.wantIDs))
			}

			for i, wantID := range tt.wantIDs {
				if i >= len(gotIDs) {
					t.Errorf("Missing control ID: %s", wantID)
					continue
				}
				if gotIDs[i] != wantID {
					t.Errorf("Control ID[%d] = %s, want %s", i, gotIDs[i], wantID)
				}
			}
		})
	}
}

func TestParseSuiteFilter(t *testing.T) {
	tests := []struct {
		name        string
		suiteFilter string
		want        []string
	}{
		{
			name:        "empty string - no filter",
			suiteFilter: "",
			want:        nil,
		},
		{
			name:        "single suite",
			suiteFilter: "unit",
			want:        []string{"unit"},
		},
		{
			name:        "composite suite - two",
			suiteFilter: "unit+integration",
			want:        []string{"unit", "integration"},
		},
		{
			name:        "composite suite - three",
			suiteFilter: "unit+integration+acceptance",
			want:        []string{"unit", "integration", "acceptance"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSuiteFilter(tt.suiteFilter)

			if tt.want == nil && got != nil {
				t.Errorf("parseSuiteFilter() = %v, want nil", got)
				return
			}

			if tt.want != nil && got == nil {
				t.Errorf("parseSuiteFilter() = nil, want %v", tt.want)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("parseSuiteFilter() returned %d suites, want %d", len(got), len(tt.want))
				return
			}

			for i, wantSuite := range tt.want {
				if got[i] != wantSuite {
					t.Errorf("parseSuiteFilter()[%d] = %s, want %s", i, got[i], wantSuite)
				}
			}
		})
	}
}
