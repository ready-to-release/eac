// Package evidence provides security evidence loading for risk assessment.
package evidence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/core/config"
)

// ScannerType represents the type of security scanner.
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

// FindLatestSecurityScan finds the security scan file for a module and scanner type.
// Scan files are stored in: <scan_path>/<module>/<scanner>.json
// Uses config.Repository.Paths.Out.Scan to determine the scan output directory.
func FindLatestSecurityScan(workspaceRoot, moduleName string, scannerType ScannerType) (string, error) {
	// Load config to get scan output path (respects repository contracts)
	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	// Construct scanner file path: out/scan/<module>/<scanner>.json
	scanFile := filepath.Join(cfg.Repository.ScanModuleOutputPathAbs(workspaceRoot, moduleName), string(scannerType)+".json")

	// Check if file exists
	if _, err := os.Stat(scanFile); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no %s scan found for module '%s'", scannerType, moduleName)
		}
		return "", fmt.Errorf("failed to access %s scan for module '%s': %w", scannerType, moduleName, err)
	}

	return scanFile, nil
}

// FindSecurityResultsForModule discovers latest security scan results for a module.
// Uses best-effort approach: returns partial results if some scanners haven't run.
func FindSecurityResultsForModule(workspaceRoot, moduleName string) (*SecurityResults, error) {
	results := &SecurityResults{
		ModuleName: moduleName,
	}

	// Track if we found ANY evidence and the latest modification time
	foundAny := false
	var latestModTime time.Time

	// Find latest file for each scanner type
	// Best-effort: some scanners may not have been run for this module

	if path, err := FindLatestSecurityScan(workspaceRoot, moduleName, ScannerSBOM); err == nil {
		results.SBOMFile = path
		foundAny = true
		if modTime := getFileModTime(path); modTime.After(latestModTime) {
			latestModTime = modTime
		}
	}

	if path, err := FindLatestSecurityScan(workspaceRoot, moduleName, ScannerVuln); err == nil {
		results.VulnFile = path
		foundAny = true
		if modTime := getFileModTime(path); modTime.After(latestModTime) {
			latestModTime = modTime
		}
	}

	if path, err := FindLatestSecurityScan(workspaceRoot, moduleName, ScannerSecrets); err == nil {
		results.SecretsFile = path
		foundAny = true
		if modTime := getFileModTime(path); modTime.After(latestModTime) {
			latestModTime = modTime
		}
	}

	if path, err := FindLatestSecurityScan(workspaceRoot, moduleName, ScannerSAST); err == nil {
		results.SASTFile = path
		foundAny = true
		if modTime := getFileModTime(path); modTime.After(latestModTime) {
			latestModTime = modTime
		}
	}

	// Optional scanners (not all modules need these)
	if path, err := FindLatestSecurityScan(workspaceRoot, moduleName, ScannerCompliance); err == nil {
		results.ComplianceFile = path
		foundAny = true
		if modTime := getFileModTime(path); modTime.After(latestModTime) {
			latestModTime = modTime
		}
	}

	if path, err := FindLatestSecurityScan(workspaceRoot, moduleName, ScannerIaC); err == nil {
		results.IaCFile = path
		foundAny = true
		if modTime := getFileModTime(path); modTime.After(latestModTime) {
			latestModTime = modTime
		}
	}

	if path, err := FindLatestSecurityScan(workspaceRoot, moduleName, ScannerDAST); err == nil {
		results.ZAPFile = path
		foundAny = true
		if modTime := getFileModTime(path); modTime.After(latestModTime) {
			latestModTime = modTime
		}
	}

	results.LastModified = latestModTime

	// Return error only if NO evidence was found at all
	if !foundAny {
		return nil, fmt.Errorf("no security scan results found for module '%s'", moduleName)
	}

	return results, nil
}

// LoadSecurityEvidence loads and parses a security evidence file.
func LoadSecurityEvidence(evidencePath string) (*SecurityEvidenceFile, error) {
	data, err := os.ReadFile(evidencePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read evidence file: %w", err)
	}

	var evidence SecurityEvidenceFile
	if err := json.Unmarshal(data, &evidence); err != nil {
		return nil, fmt.Errorf("failed to parse evidence file: %w", err)
	}

	return &evidence, nil
}

// ParseVulnerabilitySummary extracts vulnerability counts from evidence.
func ParseVulnerabilitySummary(vulnEvidence *SecurityEvidenceFile) (*VulnerabilitySummary, error) {
	if vulnEvidence == nil {
		return nil, nil
	}

	summary := &VulnerabilitySummary{}

	// Parse findings to count severities
	// This handles the Trivy vulnerability format
	var findings map[string]interface{}
	if err := json.Unmarshal(vulnEvidence.Findings, &findings); err != nil {
		// Try array format
		var findingsArray []interface{}
		if err := json.Unmarshal(vulnEvidence.Findings, &findingsArray); err != nil {
			return summary, nil // Return empty summary if parsing fails
		}
		// Count from array format
		for _, finding := range findingsArray {
			if f, ok := finding.(map[string]interface{}); ok {
				countSeverity(f, summary)
			}
		}
		return summary, nil
	}

	// Handle Trivy Results format
	if results, ok := findings["Results"].([]interface{}); ok {
		for _, result := range results {
			if r, ok := result.(map[string]interface{}); ok {
				if vulns, ok := r["Vulnerabilities"].([]interface{}); ok {
					for _, vuln := range vulns {
						if v, ok := vuln.(map[string]interface{}); ok {
							countSeverity(v, summary)
						}
					}
				}
			}
		}
	}

	// Handle direct vulnerabilities array
	if vulns, ok := findings["vulnerabilities"].([]interface{}); ok {
		for _, vuln := range vulns {
			if v, ok := vuln.(map[string]interface{}); ok {
				countSeverity(v, summary)
			}
		}
	}

	summary.Total = summary.Critical + summary.High + summary.Medium + summary.Low
	return summary, nil
}

// ParseSBOMSummary extracts component counts from SBOM evidence.
func ParseSBOMSummary(sbomEvidence *SecurityEvidenceFile) (*SBOMSummary, error) {
	if sbomEvidence == nil {
		return nil, nil
	}

	summary := &SBOMSummary{}

	// Parse findings to count components
	// This handles the CycloneDX SBOM format from Trivy
	var findings map[string]interface{}
	if err := json.Unmarshal(sbomEvidence.Findings, &findings); err != nil {
		return summary, nil // Return empty summary if parsing fails
	}

	// Handle CycloneDX format
	if components, ok := findings["components"].([]interface{}); ok {
		summary.TotalComponents = len(components)
	}

	// Handle SPDX format
	if packages, ok := findings["packages"].([]interface{}); ok {
		summary.TotalComponents = len(packages)
	}

	return summary, nil
}

// countSeverity increments the appropriate severity count based on finding.
func countSeverity(finding map[string]interface{}, summary *VulnerabilitySummary) {
	severity := ""
	if s, ok := finding["Severity"].(string); ok {
		severity = strings.ToUpper(s)
	} else if s, ok := finding["severity"].(string); ok {
		severity = strings.ToUpper(s)
	}

	switch severity {
	case "CRITICAL":
		summary.Critical++
	case "HIGH":
		summary.High++
	case "MEDIUM":
		summary.Medium++
	case "LOW":
		summary.Low++
	}
}

// getFileModTime returns the modification time of a file.
func getFileModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// IsEvidenceFresh checks if evidence file is recent enough based on policy.
func IsEvidenceFresh(evidencePath string, maxAge time.Duration) bool {
	info, err := os.Stat(evidencePath)
	if err != nil {
		return false // Missing = not fresh
	}

	age := time.Since(info.ModTime())
	return age < maxAge
}

// GetEvidenceAge returns the age of evidence file.
func GetEvidenceAge(evidencePath string) (time.Duration, error) {
	info, err := os.Stat(evidencePath)
	if err != nil {
		return 0, err
	}
	return time.Since(info.ModTime()), nil
}

// CollectAllSecurityEvidence collects security evidence from all result files.
func CollectAllSecurityEvidence(results *SecurityResults) ([]Evidence, error) {
	var evidence []Evidence

	if results.VulnFile != "" {
		evidence = append(evidence, Evidence{
			Type:    EvidenceTypeVulnerabilityScan,
			Path:    results.VulnFile,
			ModTime: getFileModTime(results.VulnFile),
			Module:  results.ModuleName,
		})
	}

	if results.SBOMFile != "" {
		evidence = append(evidence, Evidence{
			Type:    EvidenceTypeSBOM,
			Path:    results.SBOMFile,
			ModTime: getFileModTime(results.SBOMFile),
			Module:  results.ModuleName,
		})
	}

	if results.SecretsFile != "" {
		evidence = append(evidence, Evidence{
			Type:    EvidenceTypeSecretsScan,
			Path:    results.SecretsFile,
			ModTime: getFileModTime(results.SecretsFile),
			Module:  results.ModuleName,
		})
	}

	if results.SASTFile != "" {
		evidence = append(evidence, Evidence{
			Type:    EvidenceTypeSASTScan,
			Path:    results.SASTFile,
			ModTime: getFileModTime(results.SASTFile),
			Module:  results.ModuleName,
		})
	}

	if results.ComplianceFile != "" {
		evidence = append(evidence, Evidence{
			Type:    EvidenceTypeComplianceScan,
			Path:    results.ComplianceFile,
			ModTime: getFileModTime(results.ComplianceFile),
			Module:  results.ModuleName,
		})
	}

	if results.IaCFile != "" {
		evidence = append(evidence, Evidence{
			Type:    EvidenceTypeIaCScan,
			Path:    results.IaCFile,
			ModTime: getFileModTime(results.IaCFile),
			Module:  results.ModuleName,
		})
	}

	if results.ZAPFile != "" {
		evidence = append(evidence, Evidence{
			Type:    EvidenceTypeDAST,
			Path:    results.ZAPFile,
			ModTime: getFileModTime(results.ZAPFile),
			Module:  results.ModuleName,
		})
	}

	return evidence, nil
}
