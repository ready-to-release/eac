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

// preRunTestsForModules runs the test suite before assessment.
// Tests are always run to ensure fresh results. Security scans are cached.
func preRunTestsForModules(config *AssessConfig) error {
	if config.SkipAutoRun {
		assessLog.Info("Skipping test run (--skip-auto-run specified)")
		return nil
	}

	// Always run test suite to get latest test results
	assessLog.Infof("Running %s test suite...", config.TestSuite)

	if err := runTestSuite(config); err != nil {
		assessLog.Warnf("Test suite failed: %v", err)
		return fmt.Errorf("test suite failed: %w", err)
	}

	return nil
}

// collectEvidenceForModule gathers evidence for a specific module.
func collectEvidenceForModule(config *AssessConfig, moduleName string) (*evidence.EvidenceCollection, error) {
	collection := &evidence.EvidenceCollection{
		Module:      moduleName,
		CollectedAt: time.Now(),
	}

	policy := evidence.EvidenceAgePolicy{
		MaxAge:        config.MaxEvidenceAge,
		ForceTests:    config.ForceTests,
		ForceScan:     config.ForceScan,
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

		// Calculate test summary from suite results
		summary, _ := evidence.GetTestSummaryForModule(config.WorkspaceRoot, moduleName, config.TestSuite)
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

		// Calculate SBOM summary
		if securityResults.SBOMFile != "" {
			sbomEvidence, _ := evidence.LoadSecurityEvidence(securityResults.SBOMFile)
			sbomSummary, _ := evidence.ParseSBOMSummary(sbomEvidence)
			collection.SBOMSummary = sbomSummary
		}
	}

	if !collection.HasAnyEvidence() {
		assessLog.Infof("No evidence available for module '%s' (module may not have tests in suite '%s')", moduleName, config.TestSuite)
		assessLog.Info("Controls will be marked as not-satisfied")
		return collection, nil
	}

	return collection, nil
}

// collectTestEvidenceForModule collects test evidence for a specific module.
// Note: Tests are pre-run by preRunTestsForModules(), so this just loads existing evidence.
// If a module has no tests, returns nil (not an error).
func collectTestEvidenceForModule(config *AssessConfig, moduleName string, policy evidence.EvidenceAgePolicy) (*evidence.TestResults, error) {
	// Look for test results in the suite directory (e.g., out/test/acceptance/<module>/)
	results, err := evidence.FindTestResultsForModuleInSuite(config.WorkspaceRoot, moduleName, config.TestSuite)

	if err != nil {
		// Module has no test results - this is okay, many modules don't have tests in all suites
		assessLog.Debugf("No test results for %s in suite %s (module may not have tests)", moduleName, config.TestSuite)
		return nil, nil // Return nil results, not an error
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
	} else if policy.ForceScan {
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

// runTestSuite runs the test suite once for all modules.
// This is more efficient than running tests per-module and produces proper output structure.
func runTestSuite(config *AssessConfig) error {
	// Use the canonical binary path
	binaryPath := repository.CommandsBinaryPath(config.WorkspaceRoot)

	assessLog.Infof("Running %s test suite...", config.TestSuite)

	// Run test suite without module filter (covers all modules)
	cmd := exec.Command(binaryPath, "test", "--suite", config.TestSuite)
	cmd.Dir = config.WorkspaceRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("test suite failed: %w", err)
	}

	return nil
}

// runTestsForModule runs the test suite for a specific module.
// Uses the specified test suite (default: acceptance) to run tests.
// Called sequentially from preRunTestsForModules, so no locking needed.
func runTestsForModule(config *AssessConfig, moduleName string) error {
	// Use the canonical binary path
	binaryPath := repository.CommandsBinaryPath(config.WorkspaceRoot)

	assessLog.Infof("Running %s tests for %s...", config.TestSuite, moduleName)
	assessLog.Debugf("Test command: %s test --suite %s %s", binaryPath, config.TestSuite, moduleName)

	// Run test with specified suite (--suite must come before module name)
	cmd := exec.Command(binaryPath, "test", "--suite", config.TestSuite, moduleName)
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
