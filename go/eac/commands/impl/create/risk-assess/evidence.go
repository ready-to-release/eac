// File: go/eac/commands/impl/create/risk-assess/evidence.go
// Purpose: Evidence collection for risk assessments
//
// This file contains all functions related to collecting test and security
// evidence for module assessments.

package riskassess

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/impl/internal"
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

	// Collect test evidence from test manifest (read-only)
	cfg, _ := coreconfig.Load(coreconfig.LoadOptions{RepoRoot: config.WorkspaceRoot})
	if cfg != nil {
		testDir := cfg.Repository.TestModuleDirAbs(config.WorkspaceRoot, moduleName)
		manifest, err := internal.LoadTestManifest(testDir)

		if err == nil && manifest != nil {
			// Check test evidence age
			age := time.Since(manifest.TestTime)
			if !manifest.TestTime.IsZero() && age >= config.MaxEvidenceAge {
				relPath, err := filepath.Rel(config.WorkspaceRoot, testDir)
				if err != nil {
					relPath = testDir
				}

				warning := fmt.Sprintf(
					"⚠️  Test evidence is too old (Age: %s, max: %s) - Location: %s - Latest test run: %s",
					formatDuration(age),
					formatDuration(config.MaxEvidenceAge),
					relPath,
					manifest.TestTime.Format("2006-01-02 15:04:05 MST"),
				)
				collection.Warnings = append(collection.Warnings, warning)
			}

			// Convert to simplified manifest data
			collection.TestManifestData = convertToManifestData(manifest)
			collection.TestSummary = &evidence.TestSummary{
				Total:   manifest.Summary.Total,
				Passed:  manifest.Summary.Passed,
				Failed:  manifest.Summary.Failed,
				Skipped: manifest.Summary.Skipped,
			}
		} else {
			// No test manifest found - add warning
			relPath, err := filepath.Rel(config.WorkspaceRoot, testDir)
			if err != nil {
				relPath = testDir
			}
			warning := fmt.Sprintf(
				"⚠️  No test manifest found for module '%s' - Expected location: %s/test.manifest.json",
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
			// Get the actual directory that was checked using config-based paths
			cfg, err := coreconfig.Load(coreconfig.LoadOptions{RepoRoot: config.WorkspaceRoot})
			if err == nil {
				scanBaseDir := cfg.Repository.ScanModuleOutputPathAbs(config.WorkspaceRoot, moduleName)
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
			scanDir := cfg.Repository.ScanModuleOutputPathAbs(config.WorkspaceRoot, moduleName)
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

// convertToManifestData converts internal.TestManifest to simplified evidence.TestManifestData
func convertToManifestData(manifest *internal.TestManifest) *evidence.TestManifestData {
	if manifest == nil {
		return nil
	}

	// Convert test entries
	tests := make([]evidence.TestEntryData, len(manifest.Tests))
	for i, test := range manifest.Tests {
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
	for name, suite := range manifest.Suites {
		suites[name] = evidence.SuiteResultData{
			RunTime:         suite.RunTime,
			DurationSeconds: suite.DurationSeconds,
			Total:           suite.Tests.Total,
			Passed:          suite.Tests.Passed,
			Failed:          suite.Tests.Failed,
			Skipped:         suite.Tests.Skipped,
		}
	}

	// Convert artifacts
	artifacts := make([]evidence.TestArtifactData, len(manifest.Artifacts))
	for i, artifact := range manifest.Artifacts {
		artifacts[i] = evidence.TestArtifactData{
			Type: artifact.Type,
			Name: artifact.Name,
			Path: artifact.Path,
		}
	}

	return &evidence.TestManifestData{
		TestID:    manifest.TestID,
		TestAgent: manifest.TestAgent,
		GitCommit: manifest.GitCommit,
		BuildID:   manifest.BuildID,
		TestTime:  manifest.TestTime,
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
