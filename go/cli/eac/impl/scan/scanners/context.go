// Package scanners provides scan handlers using the pluggable tool system.
//
// This package delegates all handler registration and lookup to tool.GlobalScanBridge().
// Existing scanners register via RegisterScanner() in their init() functions,
// and the bridge integrates them with YAML-defined tools from tool-config.yml.
package scanners

import (
	"github.com/ready-to-release/eac/go/cli/eac/impl/scan/internal"
)

// ScanContext provides execution-time configuration for scanners.
// It's populated by the framework before calling scanner functions.
// This allows native scanners to access configuration that's only
// available at scan execution time (e.g., Docker images from eac-security contract).
type ScanContext struct {
	// Docker images from eac-security contract
	TrivyImage   string
	SemgrepImage string
	ZAPImage     string

	// Scanner-specific options
	SBOMFormat         string            // SBOM output format (e.g., "cyclonedx-json", "spdx-json")
	VulnSeverities     []internal.Severity // Severity filter for vuln scans
	SemgrepConfig      string            // Semgrep ruleset config
	ComplianceStandard string            // Compliance standard for compliance scans

	// ZAP-specific options
	ZAPTargetURL string // Target URL for DAST scanning
	ZAPScanType  string // Scan type: baseline, full, api

	// Execution context
	WorkspaceRoot string
	GitCommit     string
}

// GlobalScanContext is populated by the framework before scan execution.
// This allows native scanners to access execution-time configuration.
//
// Usage pattern:
//   1. Framework populates GlobalScanContext before calling scanner
//   2. Scanner adapter reads from GlobalScanContext
//   3. Framework clears GlobalScanContext after scanner completes
//
// Thread safety: The framework serializes scanner execution per module,
// so concurrent access to GlobalScanContext is not a concern for the
// current single-threaded-per-module design.
var GlobalScanContext *ScanContext
