package testing

import (
	"time"

	"github.com/ready-to-release/eac/go/core/workunit"
)

// TestModuleView represents aggregated test data for a single module,
// derived from UoW manifests and their artifacts.
type TestModuleView struct {
	Module          string                  `json:"module"`
	ExecutedAt      time.Time               `json:"executed_at"`
	Duration        time.Duration           `json:"duration"`
	ExitCode        int                     `json:"exit_code"`
	UoWCount        int                     `json:"uow_count"`
	Summary         TestSummary             `json:"summary"`
	Tags            workunit.TagSummary     `json:"tags,omitempty"`
	Suites          map[string]*SuiteSummary `json:"suites,omitempty"`
	Tests           []TestEntry             `json:"tests"`
	Artifacts       []ArtifactRef           `json:"artifacts"`
	CucumberReports []string                `json:"-"`
	CTRFReports     []string                `json:"-"`
	CoverageFiles   []string                `json:"-"`
}

// TestSummary holds aggregated test counts.
type TestSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// SuiteSummary holds per-suite aggregated test counts.
type SuiteSummary struct {
	Total   int       `json:"total"`
	Passed  int       `json:"passed"`
	Failed  int       `json:"failed"`
	Skipped int       `json:"skipped"`
	RunTime time.Time `json:"run_time,omitempty"`
}

// TestEntry represents a single test result.
type TestEntry struct {
	Name       string   `json:"name"`
	Module     string   `json:"module"`
	Package    string   `json:"package"`
	Type       string   `json:"type"`
	Suite      string   `json:"suite"`
	Status     string   `json:"status"`
	DurationMs int64    `json:"duration_ms,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	FilePath   string   `json:"file_path,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// ArtifactRef is a reference to a test output file located via UoW manifest.
type ArtifactRef struct {
	Type   string `json:"type"`
	Path   string `json:"path"`
	UoWDir string `json:"uow_dir"`
}

// Test status constants.
const (
	StatusPassed  = "passed"
	StatusFailed  = "failed"
	StatusSkipped = "skipped"
)

// Artifact type constants.
const (
	ArtifactTypeCucumberReport = "cucumber-report"
	ArtifactTypeCTRFReport     = "ctrf-report"
	ArtifactTypeCoverage       = "coverage"
	ArtifactTypeLog            = "log"
	ArtifactTypeManualReport   = "manual-report"
)

// AllPassed returns true if no tests failed.
func (v *TestModuleView) AllPassed() bool {
	return v.Summary.Failed == 0
}

// computeSummary recalculates summary from test entries.
func (v *TestModuleView) computeSummary() {
	v.Summary = TestSummary{}
	for i := range v.Tests {
		v.Summary.Total++
		switch v.Tests[i].Status {
		case StatusPassed:
			v.Summary.Passed++
		case StatusFailed:
			v.Summary.Failed++
		case StatusSkipped:
			v.Summary.Skipped++
		}
	}
}
