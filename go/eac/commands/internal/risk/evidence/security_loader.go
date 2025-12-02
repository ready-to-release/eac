// Package evidence provides security evidence loading for risk assessment.
package evidence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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

// FindLatestSecurityScan finds the most recent security scan file for a module and scanner type.
// Security scan files are stored in: out/security/<module>/<scanner>/<timestamp>.json
// The timestamp format is sortable: "2006-01-02T15-04-05Z"
func FindLatestSecurityScan(workspaceRoot, moduleName string, scannerType ScannerType) (string, error) {
	// Construct scanner directory path
	scanDir := filepath.Join(workspaceRoot, "out", "security", moduleName, string(scannerType))

	// Read directory contents
	entries, err := os.ReadDir(scanDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no %s scans found for module '%s' (directory does not exist)", scannerType, moduleName)
		}
		return "", fmt.Errorf("failed to read %s scan directory for module '%s': %w", scannerType, moduleName, err)
	}

	// Collect all JSON files
	var jsonFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			jsonFiles = append(jsonFiles, entry.Name())
		}
	}

	if len(jsonFiles) == 0 {
		return "", fmt.Errorf("no %s evidence files found for module '%s'", scannerType, moduleName)
	}

	// Sort by filename (timestamp) in descending order to get latest
	// Timestamp format "2006-01-02T15-04-05Z" is lexicographically sortable
	sort.Slice(jsonFiles, func(i, j int) bool {
		return jsonFiles[i] > jsonFiles[j] // Descending: latest first
	})

	// Return full path to latest file
	latestFile := filepath.Join(scanDir, jsonFiles[0])
	return latestFile, nil
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
