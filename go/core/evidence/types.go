// Package evidence provides security scan evidence types and integrity verification.
//
// All scanners write evidence files following this schema for consistency
// and regulatory compliance (traceability, integrity verification).
package evidence

import (
	"encoding/json"
	"time"

	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/tool"
)

// File represents the standardized security evidence format.
// All scanners write evidence files following this schema for consistency
// and regulatory compliance (traceability, integrity verification).
type File struct {
	Module    string          `json:"module"`    // Module moniker
	Scanner   string          `json:"scanner"`   // Scanner tool ID (trivy-sbom, trivy-vuln, etc.)
	Timestamp string          `json:"timestamp"` // RFC3339 format for unambiguous timezone
	SHA256    string          `json:"sha256"`    // Hash of findings for integrity verification
	Findings  json.RawMessage `json:"findings"`  // Scanner-specific JSON output
}

// ScannerType is a string alias for scanner tool IDs.
type ScannerType string

// Well-known scanner tool IDs from tool-config.yml.
// These use centralized constants from the tool package.
const (
	ScannerSBOM       ScannerType = ScannerType(tool.ToolTrivySBOM)
	ScannerVuln       ScannerType = ScannerType(tool.ToolTrivyVuln)
	ScannerSecrets    ScannerType = ScannerType(tool.ToolTrivySecrets)
	ScannerCompliance ScannerType = ScannerType(tool.ToolTrivyCompliance)
	ScannerIaC        ScannerType = ScannerType(tool.ToolTrivyIaC)
	ScannerSAST       ScannerType = ScannerType(tool.ToolSemgrep)
	ScannerDAST       ScannerType = ScannerType(tool.ToolZap)
)

// ValidScannerCategories returns all valid scanner category names.
func ValidScannerCategories() map[string]bool {
	return domain.ValidScannerCategories()
}

// IsValidScannerCategory returns true if the category is valid.
func IsValidScannerCategory(category string) bool {
	return domain.IsValidScannerCategory(category)
}

// ScanResult holds the outcome of a security scan.
type ScanResult struct {
	Success      bool
	OutputPath   string // Path to evidence file
	ErrorMessage string
	ExitCode     int
}

// Severity levels for vulnerability scanning.
type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// ParseSeverity converts a string to Severity type.
func ParseSeverity(s string) (Severity, bool) {
	switch s {
	case "LOW":
		return SeverityLow, true
	case "MEDIUM":
		return SeverityMedium, true
	case "HIGH":
		return SeverityHigh, true
	case "CRITICAL":
		return SeverityCritical, true
	default:
		return "", false
	}
}

// GetTimestamp returns an RFC3339-formatted timestamp.
func GetTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// GetFilenameTimestamp returns a filesystem-safe timestamp for filenames.
func GetFilenameTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15-04-05Z")
}
