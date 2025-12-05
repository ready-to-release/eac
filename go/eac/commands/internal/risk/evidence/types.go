// Package evidence provides evidence collection for risk assessment.
//
// This package handles discovering and loading evidence files from:
// - Test results (unit tests, acceptance tests)
// - Security scans (SBOM, vulnerability, SAST, etc.)
package evidence

import (
	"encoding/json"
	"time"
)

// EvidenceType represents the type of evidence.
type EvidenceType string

const (
	EvidenceTypeUnitTest       EvidenceType = "unit-test"
	EvidenceTypeAcceptanceTest EvidenceType = "acceptance-test"
	EvidenceTypeVulnerabilityScan EvidenceType = "vulnerability-scan"
	EvidenceTypeSBOM           EvidenceType = "sbom"
	EvidenceTypeSecretsScan    EvidenceType = "secrets-scan"
	EvidenceTypeSASTScan       EvidenceType = "sast-scan"
	EvidenceTypeComplianceScan EvidenceType = "compliance-scan"
	EvidenceTypeIaCScan        EvidenceType = "iac-scan"
	EvidenceTypeDAST           EvidenceType = "dast-scan"
)

// Evidence represents a piece of evidence for risk assessment.
type Evidence struct {
	Type       EvidenceType `json:"type"`
	Path       string       `json:"path"`
	ModTime    time.Time    `json:"mod_time"`
	Module     string       `json:"module,omitempty"`
	ControlIDs []string     `json:"control_ids,omitempty"` // Controls this evidence relates to
}

// TestResults contains paths to test evidence files for a module.
type TestResults struct {
	ModuleName       string    `json:"module_name"`
	UnitTestFiles    []string  `json:"unit_test_files"`
	AcceptanceFiles  []string  `json:"acceptance_files"`
	TestRunDirectory string    `json:"test_run_directory"`
	LastModified     time.Time `json:"last_modified"`
}

// SecurityResults contains paths to security evidence files for a module.
type SecurityResults struct {
	ModuleName     string `json:"module_name"`
	SBOMFile       string `json:"sbom_file,omitempty"`
	VulnFile       string `json:"vuln_file,omitempty"`
	SecretsFile    string `json:"secrets_file,omitempty"`
	SASTFile       string `json:"sast_file,omitempty"`
	ComplianceFile string `json:"compliance_file,omitempty"`
	IaCFile        string `json:"iac_file,omitempty"`
	ZAPFile        string `json:"zap_file,omitempty"`
	LastModified   time.Time `json:"last_modified"`
}

// SecurityEvidenceFile represents the standardized security evidence format.
// This matches the format used by the security command.
type SecurityEvidenceFile struct {
	Module    string          `json:"module"`
	Scanner   string          `json:"scanner"`
	Timestamp string          `json:"timestamp"`
	SHA256    string          `json:"sha256"`
	Findings  json.RawMessage `json:"findings"`
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
	Total    int `json:"total"`
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
	Skipped  int `json:"skipped"`
}

// VulnerabilitySummary provides aggregate vulnerability information.
type VulnerabilitySummary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Total    int `json:"total"`
}

// SBOMSummary provides aggregate SBOM information.
type SBOMSummary struct {
	TotalComponents int `json:"total_components"`
}

// EvidenceCollection holds all collected evidence for a module.
type EvidenceCollection struct {
	Module           string             `json:"module"`
	TestResults      *TestResults       `json:"test_results,omitempty"`
	SecurityResults  *SecurityResults   `json:"security_results,omitempty"`
	TestSummary      *TestSummary       `json:"test_summary,omitempty"`
	VulnSummary      *VulnerabilitySummary `json:"vuln_summary,omitempty"`
	SBOMSummary      *SBOMSummary       `json:"sbom_summary,omitempty"`
	CollectedAt      time.Time          `json:"collected_at"`
}

// EvidenceAgePolicy controls when to auto-run evidence collection.
type EvidenceAgePolicy struct {
	MaxAge        time.Duration `json:"max_age"`
	ForceTests    bool          `json:"force_tests"`
	ForceSecurity bool          `json:"force_security"`
	SkipAutoRun   bool          `json:"skip_auto_run"`
}

// DefaultEvidenceAgePolicy returns the default evidence age policy (24 hours).
func DefaultEvidenceAgePolicy() EvidenceAgePolicy {
	return EvidenceAgePolicy{
		MaxAge:        24 * time.Hour,
		ForceTests:    false,
		ForceSecurity: false,
		SkipAutoRun:   false,
	}
}

// HasAnyEvidence returns true if the collection has any evidence.
func (ec *EvidenceCollection) HasAnyEvidence() bool {
	if ec.TestResults != nil && (len(ec.TestResults.UnitTestFiles) > 0 || len(ec.TestResults.AcceptanceFiles) > 0) {
		return true
	}
	if ec.SecurityResults != nil && (ec.SecurityResults.VulnFile != "" || ec.SecurityResults.SBOMFile != "") {
		return true
	}
	return false
}

// HasTestEvidence returns true if test evidence exists.
func (ec *EvidenceCollection) HasTestEvidence() bool {
	return ec.TestResults != nil && (len(ec.TestResults.UnitTestFiles) > 0 || len(ec.TestResults.AcceptanceFiles) > 0)
}

// HasSecurityEvidence returns true if security evidence exists.
func (ec *EvidenceCollection) HasSecurityEvidence() bool {
	return ec.SecurityResults != nil && (ec.SecurityResults.VulnFile != "" || ec.SecurityResults.SBOMFile != "")
}

// LatestModTime returns the most recent modification time of any evidence.
func (ec *EvidenceCollection) LatestModTime() time.Time {
	var latest time.Time

	if ec.TestResults != nil && ec.TestResults.LastModified.After(latest) {
		latest = ec.TestResults.LastModified
	}
	if ec.SecurityResults != nil && ec.SecurityResults.LastModified.After(latest) {
		latest = ec.SecurityResults.LastModified
	}

	return latest
}

// ControlTestEvidence represents evidence for a specific control from tests.
type ControlTestEvidence struct {
	ControlID     string         `json:"control_id"`
	Tests         []TestEvidence `json:"tests"`
	TotalTests    int            `json:"total_tests"`
	PassedTests   int            `json:"passed_tests"`
	FailedTests   int            `json:"failed_tests"`
	SkippedTests  int            `json:"skipped_tests"`
	HasEvidence   bool           `json:"has_evidence"`
}

// TestEvidence represents a single test covering a control.
type TestEvidence struct {
	FeaturePath  string `json:"feature_path"`
	FeatureName  string `json:"feature_name"`
	ScenarioName string `json:"scenario_name"`
	Status       string `json:"status"` // "passed", "failed", "skipped", "not-run"
	Location     string `json:"location"` // file:line or file:scenario
}
