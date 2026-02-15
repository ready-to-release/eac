package evidence

import "time"

// TestResults contains paths to test evidence files for a module.
type TestResults struct {
	ModuleName       string    `json:"module_name"`
	UnitTestFiles    []string  `json:"unit_test_files"`
	AcceptanceFiles  []string  `json:"acceptance_files"`
	TestRunDirectory string    `json:"test_run_directory"`
	LastModified     time.Time `json:"last_modified"`
}

// TestCase represents a single test case result.
type TestCase struct {
	Name       string        `json:"name"`
	Status     string        `json:"status"` // "passed", "failed", "skipped"
	Duration   time.Duration `json:"duration,omitempty"`
	ControlIDs []string      `json:"control_ids,omitempty"` // @control tags
	Error      string        `json:"error,omitempty"`
}

// TestSummary provides aggregate test result information.
type TestSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// TestManifestData holds simplified test manifest information for evidence collection.
//
// Import-cycle workaround: This type duplicates fields from the test command's
// internal manifest types (in impl/test/internal). The evidence package cannot
// import impl/test/internal because the test command already depends on the
// evidence package for collecting test results into the risk assessment pipeline.
// To break this cycle, TestManifestData, TestEntryData, SuiteResultData, and
// TestArtifactData are defined here as lightweight copies containing only the
// fields needed for evidence extraction and control-tag mapping. When test
// results are loaded, the test command converts its internal types into these
// evidence-local types before passing them to the risk assessment pipeline.
type TestManifestData struct {
	TestID    string                     `json:"test_id"`
	TestAgent string                     `json:"test_agent"`
	GitCommit string                     `json:"git_commit,omitempty"`
	BuildID   string                     `json:"build_id,omitempty"`
	TestTime  time.Time                  `json:"test_time"`
	Tests     []TestEntryData            `json:"tests"`
	Suites    map[string]SuiteResultData `json:"suites,omitempty"`
	Artifacts []TestArtifactData         `json:"artifacts"`
}

// SuiteResultData holds per-suite test result information.
// See TestManifestData for import-cycle workaround explanation.
type SuiteResultData struct {
	RunTime         time.Time `json:"run_time"`
	DurationSeconds float64   `json:"duration_seconds"`
	Total           int       `json:"total"`
	Passed          int       `json:"passed"`
	Failed          int       `json:"failed"`
	Skipped         int       `json:"skipped"`
}

// TestArtifactData holds test artifact information.
// See TestManifestData for import-cycle workaround explanation.
type TestArtifactData struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// ControlTestEvidence represents evidence for a specific control from tests.
type ControlTestEvidence struct {
	ControlID    string         `json:"control_id"`
	Tests        []TestEvidence `json:"tests"`
	TotalTests   int            `json:"total_tests"`
	PassedTests  int            `json:"passed_tests"`
	FailedTests  int            `json:"failed_tests"`
	SkippedTests int            `json:"skipped_tests"`
	HasEvidence  bool           `json:"has_evidence"`
}

// TestEvidence represents a single test covering a control.
type TestEvidence struct {
	FeaturePath  string `json:"feature_path"`
	FeatureName  string `json:"feature_name"`
	ScenarioName string `json:"scenario_name"`
	Status       string `json:"status"`   // "passed", "failed", "skipped", "not-run"
	Location     string `json:"location"` // file:line or file:scenario
}
