// File: go/cli/eac/impl/create/risk-assess/evidence.go
// Purpose: Evidence collection for risk assessments
//
// This file contains all functions related to collecting test and security
// evidence for module assessments.

package riskassess

import (
	"fmt"
	"path/filepath"
	"time"

	testview "github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/commands/repository/internal/risk/evidence"
	coreconfig "github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
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

	// Collect test evidence from UoW manifests (read-only)
	cfg, cfgErr := coreconfig.Load(coreconfig.LoadOptions{RepoRoot: config.WorkspaceRoot})
	if cfgErr == nil && cfg != nil {
		view, err := testview.LoadModuleTestView(config.WorkspaceRoot, moduleName)

		if err == nil && view != nil {
			// Check test evidence age
			age := time.Since(view.ExecutedAt)
			if !view.ExecutedAt.IsZero() && age >= config.MaxEvidenceAge {
				testDir := paths.TestModuleDir(config.WorkspaceRoot, moduleName)
				relPath, err := filepath.Rel(config.WorkspaceRoot, testDir)
				if err != nil {
					relPath = testDir
				}
				relPath = filepath.ToSlash(relPath) // Normalize to forward slashes

				warning := fmt.Sprintf(
					"⚠️  Test evidence is too old (Age: %s, max: %s) - Location: %s - Latest test run: %s",
					formatDuration(age),
					formatDuration(config.MaxEvidenceAge),
					relPath,
					view.ExecutedAt.Format("2006-01-02 15:04:05 MST"),
				)
				collection.Warnings = append(collection.Warnings, warning)
			}

			// Convert to simplified manifest data
			collection.TestManifestData = convertViewToManifestData(view)
			collection.TestSummary = &evidence.TestSummary{
				Total:   view.Summary.Total,
				Passed:  view.Summary.Passed,
				Failed:  view.Summary.Failed,
				Skipped: view.Summary.Skipped,
			}
		} else {
			// No test data found - add warning
			testDir := paths.TestModuleDir(config.WorkspaceRoot, moduleName)
			relPath, err := filepath.Rel(config.WorkspaceRoot, testDir)
			if err != nil {
				relPath = testDir
			}
			relPath = filepath.ToSlash(relPath) // Normalize to forward slashes
			warning := fmt.Sprintf(
				"⚠️  No test data found for module '%s' - Expected location: %s/",
				moduleName,
				relPath,
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
			scanBaseDir := paths.ScanModuleOutputPath(config.WorkspaceRoot, moduleName)
			relPath, err := filepath.Rel(config.WorkspaceRoot, scanBaseDir)
			if err != nil {
				relPath = scanBaseDir
			}
			relPath = filepath.ToSlash(relPath) // Normalize to forward slashes

			warning := fmt.Sprintf(
				"⚠️  Security evidence is too old (Age: %s, max: %s) - Location: %s - Latest modified: %s",
				formatDuration(age),
				formatDuration(config.MaxEvidenceAge),
				relPath,
				securityResults.LastModified.Format("2006-01-02 15:04:05 MST"),
			)
			collection.Warnings = append(collection.Warnings, warning)
		}

		// Use security evidence even if stale
		collection.SecurityResults = securityResults

		// Calculate vulnerability summary
		if securityResults.VulnFile != "" {
			if vulnEvidence, err := evidence.LoadSecurityEvidence(securityResults.VulnFile); err == nil {
				if vulnSummary, err := evidence.ParseVulnerabilitySummary(vulnEvidence); err == nil {
					collection.VulnSummary = vulnSummary
				}
			}
		}

		// Calculate SBOM summary
		if securityResults.SBOMFile != "" {
			if sbomEvidence, err := evidence.LoadSecurityEvidence(securityResults.SBOMFile); err == nil {
				if sbomSummary, err := evidence.ParseSBOMSummary(sbomEvidence); err == nil {
					collection.SBOMSummary = sbomSummary
				}
			}
		}
	} else {
		// No security results found - add warning
		scanDir := paths.ScanModuleOutputPath(config.WorkspaceRoot, moduleName)
		scanRelPath, err := filepath.Rel(config.WorkspaceRoot, scanDir)
		if err != nil {
			scanRelPath = scanDir
		}
		scanRelPath = filepath.ToSlash(scanRelPath) // Normalize to forward slashes
		warning := fmt.Sprintf(
			"⚠️  No security scan evidence found - Expected location: %s",
			scanRelPath,
		)
		collection.Warnings = append(collection.Warnings, warning)
	}

	// Determine if we have valid (non-empty, fresh) evidence
	hasTestEvidence := collection.TestManifestData != nil
	hasSecurityEvidence := collection.SecurityResults != nil

	// Track if evidence exists but is too old
	testTooOld := false
	securityTooOld := false

	if hasTestEvidence && !collection.TestManifestData.TestTime.IsZero() {
		age := time.Since(collection.TestManifestData.TestTime)
		testTooOld = age >= config.MaxEvidenceAge
	}

	if hasSecurityEvidence && !collection.SecurityResults.LastModified.IsZero() {
		age := time.Since(collection.SecurityResults.LastModified)
		securityTooOld = age >= config.MaxEvidenceAge
	}

	// Error if NO evidence exists at all (neither test nor security)
	if !hasTestEvidence && !hasSecurityEvidence {
		testDir := paths.TestModuleDir(config.WorkspaceRoot, moduleName)
		testRelPath, _ := filepath.Rel(config.WorkspaceRoot, testDir)
		testRelPath = filepath.ToSlash(testRelPath) // Normalize to forward slashes
		scanDir := paths.ScanModuleOutputPath(config.WorkspaceRoot, moduleName)
		scanRelPath, _ := filepath.Rel(config.WorkspaceRoot, scanDir)
		scanRelPath = filepath.ToSlash(scanRelPath) // Normalize to forward slashes

		return nil, fmt.Errorf(`no evidence found for module '%s'

Expected locations:
  - Test results: %s/
  - Security scans: %s/*.json

Run tests and scans to generate evidence:
  test %s
  scan %s`, moduleName, testRelPath, scanRelPath, moduleName, moduleName)
	}

	// Error if ALL available evidence is too old
	// Case 1: Only test evidence exists and it's too old
	// Case 2: Only security evidence exists and it's too old
	// Case 3: Both exist but both are too old

	// Check which evidence is too old
	onlyTestTooOld := hasTestEvidence && testTooOld && !hasSecurityEvidence
	onlySecurityTooOld := hasSecurityEvidence && securityTooOld && !hasTestEvidence
	bothTooOld := hasTestEvidence && hasSecurityEvidence && testTooOld && securityTooOld

	if onlyTestTooOld || onlySecurityTooOld || bothTooOld {
		testDir := paths.TestModuleDir(config.WorkspaceRoot, moduleName)
		testRelPath, _ := filepath.Rel(config.WorkspaceRoot, testDir)
		testRelPath = filepath.ToSlash(testRelPath) // Normalize to forward slashes
		scanDir := paths.ScanModuleOutputPath(config.WorkspaceRoot, moduleName)
		scanRelPath, _ := filepath.Rel(config.WorkspaceRoot, scanDir)
		scanRelPath = filepath.ToSlash(scanRelPath) // Normalize to forward slashes

		// Customize error message based on what's too old
		if onlyTestTooOld {
			return nil, fmt.Errorf(`test evidence for module '%s' is too old (max age: %s)

Evidence location:
  - Test results: %s/

Run tests to update evidence:
  test %s`, moduleName, formatDuration(config.MaxEvidenceAge), testRelPath, moduleName)
		} else if onlySecurityTooOld {
			return nil, fmt.Errorf(`security evidence for module '%s' is too old (max age: %s)

Evidence location:
  - Security scans: %s/

Run security scans to update evidence:
  scan %s`, moduleName, formatDuration(config.MaxEvidenceAge), scanRelPath, moduleName)
		} else {
			// Both are too old
			return nil, fmt.Errorf(`all evidence for module '%s' is too old (max age: %s)

Evidence locations:
  - Test results: %s/
  - Security scans: %s/

Run tests or scans to update evidence:
  test %s
  scan %s`, moduleName, formatDuration(config.MaxEvidenceAge), testRelPath, scanRelPath, moduleName, moduleName)
		}
	}

	// Have at least some fresh evidence - return with warnings for any stale evidence
	return collection, nil
}

// convertViewToManifestData converts a testview.TestModuleView to simplified evidence.TestManifestData.
func convertViewToManifestData(view *testview.TestModuleView) *evidence.TestManifestData {
	if view == nil {
		return nil
	}

	// Convert test entries
	tests := make([]evidence.TestEntryData, len(view.Tests))
	for i := range view.Tests {
		test := &view.Tests[i]
		tests[i] = evidence.TestEntryData{
			Name:     test.Name,
			Package:  test.Package,
			Type:     test.Type,
			Suite:    test.Suite,
			Status:   test.Status,
			Tags:     test.Tags,
			FilePath: test.FilePath,
		}
	}

	// Convert suites
	suites := make(map[string]evidence.SuiteResultData)
	for name, suite := range view.Suites {
		suites[name] = evidence.SuiteResultData{
			RunTime:         suite.RunTime,
			DurationSeconds: view.Duration.Seconds(),
			Total:           suite.Total,
			Passed:          suite.Passed,
			Failed:          suite.Failed,
			Skipped:         suite.Skipped,
		}
	}

	// Convert artifacts
	artifacts := make([]evidence.TestArtifactData, len(view.Artifacts))
	for i, artifact := range view.Artifacts {
		artifacts[i] = evidence.TestArtifactData{
			Type: artifact.Type,
			Name: filepath.Base(artifact.Path),
			Path: artifact.Path,
		}
	}

	return &evidence.TestManifestData{
		TestTime:  view.ExecutedAt,
		Tests:     tests,
		Suites:    suites,
		Artifacts: artifacts,
	}
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
