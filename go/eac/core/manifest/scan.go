// Package manifest provides manifest types for build, test, and scan tracking.
package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// ScanManifest represents the scan manifest for a single module.
// This is stored per-module at out/scan/<module>/scan.manifest.json.
// It aggregates scan results from all scanner runs and tracks security scan state.
type ScanManifest struct {
	ScanID          string                   `json:"scan_id"`                    // Unique identifier (UUID locally, GitHub run ID in CI)
	ScanAgent       string                   `json:"scan_agent"`                 // Scan agent type: "ci" or "devbox"
	Moniker         string                   `json:"moniker"`                    // Module identifier
	Type            string                   `json:"type,omitempty"`             // Module type
	ScanTime        time.Time                `json:"scan_time"`                  // When this manifest was last updated
	DurationSeconds float64                  `json:"duration_seconds,omitempty"` // Total scan duration in seconds
	GitCommit       string                   `json:"git_commit,omitempty"`       // Git commit SHA at scan time
	Summary         ScanSummary              `json:"summary"`                    // Aggregated scan counts
	Scans           map[string]ScannerResult `json:"scans"`                      // Per-scanner results
	Artifacts       []ScanArtifactInfo       `json:"artifacts"`                  // Scan artifacts
	Version         string                   `json:"version"`                    // Manifest format version
}

// ScanSummary holds aggregated scan counts
type ScanSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped,omitempty"`
}

// ScannerResult tracks results for a specific scanner run
type ScannerResult struct {
	Status          string    `json:"status"`                     // passed, failed, skipped
	RunTime         time.Time `json:"run_time"`                   // When this scanner was run
	DurationSeconds float64   `json:"duration_seconds,omitempty"` // Scanner execution duration
	EvidencePath    string    `json:"evidence_path,omitempty"`    // Path to evidence file
	Error           string    `json:"error,omitempty"`            // Error message if failed
	FindingsCount   int       `json:"findings_count,omitempty"`   // Number of findings
}

// ScanArtifactInfo describes a scan artifact
type ScanArtifactInfo struct {
	Type    string `json:"type"`             // evidence, report, log
	ID      string `json:"id"`               // Artifact identifier
	Name    string `json:"name,omitempty"`   // Filename
	Path    string `json:"path"`             // Relative path from module security dir
	Scanner string `json:"scanner"`          // Scanner that produced this artifact
	Size    int64  `json:"size,omitempty"`   // File size in bytes
	SHA256  string `json:"sha256,omitempty"` // SHA-256 hash
}

// Scanner status constants
const (
	ScanStatusPassed  = "passed"
	ScanStatusFailed  = "failed"
	ScanStatusSkipped = "skipped"
)

// Scan artifact type constants
const (
	ScanArtifactTypeEvidence = "evidence"
	ScanArtifactTypeReport   = "report"
	ScanArtifactTypeLog      = "log"
)

const scanManifestVersion = "1.0"
const scanManifestFileName = "scan.manifest.json"

// ScanAgentCI is the scan agent value for CI runs (GitHub Actions)
const ScanAgentCI = "ci"

// ScanAgentDevbox is the scan agent value for local developer runs
const ScanAgentDevbox = "devbox"

// NewScanManifest creates a new scan manifest for a module.
// In CI (GITHUB_RUN_ID set), uses the run ID as ScanID and sets ScanAgent to "ci".
// Locally, generates a UUID for ScanID and sets ScanAgent to "devbox".
func NewScanManifest(moniker, moduleType, gitCommit string) *ScanManifest {
	scanID := uuid.New().String()
	scanAgent := ScanAgentDevbox

	// In CI, use GitHub run ID for traceability
	if runID := os.Getenv("GITHUB_RUN_ID"); runID != "" {
		scanID = runID
		scanAgent = ScanAgentCI
	}

	return &ScanManifest{
		ScanID:    scanID,
		ScanAgent: scanAgent,
		Moniker:   moniker,
		Type:      moduleType,
		ScanTime:  time.Now(),
		GitCommit: gitCommit,
		Summary:   ScanSummary{},
		Scans:     make(map[string]ScannerResult),
		Artifacts: []ScanArtifactInfo{},
		Version:   scanManifestVersion,
	}
}

// Save writes the manifest to the module's security output directory.
// The manifest is stored at <moduleSecurityDir>/scan.manifest.json
func (m *ScanManifest) Save(moduleSecurityDir string) error {
	manifestPath := filepath.Join(moduleSecurityDir, scanManifestFileName)

	// Create directory if needed
	if err := os.MkdirAll(moduleSecurityDir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scan manifest: %w", err)
	}

	// Write to file
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write scan manifest: %w", err)
	}

	return nil
}

// LoadScanManifest loads a module's scan manifest from its security output directory
func LoadScanManifest(moduleSecurityDir string) (*ScanManifest, error) {
	manifestPath := filepath.Join(moduleSecurityDir, scanManifestFileName)

	// Check if manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("scan manifest not found at %s", manifestPath)
	}

	// Read file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read scan manifest: %w", err)
	}

	// Unmarshal
	var manifest ScanManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scan manifest: %w", err)
	}

	return &manifest, nil
}

// LoadOrCreateScanManifest loads an existing manifest or creates a new one if not found.
// This is useful for aggregating results across multiple scanner runs.
func LoadOrCreateScanManifest(moduleSecurityDir, moniker, moduleType, gitCommit string) (*ScanManifest, error) {
	manifest, err := LoadScanManifest(moduleSecurityDir)
	if err == nil {
		// Update metadata for this run
		manifest.ScanTime = time.Now()
		if gitCommit != "" {
			manifest.GitCommit = gitCommit
		}
		return manifest, nil
	}

	// Create new manifest if not found
	return NewScanManifest(moniker, moduleType, gitCommit), nil
}

// AddScannerResult adds or updates results for a specific scanner.
// This aggregates results when running multiple scanners against the same module.
func (m *ScanManifest) AddScannerResult(scannerType string, result ScannerResult) {
	if m.Scans == nil {
		m.Scans = make(map[string]ScannerResult)
	}
	m.Scans[scannerType] = result
	m.recalculateSummary()
}

// AddArtifact adds a scan artifact to the manifest.
// If an artifact with the same ID exists, it is updated.
func (m *ScanManifest) AddArtifact(artifact ScanArtifactInfo) {
	for i, existing := range m.Artifacts {
		if existing.ID == artifact.ID {
			m.Artifacts[i] = artifact
			return
		}
	}
	m.Artifacts = append(m.Artifacts, artifact)
}

// recalculateSummary updates the summary counts from scanner results
func (m *ScanManifest) recalculateSummary() {
	m.Summary = ScanSummary{}
	for _, scan := range m.Scans {
		m.Summary.Total++
		switch scan.Status {
		case ScanStatusPassed:
			m.Summary.Passed++
		case ScanStatusFailed:
			m.Summary.Failed++
		case ScanStatusSkipped:
			m.Summary.Skipped++
		}
	}
}

// AllPassed returns true if all scans passed (no failures)
func (m *ScanManifest) AllPassed() bool {
	return m.Summary.Failed == 0
}

// GetScanManifestPath returns the path to a module's scan manifest
func GetScanManifestPath(moduleSecurityDir string) string {
	return filepath.Join(moduleSecurityDir, scanManifestFileName)
}

// ScanManifestExists checks if a scan manifest exists for a module
func ScanManifestExists(moduleSecurityDir string) bool {
	_, err := os.Stat(GetScanManifestPath(moduleSecurityDir))
	return err == nil
}

// GetScannerNames returns the list of scanner names that have been run for this module.
func (m *ScanManifest) GetScannerNames() []string {
	if m.Scans == nil {
		return nil
	}
	names := make([]string, 0, len(m.Scans))
	for name := range m.Scans {
		names = append(names, name)
	}
	return names
}

// HasScannerRun checks if a specific scanner has been run for this module.
func (m *ScanManifest) HasScannerRun(scannerType string) bool {
	if m.Scans == nil {
		return false
	}
	_, exists := m.Scans[scannerType]
	return exists
}

// GetScannerResult returns the result for a specific scanner.
// Returns nil if the scanner hasn't been run.
func (m *ScanManifest) GetScannerResult(scannerType string) *ScannerResult {
	if m.Scans == nil {
		return nil
	}
	if result, ok := m.Scans[scannerType]; ok {
		return &result
	}
	return nil
}

// ClearScanner removes results for a specific scanner.
// Useful when re-running a single scanner.
func (m *ScanManifest) ClearScanner(scannerType string) {
	if m.Scans == nil {
		return
	}
	delete(m.Scans, scannerType)

	// Also remove associated artifacts
	filtered := make([]ScanArtifactInfo, 0, len(m.Artifacts))
	for _, artifact := range m.Artifacts {
		if artifact.Scanner != scannerType {
			filtered = append(filtered, artifact)
		}
	}
	m.Artifacts = filtered

	m.recalculateSummary()
}

// SetDuration sets the total scan duration
func (m *ScanManifest) SetDuration(duration time.Duration) {
	m.DurationSeconds = duration.Seconds()
}

// GetFailedScanners returns the list of scanners that failed
func (m *ScanManifest) GetFailedScanners() []string {
	var failed []string
	for name, result := range m.Scans {
		if result.Status == ScanStatusFailed {
			failed = append(failed, name)
		}
	}
	return failed
}

// GetPassedScanners returns the list of scanners that passed
func (m *ScanManifest) GetPassedScanners() []string {
	var passed []string
	for name, result := range m.Scans {
		if result.Status == ScanStatusPassed {
			passed = append(passed, name)
		}
	}
	return passed
}
