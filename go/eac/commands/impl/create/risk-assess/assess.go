// Command: create risk-assess
// Short: Update OSCAL assessment-results with test and security evidence
// Long: The create risk-assess command creates or updates OSCAL assessment-results for modules
// Long: by collecting test results and security scan evidence. It maps @control tags in
// Long: feature files to OSCAL control IDs and determines satisfied/not-satisfied status.
// Long:
// Long: Evidence is automatically collected from:
// Long: - Test results: out/test/<timestamp>/<module>/*.cucumber.json
// Long: - Security scans: out/security/<module>/**/*.json
// Long:
// Long: If evidence is older than 24 hours, tests and scans are automatically re-run.
// Flag.profile: type=string, shorthand=p, required=true, usage=Path to OSCAL profile JSON file
// Flag.max-evidence-age: type=string, default=24h, usage=Maximum age for evidence before auto-refresh (e.g., 24h, 7d)
// Flag.force-tests: type=bool, default=false, usage=Force re-run tests regardless of age
// Flag.force-security: type=bool, default=false, usage=Force re-run security scans regardless of age
// Flag.skip-auto-run: type=bool, default=false, usage=Use existing evidence only, fail if missing
// Flag.sequential: type=bool, default=false, usage=Run assessments sequentially instead of parallel
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs to out/logs/risk/
// Args: modules
package riskassess

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/scoring"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var assessLog = logging.C()

func init() {
	registry.Register(CreateRiskAssess)
}

// AssessConfig holds configuration for risk assess command.
type AssessConfig struct {
	Modules        []string // Module names to assess
	ProfilePath    string
	MaxEvidenceAge time.Duration
	ForceTests     bool
	ForceSecurity  bool
	SkipAutoRun    bool
	Sequential     bool // Disable parallel execution
	Debug          bool
	WorkspaceRoot  string
	Logger         *logging.Logger
}

// ModuleAssessmentResult holds the results of assessing a single module.
type ModuleAssessmentResult struct {
	Module            string
	AssessmentResults *oscalTypes.AssessmentResults
	OutputPath        string
	Satisfied         int
	NotSatisfied      int
	RiskScore         *scoring.RiskScore
	Error             error
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
			zap.Strings("modules", config.Modules),
			zap.String("profile", config.ProfilePath),
			zap.Duration("max_evidence_age", config.MaxEvidenceAge),
			zap.Bool("sequential", config.Sequential),
			zap.Bool("debug", config.Debug))
	}

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
	}

	if config.Logger != nil {
		config.Logger.Info("Assessment completed",
			zap.Int("successful", len(successfulResults)),
			zap.Int("failed", len(failedModules)))
	}

	return 0
}

// loadModuleRegistry loads the module registry from the workspace.
func loadModuleRegistry(workspaceRoot string) (*modules.Registry, error) {
	return modules.LoadFromWorkspace(workspaceRoot)
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

	// Collect positional arguments (module names) before flags
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Stop at first flag
		if strings.HasPrefix(arg, "-") {
			break
		}

		// Collect module name
		config.Modules = append(config.Modules, arg)
	}

	// Parse flags starting from where modules ended
	i := len(config.Modules)
	for i < len(args) {
		arg := args[i]

		switch {
		case arg == "--help" || arg == "-h":
			return nil, fmt.Errorf("help requested")

		case arg == "--profile" || arg == "-p":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--profile requires a value")
			}
			config.ProfilePath = args[i+1]
			// Make path absolute if relative
			if !filepath.IsAbs(config.ProfilePath) {
				config.ProfilePath = filepath.Join(config.WorkspaceRoot, config.ProfilePath)
			}
			i += 2

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

		case arg == "--sequential":
			config.Sequential = true
			i++

		case arg == "--debug" || arg == "-d":
			config.Debug = true
			i++

		case strings.HasPrefix(arg, "-"):
			return nil, fmt.Errorf("unknown flag: %s", arg)

		default:
			return nil, fmt.Errorf("unexpected positional argument after flags: %s", arg)
		}
	}

	// If no modules specified, discover all modules
	if len(config.Modules) == 0 {
		registry, err := loadModuleRegistry(workspaceRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to load modules: %w", err)
		}

		allModules := registry.All()
		for _, mod := range allModules {
			config.Modules = append(config.Modules, mod.Moniker)
		}

		if len(config.Modules) == 0 {
			return nil, fmt.Errorf("no modules found in workspace")
		}
	}

	// Validate required flags
	if config.ProfilePath == "" {
		return nil, fmt.Errorf("--profile flag is required")
	}

	return config, nil
}

// showAssessHelp displays help information.
func showAssessHelp() {
	help := `Usage: create risk-assess [module...] --profile <path> [flags]

Update OSCAL assessment-results with test and security evidence

Arguments:
  module...            Module name(s) to assess (space-separated)
                       If omitted, all modules are assessed

Required Flags:
  -p, --profile <path>               Path to OSCAL profile JSON file

Optional Flags:
      --max-evidence-age <duration>  Maximum age for evidence (default: 24h)
      --force-tests                  Force re-run tests regardless of age
      --force-security               Force re-run security scans regardless of age
      --skip-auto-run                Use existing evidence only, fail if missing
      --sequential                   Run assessments sequentially (default: parallel)
  -d, --debug                        Save intermediate outputs to out/logs/risk/

Output:
  out/risk/<module>/assessment-results.json (per module)

Examples:
  # Assess all modules with shared profile (parallel by default)
  create risk-assess --profile specs/.risk-controls/risk-profile.json

  # Assess specific modules (parallel, space-separated)
  create risk-assess billing api auth --profile specs/.risk-controls/risk-profile.json

  # Assess single module
  create risk-assess billing --profile specs/.risk-controls/risk-profile.json

  # Sequential execution for debugging
  create risk-assess --profile specs/.risk-controls/risk-profile.json --sequential

  # Force fresh evidence for all modules
  create risk-assess --profile specs/.risk-controls/risk-profile.json --force-tests

  # Set custom evidence freshness threshold
  create risk-assess data --profile specs/.risk-controls/risk-profile.json --max-evidence-age 1h
`
	assessLog.Info(help)
}
