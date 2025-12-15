// Package internal provides test manifest functionality
package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// TestManifest represents the test manifest for a single module.
// This is stored per-module at out/test/<module>/test.manifest.json.
// It aggregates test results across all suite runs and tracks incremental test state.
type TestManifest struct {
	TestID              string                 `json:"test_id"`                         // Unique identifier (UUID locally, GitHub run ID in CI)
	TestAgent           string                 `json:"test_agent"`                      // Test agent type: "ci" or "devbox"
	Moniker             string                 `json:"moniker"`                         // Module identifier
	Type                string                 `json:"type"`                            // Module type
	TestTime            time.Time              `json:"test_time"`                       // When this manifest was last updated
	DurationSeconds     float64                `json:"duration_seconds,omitempty"`      // Total test duration in seconds
	GitCommit           string                 `json:"git_commit,omitempty"`            // Git commit SHA at test time
	Summary             TestSummary            `json:"summary"`                         // Aggregated test counts
	Suites              map[string]SuiteResult `json:"suites,omitempty"`                // Per-suite results
	Tests               []TestEntry            `json:"tests"`                           // Individual test results
	Artifacts           []TestArtifactInfo     `json:"artifacts"`                       // Test artifacts
	VerifiedUnchangedAt string                 `json:"verified_unchanged_at,omitempty"` // Git SHA when verified unchanged
	Version             string                 `json:"version"`                         // Manifest format version

	// Incremental test state
	SourceHash   string   `json:"source_hash,omitempty"`   // SHA-256 hash of module source files
	TestHash     string   `json:"test_hash,omitempty"`     // SHA-256 hash of module test files
	BuildID      string   `json:"build_id,omitempty"`      // BuildID from build manifest (links tests to specific build)
	Dependencies []string `json:"dependencies,omitempty"`  // Module dependencies at time of test
}

// TestSummary holds aggregated test counts
type TestSummary struct {
	Total     int `json:"total"`
	Passed    int `json:"passed"`
	Failed    int `json:"failed"`
	Skipped   int `json:"skipped"`
	Pending   int `json:"pending,omitempty"`
	Undefined int `json:"undefined,omitempty"`
}

// SuiteResult tracks results for a specific test suite run
type SuiteResult struct {
	RunTime         time.Time   `json:"run_time"`         // When this suite was last run
	DurationSeconds float64     `json:"duration_seconds"` // Suite execution duration
	Tests           TestSummary `json:"tests"`            // Test counts for this suite
}

// TestEntry represents a single test with its result
type TestEntry struct {
	Name       string   `json:"name"`                 // Test name or scenario name
	Package    string   `json:"package"`              // Package path
	Type       string   `json:"type"`                 // gotest, godog, mocha, tscucumber
	Suite      string   `json:"suite"`                // Suite moniker (component, integration, etc.)
	Status     string   `json:"status"`               // passed, failed, skipped, undefined, pending
	DurationMs int64    `json:"duration_ms,omitempty"`// Test duration in milliseconds
	Tags       []string `json:"tags,omitempty"`       // Test tags
	Error      string   `json:"error,omitempty"`      // Error message if failed
	FilePath   string   `json:"file_path,omitempty"`  // Source file path
}

// TestArtifactInfo describes a test artifact
type TestArtifactInfo struct {
	Type   string `json:"type"`             // log, report, coverage, file
	ID     string `json:"id"`               // Artifact identifier
	Name   string `json:"name"`             // Filename
	Path   string `json:"path"`             // Relative path from module test dir
	Size   int64  `json:"size,omitempty"`   // File size in bytes
	SHA256 string `json:"sha256,omitempty"` // SHA-256 hash
}

// Test status constants
const (
	TestStatusPassed    = "passed"
	TestStatusFailed    = "failed"
	TestStatusSkipped   = "skipped"
	TestStatusUndefined = "undefined"
	TestStatusPending   = "pending"
)

// Test artifact type constants
const (
	TestArtifactTypeLog      = "log"
	TestArtifactTypeReport   = "report"
	TestArtifactTypeCoverage = "coverage"
	TestArtifactTypeFile     = "file"
)

const testManifestVersion = "1.0"
const testManifestFileName = "test.manifest.json"

// TestAgentCI is the test agent value for CI runs (GitHub Actions)
const TestAgentCI = "ci"

// TestAgentDevbox is the test agent value for local developer runs
const TestAgentDevbox = "devbox"

// NewTestManifest creates a new test manifest for a module.
// In CI (GITHUB_RUN_ID set), uses the run ID as TestID and sets TestAgent to "ci".
// Locally, generates a UUID for TestID and sets TestAgent to "devbox".
func NewTestManifest(moniker, moduleType, gitCommit string) *TestManifest {
	testID := uuid.New().String()
	testAgent := TestAgentDevbox

	// In CI, use GitHub run ID for traceability
	if runID := os.Getenv("GITHUB_RUN_ID"); runID != "" {
		testID = runID
		testAgent = TestAgentCI
	}

	return &TestManifest{
		TestID:    testID,
		TestAgent: testAgent,
		Moniker:   moniker,
		Type:      moduleType,
		TestTime:  time.Now(),
		GitCommit: gitCommit,
		Summary:   TestSummary{},
		Suites:    make(map[string]SuiteResult),
		Tests:     []TestEntry{},
		Artifacts: []TestArtifactInfo{},
		Version:   testManifestVersion,
	}
}

// Save writes the manifest to the module's test output directory.
// The manifest is stored at <moduleTestDir>/test.manifest.json
func (m *TestManifest) Save(moduleTestDir string) error {
	manifestPath := filepath.Join(moduleTestDir, testManifestFileName)

	// Create directory if needed
	if err := os.MkdirAll(moduleTestDir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test manifest: %w", err)
	}

	// Write to file
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write test manifest: %w", err)
	}

	return nil
}

// LoadTestManifest loads a module's test manifest from its test output directory
func LoadTestManifest(moduleTestDir string) (*TestManifest, error) {
	manifestPath := filepath.Join(moduleTestDir, testManifestFileName)

	// Check if manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("test manifest not found at %s", manifestPath)
	}

	// Read file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test manifest: %w", err)
	}

	// Unmarshal
	var manifest TestManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal test manifest: %w", err)
	}

	return &manifest, nil
}

// LoadOrCreateTestManifest loads an existing manifest or creates a new one if not found.
// This is useful for aggregating results across multiple suite runs.
func LoadOrCreateTestManifest(moduleTestDir, moniker, moduleType, gitCommit string) (*TestManifest, error) {
	manifest, err := LoadTestManifest(moduleTestDir)
	if err == nil {
		// Update metadata for this run
		manifest.TestTime = time.Now()
		if gitCommit != "" {
			manifest.GitCommit = gitCommit
		}
		return manifest, nil
	}

	// Create new manifest if not found
	return NewTestManifest(moniker, moduleType, gitCommit), nil
}

// AddSuiteResult adds or updates results for a specific test suite.
// This aggregates results when running multiple suites against the same module.
func (m *TestManifest) AddSuiteResult(suiteName string, result SuiteResult) {
	if m.Suites == nil {
		m.Suites = make(map[string]SuiteResult)
	}
	m.Suites[suiteName] = result
	m.recalculateSummary()
}

// AddTest adds a test entry to the manifest.
// If a test with the same name, package, and suite exists, it is updated.
func (m *TestManifest) AddTest(entry TestEntry) {
	// Check if test already exists (same name, package, suite)
	for i, existing := range m.Tests {
		if existing.Name == entry.Name && existing.Package == entry.Package && existing.Suite == entry.Suite {
			m.Tests[i] = entry
			m.recalculateSummary()
			return
		}
	}
	// Add new test
	m.Tests = append(m.Tests, entry)
	m.recalculateSummary()
}

// ClearSuiteTests removes all tests for a specific suite.
// Call this before adding fresh test results for a suite run.
func (m *TestManifest) ClearSuiteTests(suiteName string) {
	filtered := make([]TestEntry, 0, len(m.Tests))
	for _, t := range m.Tests {
		if t.Suite != suiteName {
			filtered = append(filtered, t)
		}
	}
	m.Tests = filtered
}

// AddArtifact adds a test artifact to the manifest.
// If an artifact with the same ID exists, it is updated.
func (m *TestManifest) AddArtifact(artifact TestArtifactInfo) {
	for i, existing := range m.Artifacts {
		if existing.ID == artifact.ID {
			m.Artifacts[i] = artifact
			return
		}
	}
	m.Artifacts = append(m.Artifacts, artifact)
}

// recalculateSummary updates the summary counts from individual test entries
func (m *TestManifest) recalculateSummary() {
	m.Summary = TestSummary{}
	for _, t := range m.Tests {
		m.Summary.Total++
		switch t.Status {
		case TestStatusPassed:
			m.Summary.Passed++
		case TestStatusFailed:
			m.Summary.Failed++
		case TestStatusSkipped:
			m.Summary.Skipped++
		case TestStatusPending:
			m.Summary.Pending++
		case TestStatusUndefined:
			m.Summary.Undefined++
		}
	}
}

// AllPassed returns true if all tests passed (no failures, undefined, or pending)
func (m *TestManifest) AllPassed() bool {
	return m.Summary.Failed == 0 && m.Summary.Undefined == 0 && m.Summary.Pending == 0
}

// UpdateVerifiedUnchangedAt updates the verification timestamp and saves the manifest.
// This is used when tests are skipped because the module is unchanged.
func (m *TestManifest) UpdateVerifiedUnchangedAt(moduleTestDir, gitCommit string) error {
	m.VerifiedUnchangedAt = gitCommit
	return m.Save(moduleTestDir)
}

// GetTestManifestPath returns the path to a module's test manifest
func GetTestManifestPath(moduleTestDir string) string {
	return filepath.Join(moduleTestDir, testManifestFileName)
}

// TestManifestExists checks if a test manifest exists for a module
func TestManifestExists(moduleTestDir string) bool {
	_, err := os.Stat(GetTestManifestPath(moduleTestDir))
	return err == nil
}

// SetIncrementalState sets the incremental test state fields.
// This should be called after tests complete to record state for change detection.
func (m *TestManifest) SetIncrementalState(sourceHash, testHash, buildID string, dependencies []string) {
	m.SourceHash = sourceHash
	m.TestHash = testHash
	m.BuildID = buildID
	m.Dependencies = dependencies
}

// NeedsRetest checks if the module needs retesting based on stored incremental state.
// Returns (needsRetest bool, reason string).
// A module needs retesting if:
// 1. Source files changed (SourceHash differs)
// 2. Test files changed (TestHash differs)
// 3. Build changed (BuildID differs)
// 4. Previous tests failed
// 5. No previous test state exists
func (m *TestManifest) NeedsRetest(currentSourceHash, currentTestHash, currentBuildID string) (bool, string) {
	// No source hash stored = no prior test state
	if m.SourceHash == "" {
		return true, "no prior test state"
	}

	// Previous tests failed
	if !m.AllPassed() {
		return true, "previous test failed"
	}

	// Source files changed
	if currentSourceHash != "" && m.SourceHash != currentSourceHash {
		return true, "source files changed"
	}

	// Test files changed
	if currentTestHash != "" && m.TestHash != currentTestHash {
		return true, "test files changed"
	}

	// Build changed (rebuild detected)
	if currentBuildID != "" && m.BuildID != "" && m.BuildID != currentBuildID {
		return true, "build changed (rebuild detected)"
	}

	// New build detected (module was tested without build, now has one)
	if currentBuildID != "" && m.BuildID == "" {
		return true, "new build detected"
	}

	return false, ""
}

// GetSuiteNames returns the list of suite names that have been run for this module.
func (m *TestManifest) GetSuiteNames() []string {
	if m.Suites == nil {
		return nil
	}
	names := make([]string, 0, len(m.Suites))
	for name := range m.Suites {
		names = append(names, name)
	}
	return names
}

// HasSuiteRun checks if a specific suite has been run for this module.
func (m *TestManifest) HasSuiteRun(suiteName string) bool {
	if m.Suites == nil {
		return false
	}
	_, exists := m.Suites[suiteName]
	return exists
}

// GetLastSuiteRunTime returns when a specific suite was last run.
// Returns zero time if suite hasn't been run.
func (m *TestManifest) GetLastSuiteRunTime(suiteName string) time.Time {
	if m.Suites == nil {
		return time.Time{}
	}
	if result, ok := m.Suites[suiteName]; ok {
		return result.RunTime
	}
	return time.Time{}
}
