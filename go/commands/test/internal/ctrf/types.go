// Package ctrf provides types and utilities for Common Test Report Format (CTRF).
// CTRF is a JSON schema for universal test reporting across frameworks.
// See: https://ctrf.io/
package ctrf

// Report represents a CTRF test report.
type Report struct {
	Results Results `json:"results"`
}

// Results contains the test results.
type Results struct {
	Tool        Tool         `json:"tool"`
	Summary     Summary      `json:"summary"`
	Tests       []Test       `json:"tests"`
	Environment *Environment `json:"environment,omitempty"`
}

// Tool identifies the test framework.
type Tool struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// Summary provides aggregate test counts.
type Summary struct {
	Tests   int   `json:"tests"`
	Passed  int   `json:"passed"`
	Failed  int   `json:"failed"`
	Pending int   `json:"pending"`
	Skipped int   `json:"skipped"`
	Other   int   `json:"other"`
	Start   int64 `json:"start"` // Unix milliseconds
	Stop    int64 `json:"stop"`  // Unix milliseconds
}

// Test represents a single test result.
type Test struct {
	Name      string `json:"name"`
	Status    Status `json:"status"`
	Duration  int64  `json:"duration"` // Milliseconds
	Message   string `json:"message,omitempty"`
	Trace     string `json:"trace,omitempty"`
	RawStatus string `json:"rawStatus,omitempty"`
	Suite     string `json:"suite,omitempty"`
	FilePath  string `json:"filePath,omitempty"`
}

// Status represents the test result status.
type Status string

const (
	StatusPassed  Status = "passed"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
	StatusPending Status = "pending"
	StatusOther   Status = "other"
)

// Environment provides optional environment metadata.
type Environment struct {
	OS       string `json:"os,omitempty"`
	Arch     string `json:"arch,omitempty"`
	Language string `json:"language,omitempty"`
}
