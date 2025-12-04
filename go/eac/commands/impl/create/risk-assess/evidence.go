// File: go/eac/commands/impl/create/risk-assess/evidence.go
// Purpose: Evidence collection for risk assessments
//
// This file contains all functions related to collecting test and security
// evidence for module assessments.

package riskassess

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/evidence"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// collectEvidenceForModule gathers evidence for a specific module.
func collectEvidenceForModule(config *AssessConfig, moduleName string) (*evidence.EvidenceCollection, error) {
	collection := &evidence.EvidenceCollection{
		Module:      moduleName,
		CollectedAt: time.Now(),
	}

	policy := evidence.EvidenceAgePolicy{
		MaxAge:        config.MaxEvidenceAge,
		ForceTests:    config.ForceTests,
		ForceSecurity: config.ForceSecurity,
		SkipAutoRun:   config.SkipAutoRun,
	}

	// Collect test evidence
	testResults, err := collectTestEvidenceForModule(config, moduleName, policy)
	if err != nil {
		if config.SkipAutoRun {
			assessLog.Warnf("No test evidence available for %s: %v", moduleName, err)
		} else {
			assessLog.Warnf("Test evidence collection failed for %s: %v", moduleName, err)
		}
	} else {
		collection.TestResults = testResults

		// Calculate test summary
		summary, _ := evidence.GetTestSummaryForModule(config.WorkspaceRoot, moduleName)
		collection.TestSummary = summary
	}

	// Collect security evidence
	securityResults, err := collectSecurityEvidenceForModule(config, moduleName, policy)
	if err != nil {
		if config.SkipAutoRun {
			assessLog.Warnf("No security evidence available for %s: %v", moduleName, err)
		} else {
			assessLog.Warnf("Security evidence collection failed for %s: %v", moduleName, err)
		}
	} else {
		collection.SecurityResults = securityResults

		// Calculate vulnerability summary
		if securityResults.VulnFile != "" {
			vulnEvidence, _ := evidence.LoadSecurityEvidence(securityResults.VulnFile)
			vulnSummary, _ := evidence.ParseVulnerabilitySummary(vulnEvidence)
			collection.VulnSummary = vulnSummary
		}
	}

	if !collection.HasAnyEvidence() {
		if config.SkipAutoRun {
			assessLog.Warnf("⚠️  No evidence available for module '%s'", moduleName)
			assessLog.Warn("Controls will be marked as not-satisfied")
			return collection, nil
		}
		return nil, fmt.Errorf("no evidence found for module '%s'", moduleName)
	}

	return collection, nil
}

// collectTestEvidenceForModule collects test evidence for a specific module.
func collectTestEvidenceForModule(config *AssessConfig, moduleName string, policy evidence.EvidenceAgePolicy) (*evidence.TestResults, error) {
	results, err := evidence.FindTestResultsForModule(config.WorkspaceRoot, moduleName)

	// Determine if we need to run tests
	needsRun := false
	if err != nil {
		needsRun = true
	} else if policy.ForceTests {
		needsRun = true
		assessLog.Infof("Forcing test re-run for %s...", moduleName)
	} else if !evidence.IsEvidenceFresh(results.TestRunDirectory, policy.MaxAge) {
		needsRun = true
		age, _ := evidence.GetEvidenceAge(results.TestRunDirectory)
		assessLog.Infof("Test results for %s are %s old, running tests...", moduleName, formatDuration(age))
	}

	if needsRun {
		if policy.SkipAutoRun {
			return nil, fmt.Errorf("test results are stale and --skip-auto-run was specified")
		}

		// Run tests
		if err := runTestsForModule(config, moduleName); err != nil {
			return nil, fmt.Errorf("auto-run tests failed: %w", err)
		}

		// Retry discovery
		results, err = evidence.FindTestResultsForModule(config.WorkspaceRoot, moduleName)
		if err != nil {
			return nil, fmt.Errorf("no test results after auto-run: %w", err)
		}
	}

	return results, nil
}

// collectSecurityEvidenceForModule collects security evidence for a specific module.
func collectSecurityEvidenceForModule(config *AssessConfig, moduleName string, policy evidence.EvidenceAgePolicy) (*evidence.SecurityResults, error) {
	results, err := evidence.FindSecurityResultsForModule(config.WorkspaceRoot, moduleName)

	// Determine if we need to run scans
	needsRun := false
	if err != nil {
		needsRun = true
	} else if policy.ForceSecurity {
		needsRun = true
		assessLog.Infof("Forcing security scan re-run for %s...", moduleName)
	} else if results.VulnFile != "" && !evidence.IsEvidenceFresh(results.VulnFile, policy.MaxAge) {
		needsRun = true
		age, _ := evidence.GetEvidenceAge(results.VulnFile)
		assessLog.Infof("Security evidence for %s is %s old, running scans...", moduleName, formatDuration(age))
	}

	if needsRun && !policy.SkipAutoRun {
		// Run security scans
		if err := runSecurityScansForModule(config, moduleName); err != nil {
			assessLog.Warnf("Security scan failed for %s: %v", moduleName, err)
			// Continue with partial results
		}

		// Retry discovery
		results, err = evidence.FindSecurityResultsForModule(config.WorkspaceRoot, moduleName)
	}

	return results, err
}

// runTestsForModule runs the test suite for a specific module.
func runTestsForModule(config *AssessConfig, moduleName string) error {
	assessLog.Infof("Running tests for %s...", moduleName)

	// Use the canonical binary path
	binaryPath := repository.CommandsBinaryPath(config.WorkspaceRoot)
	cmd := exec.Command(binaryPath, "test", moduleName)
	cmd.Dir = config.WorkspaceRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("test command failed: %w", err)
	}

	return nil
}

// runSecurityScansForModule runs security scans for a specific module.
func runSecurityScansForModule(config *AssessConfig, moduleName string) error {
	assessLog.Infof("Running security scans for %s...", moduleName)

	// Use the canonical binary path
	binaryPath := repository.CommandsBinaryPath(config.WorkspaceRoot)

	// Run vulnerability scan
	cmd := exec.Command(binaryPath, "security", "vuln", moduleName)
	cmd.Dir = config.WorkspaceRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run() // Don't fail on security scan errors

	// Run SBOM
	cmd = exec.Command(binaryPath, "security", "sbom", moduleName)
	cmd.Dir = config.WorkspaceRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	return nil
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}
