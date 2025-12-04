// Command: create risk-assess
// Short: Update OSCAL assessment-results with test and security evidence
// Long: The create risk-assess command creates or updates OSCAL assessment-results for a module
// Long: by collecting test results and security scan evidence. It maps @control tags in
// Long: feature files to OSCAL control IDs and determines satisfied/not-satisfied status.
// Long:
// Long: Evidence is automatically collected from:
// Long: - Test results: out/test/<timestamp>/<module>/*.cucumber.json
// Long: - Security scans: out/security/<module>/**/*.json
// Long:
// Long: If evidence is older than 24 hours, tests and scans are automatically re-run.
// Flag.max-evidence-age: type=string, default=24h, usage=Maximum age for evidence before auto-refresh (e.g., 24h, 7d)
// Flag.force-tests: type=bool, default=false, usage=Force re-run tests regardless of age
// Flag.force-security: type=bool, default=false, usage=Force re-run security scans regardless of age
// Flag.skip-auto-run: type=bool, default=false, usage=Use existing evidence only, fail if missing
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs to out/logs/risk/
// Args: modules
package risk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/evidence"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/scoring"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

var assessLog = logging.C()

func init() {
	registry.Register(CreateRiskAssess)
}

// AssessConfig holds configuration for risk assess command.
type AssessConfig struct {
	Module         string
	MaxEvidenceAge time.Duration
	ForceTests     bool
	ForceSecurity  bool
	SkipAutoRun    bool
	Debug          bool
	WorkspaceRoot  string
	Logger         *logging.Logger
}

// CreateRiskAssess is the entry point for the create risk-assess command.
func CreateRiskAssess() int {
	config, err := parseAssessConfig()
	if err != nil {
		if err.Error() == "help requested" {
			showAssessHelp()
			return 0
		}
		assessLog.Errorf("Error: %v", err)
		return 1
	}

	// Initialize logger
	var logger *logging.Logger
	if config.Debug {
		logger, err = logging.NewWithDebug("risk", config.WorkspaceRoot)
	} else {
		logger, err = logging.NewDefault("risk", config.WorkspaceRoot)
	}
	if err != nil {
		assessLog.Errorf("Warning: Failed to initialize logger: %v", err)
	} else {
		config.Logger = logger
		defer logger.Sync()
	}

	if config.Logger != nil {
		config.Logger.Info("Starting risk assess",
			zap.String("module", config.Module),
			zap.Duration("max_evidence_age", config.MaxEvidenceAge),
			zap.Bool("debug", config.Debug))
	}

	// Load profile for the module
	profilePath := oscal.GetProfilePath(config.WorkspaceRoot, config.Module)
	profile, err := oscal.LoadProfile(profilePath)
	if err != nil {
		assessLog.Errorf("Error: Profile not found for module '%s'", config.Module)
		assessLog.Errorf("Expected at: %s", profilePath)
		assessLog.Error("")
		assessLog.Error("Create a profile first:")
		assessLog.Errorf("  create risk <assessment.md> --module %s", config.Module)
		return 1
	}

	controlIDs := profile.GetControlIDs()
	assessLog.Infof("Loaded profile with %d controls: %s", len(controlIDs), strings.Join(controlIDs, ", "))

	// Collect evidence
	evidenceCollection, err := collectEvidence(config)
	if err != nil {
		assessLog.Errorf("Error collecting evidence: %v", err)
		return 1
	}

	// Build or update assessment-results
	arPath := oscal.GetAssessmentResultsPath(config.WorkspaceRoot, config.Module)
	ar, err := buildAssessmentResults(config, profile, evidenceCollection)
	if err != nil {
		assessLog.Errorf("Error building assessment results: %v", err)
		return 1
	}

	// Write assessment-results
	if err := oscal.WriteAssessmentResults(arPath, ar); err != nil {
		assessLog.Errorf("Error writing assessment results: %v", err)
		return 1
	}

	// Report results
	reportResults(config, ar, arPath)

	if config.Logger != nil {
		config.Logger.Info("Assessment completed successfully",
			zap.String("path", arPath))
	}

	return 0
}

// parseAssessConfig parses command line configuration.
func parseAssessConfig() (*AssessConfig, error) {
	args := os.Args[3:] // Skip program name, "create", and "risk-assess"

	config := &AssessConfig{
		MaxEvidenceAge: 24 * time.Hour,
	}

	// Get workspace root
	workspaceRoot, err := registry.GetWorkspaceRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find workspace root: %w", err)
	}
	config.WorkspaceRoot = workspaceRoot

	// Parse flags and arguments
	i := 0
	for i < len(args) {
		arg := args[i]

		switch {
		case arg == "--help" || arg == "-h":
			return nil, fmt.Errorf("help requested")

		case arg == "--max-evidence-age":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--max-evidence-age requires a value")
			}
			duration, err := time.ParseDuration(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid duration: %s", args[i+1])
			}
			config.MaxEvidenceAge = duration
			i += 2

		case arg == "--force-tests":
			config.ForceTests = true
			i++

		case arg == "--force-security":
			config.ForceSecurity = true
			i++

		case arg == "--skip-auto-run":
			config.SkipAutoRun = true
			i++

		case arg == "--debug" || arg == "-d":
			config.Debug = true
			i++

		case strings.HasPrefix(arg, "-"):
			return nil, fmt.Errorf("unknown flag: %s", arg)

		default:
			// Positional argument: module name
			if config.Module == "" {
				config.Module = arg
			}
			i++
		}
	}

	// Validate required arguments
	if config.Module == "" {
		return nil, fmt.Errorf("module name required")
	}

	return config, nil
}

// collectEvidence gathers test and security evidence for a module.
func collectEvidence(config *AssessConfig) (*evidence.EvidenceCollection, error) {
	collection := &evidence.EvidenceCollection{
		Module:      config.Module,
		CollectedAt: time.Now(),
	}

	policy := evidence.EvidenceAgePolicy{
		MaxAge:        config.MaxEvidenceAge,
		ForceTests:    config.ForceTests,
		ForceSecurity: config.ForceSecurity,
		SkipAutoRun:   config.SkipAutoRun,
	}

	// Collect test evidence
	testResults, err := collectTestEvidence(config, policy)
	if err != nil {
		if config.SkipAutoRun {
			assessLog.Warnf("No test evidence available: %v", err)
		} else {
			assessLog.Warnf("Test evidence collection failed: %v", err)
		}
	} else {
		collection.TestResults = testResults

		// Calculate test summary
		summary, _ := evidence.GetTestSummaryForModule(config.WorkspaceRoot, config.Module)
		collection.TestSummary = summary
	}

	// Collect security evidence
	securityResults, err := collectSecurityEvidence(config, policy)
	if err != nil {
		if config.SkipAutoRun {
			assessLog.Warnf("No security evidence available: %v", err)
		} else {
			assessLog.Warnf("Security evidence collection failed: %v", err)
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
			// Allow proceeding with no evidence when --skip-auto-run is used
			// This will result in all controls being marked as not-satisfied
			assessLog.Warn("⚠️  No evidence available for this module")
			assessLog.Warn("All controls will be marked as not-satisfied")
			return collection, nil
		}
		return nil, fmt.Errorf("no evidence found for module '%s'", config.Module)
	}

	return collection, nil
}

// collectTestEvidence collects test evidence, optionally running tests if stale.
func collectTestEvidence(config *AssessConfig, policy evidence.EvidenceAgePolicy) (*evidence.TestResults, error) {
	results, err := evidence.FindTestResultsForModule(config.WorkspaceRoot, config.Module)

	// Determine if we need to run tests
	needsRun := false
	if err != nil {
		needsRun = true
	} else if policy.ForceTests {
		needsRun = true
		assessLog.Info("Forcing test re-run...")
	} else if !evidence.IsEvidenceFresh(results.TestRunDirectory, policy.MaxAge) {
		needsRun = true
		age, _ := evidence.GetEvidenceAge(results.TestRunDirectory)
		assessLog.Infof("Test results are %s old, running tests...", formatDuration(age))
	}

	if needsRun {
		if policy.SkipAutoRun {
			return nil, fmt.Errorf("test results are stale and --skip-auto-run was specified")
		}

		// Run tests
		if err := runTests(config); err != nil {
			return nil, fmt.Errorf("auto-run tests failed: %w", err)
		}

		// Retry discovery
		results, err = evidence.FindTestResultsForModule(config.WorkspaceRoot, config.Module)
		if err != nil {
			return nil, fmt.Errorf("no test results after auto-run: %w", err)
		}
	}

	return results, nil
}

// collectSecurityEvidence collects security evidence, optionally running scans if stale.
func collectSecurityEvidence(config *AssessConfig, policy evidence.EvidenceAgePolicy) (*evidence.SecurityResults, error) {
	results, err := evidence.FindSecurityResultsForModule(config.WorkspaceRoot, config.Module)

	// Determine if we need to run scans
	needsRun := false
	if err != nil {
		needsRun = true
	} else if policy.ForceSecurity {
		needsRun = true
		assessLog.Info("Forcing security scan re-run...")
	} else if results.VulnFile != "" && !evidence.IsEvidenceFresh(results.VulnFile, policy.MaxAge) {
		needsRun = true
		age, _ := evidence.GetEvidenceAge(results.VulnFile)
		assessLog.Infof("Security evidence is %s old, running scans...", formatDuration(age))
	}

	if needsRun && !policy.SkipAutoRun {
		// Run security scans
		if err := runSecurityScans(config); err != nil {
			assessLog.Warnf("Security scan failed: %v", err)
			// Continue with partial results
		}

		// Retry discovery
		results, err = evidence.FindSecurityResultsForModule(config.WorkspaceRoot, config.Module)
	}

	return results, err
}

// runTests runs the test suite for a module.
func runTests(config *AssessConfig) error {
	assessLog.Infof("Running tests for %s...", config.Module)

	// Use the canonical binary path
	binaryPath := repository.CommandsBinaryPath(config.WorkspaceRoot)
	cmd := exec.Command(binaryPath, "test", config.Module)
	cmd.Dir = config.WorkspaceRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("test command failed: %w", err)
	}

	return nil
}

// runSecurityScans runs security scans for a module.
func runSecurityScans(config *AssessConfig) error {
	assessLog.Infof("Running security scans for %s...", config.Module)

	// Use the canonical binary path
	binaryPath := repository.CommandsBinaryPath(config.WorkspaceRoot)

	// Run vulnerability scan
	cmd := exec.Command(binaryPath, "security", "vuln", config.Module)
	cmd.Dir = config.WorkspaceRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run() // Don't fail on security scan errors

	// Run SBOM
	cmd = exec.Command(binaryPath, "security", "sbom", config.Module)
	cmd.Dir = config.WorkspaceRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	return nil
}

// buildAssessmentResults creates OSCAL assessment-results from profile and evidence.
func buildAssessmentResults(config *AssessConfig, profile *oscal.Profile, ec *evidence.EvidenceCollection) (*oscal.AssessmentResults, error) {
	controlIDs := profile.GetControlIDs()

	// Get control to test mapping
	controlTestMap, _ := evidence.GetControlToTestMapping(config.WorkspaceRoot, config.Module)

	// Create assessment-results
	arUUID := uuid.New().String()
	profilePath := oscal.GetProfilePath(config.WorkspaceRoot, config.Module)
	relativeProfilePath, _ := filepath.Rel(config.WorkspaceRoot, profilePath)

	ar := oscal.NewAssessmentResults(
		arUUID,
		fmt.Sprintf("Risk Assessment for %s", config.Module),
		relativeProfilePath,
	)

	// Create result with observations and findings
	result := oscal.Result{
		UUID:        uuid.New().String(),
		Title:       fmt.Sprintf("%s Assessment", config.Module),
		Description: "Assessment-results generated from test and security evidence",
		Start:       time.Now().UTC().Format(time.RFC3339),
		Props: []oscal.Prop{
			{Name: "assessment-type", Value: "automated"},
		},
		ReviewedControls: &oscal.ReviewedControls{
			ControlSelections: []oscal.ControlSelection{
				{
					IncludeControls: make([]oscal.ControlRef, len(controlIDs)),
				},
			},
		},
	}

	// Add control refs
	for i, id := range controlIDs {
		result.ReviewedControls.ControlSelections[0].IncludeControls[i] = oscal.ControlRef{
			ControlID: id,
		}
	}

	// Create observations from evidence
	observations := createObservations(config, ec)
	result.Observations = observations

	// Create findings for each control
	findings := createFindings(config, controlIDs, controlTestMap, ec, observations)
	result.Findings = findings

	ar.AddResult(result)
	return ar, nil
}

// createObservations creates OSCAL observations from evidence.
func createObservations(config *AssessConfig, ec *evidence.EvidenceCollection) []oscal.Observation {
	var observations []oscal.Observation
	now := time.Now().UTC().Format(time.RFC3339)

	// Create observation for test evidence
	if ec.TestResults != nil {
		testObs := oscal.Observation{
			UUID:        uuid.New().String(),
			Title:       "Test Results",
			Description: fmt.Sprintf("Test evidence for module %s", config.Module),
			Methods:     []string{oscal.MethodTestAutomated},
			Collected:   now,
		}

		// Add evidence references
		for _, file := range ec.TestResults.AcceptanceFiles {
			relPath, _ := filepath.Rel(config.WorkspaceRoot, file)
			testObs.RelevantEvidence = append(testObs.RelevantEvidence, oscal.RelevantEvidence{
				Href:        relPath,
				Description: "Acceptance test results (Cucumber JSON)",
			})
		}

		for _, file := range ec.TestResults.UnitTestFiles {
			relPath, _ := filepath.Rel(config.WorkspaceRoot, file)
			testObs.RelevantEvidence = append(testObs.RelevantEvidence, oscal.RelevantEvidence{
				Href:        relPath,
				Description: "Unit test results",
			})
		}

		// Add summary props
		if ec.TestSummary != nil {
			testObs.Props = append(testObs.Props,
				oscal.Prop{Name: "total-tests", Value: fmt.Sprintf("%d", ec.TestSummary.Total)},
				oscal.Prop{Name: "passed-tests", Value: fmt.Sprintf("%d", ec.TestSummary.Passed)},
				oscal.Prop{Name: "failed-tests", Value: fmt.Sprintf("%d", ec.TestSummary.Failed)},
			)
		}

		observations = append(observations, testObs)
	}

	// Create observation for security evidence
	if ec.SecurityResults != nil {
		secObs := oscal.Observation{
			UUID:        uuid.New().String(),
			Title:       "Security Scan Results",
			Description: fmt.Sprintf("Security evidence for module %s", config.Module),
			Methods:     []string{oscal.MethodTestAutomated},
			Collected:   now,
		}

		if ec.SecurityResults.VulnFile != "" {
			relPath, _ := filepath.Rel(config.WorkspaceRoot, ec.SecurityResults.VulnFile)
			secObs.RelevantEvidence = append(secObs.RelevantEvidence, oscal.RelevantEvidence{
				Href:        relPath,
				Description: "Vulnerability scan results",
			})
		}

		if ec.SecurityResults.SBOMFile != "" {
			relPath, _ := filepath.Rel(config.WorkspaceRoot, ec.SecurityResults.SBOMFile)
			secObs.RelevantEvidence = append(secObs.RelevantEvidence, oscal.RelevantEvidence{
				Href:        relPath,
				Description: "Software Bill of Materials (SBOM)",
			})
		}

		// Add vulnerability summary props
		if ec.VulnSummary != nil {
			secObs.Props = append(secObs.Props,
				oscal.Prop{Name: "critical-vulns", Value: fmt.Sprintf("%d", ec.VulnSummary.Critical)},
				oscal.Prop{Name: "high-vulns", Value: fmt.Sprintf("%d", ec.VulnSummary.High)},
				oscal.Prop{Name: "medium-vulns", Value: fmt.Sprintf("%d", ec.VulnSummary.Medium)},
				oscal.Prop{Name: "low-vulns", Value: fmt.Sprintf("%d", ec.VulnSummary.Low)},
			)
		}

		observations = append(observations, secObs)
	}

	return observations
}

// createFindings creates OSCAL findings for each control.
func createFindings(config *AssessConfig, controlIDs []string, controlTestMap map[string][]string, ec *evidence.EvidenceCollection, observations []oscal.Observation) []oscal.Finding {
	var findings []oscal.Finding

	for _, controlID := range controlIDs {
		finding := oscal.Finding{
			UUID:        uuid.New().String(),
			Title:       fmt.Sprintf("Control %s Assessment", strings.ToUpper(controlID)),
			Description: fmt.Sprintf("Assessment finding for control %s", controlID),
			Target: oscal.Target{
				Type:     oscal.TargetTypeControlID,
				TargetID: controlID,
			},
		}

		// Determine status based on evidence
		state := determineControlStatus(controlID, controlTestMap, ec)
		finding.Target.Status = oscal.Status{
			State: state,
		}

		// Link to observations
		for _, obs := range observations {
			finding.RelatedObservations = append(finding.RelatedObservations, oscal.RelatedObservation{
				ObservationUUID: obs.UUID,
			})
		}

		// Add test coverage info
		if tests, ok := controlTestMap[controlID]; ok {
			finding.Props = append(finding.Props, oscal.Prop{
				Name:  "test-coverage",
				Value: fmt.Sprintf("%d tests", len(tests)),
			})
			finding.Remarks = fmt.Sprintf("Tested by: %s", strings.Join(tests, ", "))
		} else {
			finding.Props = append(finding.Props, oscal.Prop{
				Name:  "test-coverage",
				Value: "no tests",
			})
		}

		findings = append(findings, finding)
	}

	return findings
}

// determineControlStatus determines the satisfied/not-satisfied status for a control.
func determineControlStatus(controlID string, controlTestMap map[string][]string, ec *evidence.EvidenceCollection) string {
	// Check if we have tests for this control
	tests, hasTests := controlTestMap[controlID]

	// If no tests and no security evidence, status is not-satisfied
	if !hasTests && (ec.VulnSummary == nil || ec.VulnSummary.Total == 0) {
		return oscal.StateNotSatisfied
	}

	// If tests exist, check if they pass
	if hasTests && len(tests) > 0 {
		// If all tests pass and no critical/high vulnerabilities, satisfied
		if ec.TestSummary != nil && ec.TestSummary.Failed == 0 {
			if ec.VulnSummary == nil || (ec.VulnSummary.Critical == 0 && ec.VulnSummary.High == 0) {
				return oscal.StateSatisfied
			}
		}
	}

	// Default to not-satisfied if any concerns
	if ec.TestSummary != nil && ec.TestSummary.Failed > 0 {
		return oscal.StateNotSatisfied
	}

	if ec.VulnSummary != nil && (ec.VulnSummary.Critical > 0 || ec.VulnSummary.High > 0) {
		return oscal.StateNotSatisfied
	}

	// If we have passing tests, satisfied
	if ec.TestSummary != nil && ec.TestSummary.Passed > 0 {
		return oscal.StateSatisfied
	}

	return oscal.StateNotSatisfied
}

// reportResults prints assessment summary.
func reportResults(config *AssessConfig, ar *oscal.AssessmentResults, arPath string) {
	assessLog.Info("")
	assessLog.Info("═══════════════════════════════════════════════════════════")
	assessLog.Infof("  Assessment Results: %s", config.Module)
	assessLog.Info("═══════════════════════════════════════════════════════════")
	assessLog.Info("")

	if len(ar.AssessmentResults.Results) > 0 {
		result := ar.AssessmentResults.Results[0]

		satisfied := 0
		notSatisfied := 0
		for _, finding := range result.Findings {
			if finding.Target.Status.State == oscal.StateSatisfied {
				satisfied++
			} else {
				notSatisfied++
			}
		}

		total := len(result.Findings)
		assessLog.Infof("  Controls assessed: %d", total)
		assessLog.Infof("  ✓ Satisfied:       %d", satisfied)
		assessLog.Infof("  ✗ Not satisfied:   %d", notSatisfied)

		if notSatisfied > 0 {
			assessLog.Info("")
			assessLog.Info("  Controls needing attention:")
			for _, finding := range result.Findings {
				if finding.Target.Status.State == oscal.StateNotSatisfied {
					assessLog.Infof("    - %s", finding.Target.TargetID)
				}
			}
		}

		// Calculate risk score
		likelihood := scoring.CalculateBaseLikelihood(0, 0, 0, notSatisfied)
		impact := scoring.GetDefaultImpact("service")
		riskScore := scoring.ComputeRiskScore(config.Module, likelihood, impact, "Based on control satisfaction")

		assessLog.Info("")
		assessLog.Infof("  Risk Score: %s", scoring.FormatRiskScore(riskScore))
	}

	assessLog.Info("")
	assessLog.Infof("  Output: %s", arPath)
	assessLog.Info("")
	assessLog.Info("  Next steps:")
	assessLog.Infof("    show risk-report              # View aggregated report")
	assessLog.Infof("    validate risk %s  # Validate OSCAL", filepath.Base(arPath))
	assessLog.Info("")
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%d days", days)
	}
	hours := int(d.Hours())
	if hours > 0 {
		return fmt.Sprintf("%d hours", hours)
	}
	return fmt.Sprintf("%d minutes", int(d.Minutes()))
}

// showAssessHelp displays help information.
func showAssessHelp() {
	help := `Usage: create risk-assess <module> [flags]

Update OSCAL assessment-results with test and security evidence

Arguments:
  module               Module name to assess

Flags:
      --max-evidence-age <duration>  Maximum age for evidence (default: 24h)
      --force-tests                  Force re-run tests regardless of age
      --force-security               Force re-run security scans regardless of age
      --skip-auto-run                Use existing evidence only, fail if missing
  -d, --debug                        Save intermediate outputs to out/logs/risk/
  -h, --help                         Show this help message

Output:
  out/risk/<module>/assessment-results.json

Examples:
  # Assess a module (auto-refreshes stale evidence)
  create risk-assess billing-service

  # Force fresh test results
  create risk-assess api-gateway --force-tests

  # Use existing evidence only
  create risk-assess my-module --skip-auto-run

  # Set custom evidence freshness threshold
  create risk-assess my-module --max-evidence-age 1h
`
	assessLog.Info(help)
}
