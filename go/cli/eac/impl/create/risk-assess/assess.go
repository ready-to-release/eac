package riskassess

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/evidence"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/scoring"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
)

type createRiskAssessCommand struct{}

var _ core.SimpleCommandPort = (*createRiskAssessCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&createRiskAssessCommand{},
	}
}

func (c *createRiskAssessCommand) Name() string { return "create risk-assess" }

func (c *createRiskAssessCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "create-risk-assess",
		Short:         "Create OSCAL assessment-results from existing test and security evidence",
		Long:          "The create risk-assess command creates OSCAL assessment-results for modules\nby reading existing test results and security scan evidence. It maps @control tags in\nfeature files to OSCAL control IDs and determines satisfied/not-satisfied status.\n\nThis command does NOT run tests or scans. It only reads existing evidence.\nThe command will warn (but continue) if:\n- Evidence is missing\n- Evidence is older than max-evidence-age (default: 24h)\n\nEvidence is collected from:\n- Test results: out/test/<module>/*.json\n- Security scans: out/scan/<module>/**/*.json\n\nExpected Output:\n- OSCAL assessment-results JSON file\n- Control status (satisfied/not-satisfied) based on test results\n- Risk assessment reports in Markdown format",
		Args:          "modules",
		Flags: []core.FlagSpec{
			{Name: "profile", Shorthand: "p", Type: "string", Required: true, Usage: "Path to OSCAL profile JSON file"},
			{Name: "max-evidence-age", Type: "string", DefaultValue: "24h", Usage: "Maximum age for evidence before warning (e.g., 24h, 7d)"},
			{Name: "suites", Type: "[]string", DefaultValue: "all", Usage: "Test suites to check for evidence (e.g., all, integration, acceptance)"},
			{Name: "sequential", Type: "bool", DefaultValue: "false", Usage: "Run assessments sequentially instead of parallel"},
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Save intermediate outputs to out/commands.log"},
		},
	}
}

func (c *createRiskAssessCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return CreateRiskAssess()
}

var assessLog = logging.C()

// AssessConfig holds configuration for risk assess command.
type AssessConfig struct {
	Modules        []string // Module names to assess
	ProfilePath    string
	MaxEvidenceAge time.Duration
	Sequential     bool // Disable parallel execution
	Debug          bool
	WorkspaceRoot  string
	OutputDir      string // Base output directory: out/risk/
}

// ModuleAssessmentResult holds the results of assessing a single module.
type ModuleAssessmentResult struct {
	Module            string
	AssessmentResults *oscalTypes.AssessmentResults
	Evidence          *evidence.EvidenceCollection // Evidence collection for reporting
	OutputPath        string
	Satisfied         int
	NotSatisfied      int
	RiskScore         *scoring.RiskScore
	Error             error
	Warnings          []string // Evidence collection warnings
}

// CreateRiskAssess is the entry point for the create risk-assess command.
func CreateRiskAssess() int {
	config, err := parseAssessConfig()
	if err != nil {
		assessLog.Errorf("Error: %v", err)
		return 1
	}

	// Initialize output directory with timestamp for uniqueness
	timestamp := time.Now().UTC().Format("20060102-150405")
	baseRiskDir := paths.RiskOutputPath(config.WorkspaceRoot, "")
	config.OutputDir = filepath.Join(baseRiskDir, timestamp)

	// Create timestamped output directory
	if err := os.MkdirAll(config.OutputDir, 0o755); err != nil {
		assessLog.Errorf("Failed to create output directory: %v", err)
		return 1
	}

	// Configure logging system (component loggers + file logging)
	if err := logging.ConfigureLoggingSimple(config.WorkspaceRoot, "commands", nil, config.Debug); err != nil {
		assessLog.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	assessLog.Info("Starting risk assess")
	assessLog.Debugf("Modules: modules=%v", config.Modules)
	assessLog.Debugf("Profile: %s", config.ProfilePath)
	assessLog.Debugf("Max evidence age: %v", config.MaxEvidenceAge)
	assessLog.Debugf("Sequential: sequential=%v", config.Sequential)
	assessLog.Debugf("Debug: debug=%v", config.Debug)

	// Load profile from specified path
	profile, err := oscal.LoadProfile(config.ProfilePath)
	if err != nil {
		assessLog.Errorf("Error: Failed to load profile: %s", config.ProfilePath)
		assessLog.Errorf("Error details: %v", err)
		assessLog.Error("")
		assessLog.Error("Create a profile first:")
		assessLog.Error("  create risk-profile <assessment.md>")
		return 1
	}

	controlIDs := oscal.GetProfileControlIDs(profile)
	assessLog.Infof("Loaded profile with %d controls: %s", len(controlIDs), strings.Join(controlIDs, ", "))
	assessLog.Info("")

	// Run assessments (parallel or sequential)
	var results []*ModuleAssessmentResult

	if config.Sequential || len(config.Modules) == 1 {
		// Sequential: single module or --sequential flag
		if len(config.Modules) == 1 {
			assessLog.Infof("Assessing module: %s", config.Modules[0])
		} else {
			assessLog.Infof("Assessing %d modules sequentially...", len(config.Modules))
		}
		results = assessModulesSequential(config, profile)
	} else {
		// Parallel: multiple modules (default)
		assessLog.Infof("Assessing %d modules in parallel...", len(config.Modules))
		results = assessModulesParallel(config, profile)
	}

	// Check for errors and separate successful results
	var successfulResults []*ModuleAssessmentResult
	var failedModules []string

	for _, result := range results {
		if result.Error != nil {
			failedModules = append(failedModules, result.Module)
			assessLog.Errorf("Module %s failed: %v", result.Module, result.Error)
		} else {
			successfulResults = append(successfulResults, result)
			// Display warnings for this module if any
			if len(result.Warnings) > 0 {
				for _, warning := range result.Warnings {
					assessLog.Warnf("[%s] %s", result.Module, warning)
				}
			}
		}
	}

	// Generate AI-powered risk assessment (always generate executive summary)
	var aiOutput *AIRiskAssessmentOutput
	assessLog.Info("")
	assessLog.Info("🤖 Generating AI-powered risk assessment...")

	aiInput := buildAIRiskAssessmentInput(config, successfulResults, profile)

	aiResult, err := GenerateRiskAssessment(context.Background(), config, aiInput)
	if err != nil {
		assessLog.Warnf("⚠️  AI risk assessment failed: %v", err)
		if len(successfulResults) > 0 {
			assessLog.Info("Falling back to basic risk scoring...")
			// Apply basic fallback scoring for successful results
			applyBasicRiskScoring(successfulResults)
		}
		aiOutput = nil // No AI output available
	} else {
		assessLog.Info("✅ AI risk assessment completed")
		if len(successfulResults) > 0 {
			// Apply AI-generated risk scores to modules (if any)
			ApplyAIRiskAssessment(successfulResults, aiResult, assessLog)
		}
		aiOutput = aiResult // Store for report generation
	}

	// Create aggregated report file
	if len(successfulResults) > 0 {
		if err := writeAggregatedReport(config, successfulResults, profile, aiOutput); err != nil {
			assessLog.Errorf("⚠️  Failed to write aggregated report: %v", err)
			assessLog.Error("Individual module reports are still available in module subdirectories")
		}
	}

	// Report aggregate results
	if len(successfulResults) > 0 {
		reportAggregateResults(config, successfulResults)
	}

	// Handle partial or total failure
	if len(failedModules) > 0 {
		assessLog.Error("")
		assessLog.Errorf("⚠️  %d module(s) failed: %s", len(failedModules), strings.Join(failedModules, ", "))

		if len(successfulResults) == 0 {
			assessLog.Error("All assessments failed")
			return 1
		}

		assessLog.Infof("✓ %d module(s) completed successfully", len(successfulResults))
		// Return success if at least some modules passed
	} else if len(successfulResults) > 0 {
		// All modules succeeded
		assessLog.Info("")
		assessLog.Infof("✓ %d module(s) completed successfully", len(successfulResults))
	}

	assessLog.Info("Assessment completed")
	assessLog.Debugf("Assessment completed: successful=%d, failed=%d", len(successfulResults), len(failedModules))

	return 0
}
