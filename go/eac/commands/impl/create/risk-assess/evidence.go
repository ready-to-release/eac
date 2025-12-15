// File: go/eac/commands/impl/create/risk-assess/evidence.go
// Purpose: Evidence collection for risk assessments
//
// This file contains all functions related to collecting test and security
// evidence for module assessments.

package riskassess

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/evidence"
	coreconfig "github.com/ready-to-release/eac/go/eac/core/config"
)

// collectEvidenceForModule gathers evidence for a specific module (read-only).
// This function only reads existing evidence and does NOT run tests or scans.
// It collects whatever evidence is available and records warnings for missing or stale evidence.
func collectEvidenceForModule(config *AssessConfig, moduleName string) (*evidence.EvidenceCollection, error) {
	collection := &evidence.EvidenceCollection{
		Module:      moduleName,
		CollectedAt: time.Now(),
		Warnings:    []string{},
	}

	// Collect test evidence from all specified suites (read-only)
	var foundTestResults *evidence.TestResults
	var checkedSuites []string

	for _, suite := range config.TestSuites {
		testResults, err := evidence.FindTestResultsForModuleInSuite(
			config.WorkspaceRoot,
			moduleName,
			suite,
		)
		checkedSuites = append(checkedSuites, suite)

		if err == nil && testResults != nil {
			// Check test evidence age
			age := time.Since(testResults.LastModified)
			if !testResults.LastModified.IsZero() && age >= config.MaxEvidenceAge {
				// Get the actual directory that was checked
				actualPath := testResults.TestRunDirectory
				relPath, err := filepath.Rel(config.WorkspaceRoot, actualPath)
				if err != nil {
					relPath = actualPath
				}

				warning := fmt.Sprintf(
					"⚠️  Test evidence for suite '%s' is too old (Age: %s, max: %s) - Location: %s - Latest modified: %s",
					suite,
					formatDuration(age),
					formatDuration(config.MaxEvidenceAge),
					relPath,
					testResults.LastModified.Format("2006-01-02 15:04:05 MST"),
				)
				collection.Warnings = append(collection.Warnings, warning)

				// Still use the stale evidence but warn about it
				if foundTestResults == nil {
					foundTestResults = testResults
					summary, _ := evidence.GetTestSummaryForModule(config.WorkspaceRoot, moduleName, suite)
					collection.TestSummary = summary
				}
			} else {
				// Use the first valid test results found
				if foundTestResults == nil {
					foundTestResults = testResults
					// Calculate test summary
					summary, _ := evidence.GetTestSummaryForModule(config.WorkspaceRoot, moduleName, suite)
					collection.TestSummary = summary
				}
			}
		}
	}

	if foundTestResults != nil {
		collection.TestResults = foundTestResults
	} else {
		// No test results found - add warning
		// Load config to get correct test output paths
		cfg, err := coreconfig.Load(coreconfig.LoadOptions{RepoRoot: config.WorkspaceRoot})
		if err == nil {
			suiteList := strings.Join(checkedSuites, ", ")
			var testLocations []string
			for _, suite := range checkedSuites {
				suiteDir := cfg.Repository.TestModuleOutputPathAbs(config.WorkspaceRoot, suite, moduleName)
				relPath, err := filepath.Rel(config.WorkspaceRoot, suiteDir)
				if err != nil {
					relPath = suiteDir
				}
				testLocations = append(testLocations, relPath)
			}
			warning := fmt.Sprintf(
				"⚠️  No test evidence found in suites: %s - Checked: %s",
				suiteList,
				strings.Join(testLocations, ", "),
			)
			collection.Warnings = append(collection.Warnings, warning)
		}
	}

	// Collect security evidence (read-only)
	securityResults, err := evidence.FindSecurityResultsForModule(config.WorkspaceRoot, moduleName)
	if err == nil && securityResults != nil {
		// Check security evidence age
		age := time.Since(securityResults.LastModified)
		if !securityResults.LastModified.IsZero() && age >= config.MaxEvidenceAge {
			// Get the actual directory that was checked using config-based paths
			cfg, err := coreconfig.Load(coreconfig.LoadOptions{RepoRoot: config.WorkspaceRoot})
			if err == nil {
				scanBaseDir := cfg.Repository.SecurityModuleOutputPathAbs(config.WorkspaceRoot, moduleName)
				relPath, err := filepath.Rel(config.WorkspaceRoot, scanBaseDir)
				if err != nil {
					relPath = scanBaseDir
				}

				warning := fmt.Sprintf(
					"⚠️  Security evidence is too old (Age: %s, max: %s) - Location: %s - Latest modified: %s",
					formatDuration(age),
					formatDuration(config.MaxEvidenceAge),
					relPath,
					securityResults.LastModified.Format("2006-01-02 15:04:05 MST"),
				)
				collection.Warnings = append(collection.Warnings, warning)
			}
		}

		// Use security evidence even if stale
		collection.SecurityResults = securityResults

		// Calculate vulnerability summary
		if securityResults.VulnFile != "" {
			vulnEvidence, _ := evidence.LoadSecurityEvidence(securityResults.VulnFile)
			vulnSummary, _ := evidence.ParseVulnerabilitySummary(vulnEvidence)
			collection.VulnSummary = vulnSummary
		}

		// Calculate SBOM summary
		if securityResults.SBOMFile != "" {
			sbomEvidence, _ := evidence.LoadSecurityEvidence(securityResults.SBOMFile)
			sbomSummary, _ := evidence.ParseSBOMSummary(sbomEvidence)
			collection.SBOMSummary = sbomSummary
		}
	} else {
		// No security results found - add warning using config-based paths
		cfg, err := coreconfig.Load(coreconfig.LoadOptions{RepoRoot: config.WorkspaceRoot})
		if err == nil {
			scanDir := cfg.Repository.SecurityModuleOutputPathAbs(config.WorkspaceRoot, moduleName)
			scanRelPath, err := filepath.Rel(config.WorkspaceRoot, scanDir)
			if err != nil {
				scanRelPath = scanDir
			}
			warning := fmt.Sprintf(
				"⚠️  No security scan evidence found - Expected location: %s",
				scanRelPath,
			)
			collection.Warnings = append(collection.Warnings, warning)
		}
	}

	// Always return the collection with whatever evidence was found
	// The warnings field will document any issues
	return collection, nil
}

// formatDuration formats a duration in a human-readable way with more precision.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f seconds", d.Seconds())
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		if seconds > 0 {
			return fmt.Sprintf("%d minutes %d seconds", minutes, seconds)
		}
		return fmt.Sprintf("%d minutes", minutes)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%d hours %d minutes", hours, minutes)
		}
		return fmt.Sprintf("%d hours", hours)
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%d days %d hours", days, hours)
	}
	return fmt.Sprintf("%d days", days)
}
