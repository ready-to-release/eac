// Package internal provides shared types and utilities for security scanners.
package internal

import (
	"encoding/json"
	"time"
)

// EvidenceFile represents the standardized security evidence format.
// All scanners write evidence files following this schema for consistency
// and regulatory compliance (traceability, integrity verification).
type EvidenceFile struct {
	Module    string          `json:"module"`    // Module moniker
	Scanner   string          `json:"scanner"`   // Scanner type (sbom, vuln, etc.)
	Timestamp string          `json:"timestamp"` // RFC3339 format for unambiguous timezone
	SHA256    string          `json:"sha256"`    // Hash of findings for integrity verification
	Findings  json.RawMessage `json:"findings"`  // Scanner-specific JSON output
}

// ScannerType represents the type of security scanner
type ScannerType string

const (
	ScannerSBOM       ScannerType = "sbom"
	ScannerVuln       ScannerType = "vuln"
	ScannerSecrets    ScannerType = "secrets"
	ScannerCompliance ScannerType = "compliance"
	ScannerIaC        ScannerType = "iac"
	ScannerSAST       ScannerType = "sast"
	ScannerDAST       ScannerType = "zap"
)

// ScanResult holds the outcome of a security scan
type ScanResult struct {
	Success      bool
	OutputPath   string // Path to evidence file
	ErrorMessage string
	ExitCode     int
}

// Severity levels for vulnerability scanning
type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// ParseSeverity converts a string to Severity type
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

// GetTimestamp returns an RFC3339-formatted timestamp
func GetTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// GetFilenameTimestamp returns a filesystem-safe timestamp for filenames
func GetFilenameTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15-04-05Z")
}
