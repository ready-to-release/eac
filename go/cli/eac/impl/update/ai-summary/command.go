// Command: update ai-summary
// Short: Generate AI-powered analysis summaries for modules
// Long: Analyzes modules using AI to generate comprehensive status summaries.
// Long: The analysis covers architecture (DSL), specifications (Gherkin),
// Long: source code, and documentation. Each analysis type produces a
// Long: separate status file in the output directory.
// Long:
// Long: Analysis types:
// Long:   - ai-dsl: Analyzes Structurizr DSL architecture files
// Long:   - ai-specs: Analyzes Gherkin BDD specifications
// Long:   - ai-source: Analyzes source code (depends on dsl and specs)
// Long:   - ai-docs: Analyzes documentation files
// Long:
// Long: Expected Output:
// Long:   - out/ai-summary/<module>/dsl-status.md
// Long:   - out/ai-summary/<module>/specs-status.md
// Long:   - out/ai-summary/<module>/source-status.md
// Long:   - out/ai-summary/<module>/docs-status.md
// Long:
// Long: Example:
// Long:   update ai-summary                  # Analyze all modules
// Long:   update ai-summary eac          # Analyze single module
// Long:   update ai-summary --type=dsl core  # Analyze only DSL for core
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
// Flag.type: type=string, shorthand=t, default=, usage=Specific analysis type (dsl, specs, source, docs)
// Flag.skip-cache: type=bool, usage=Force regeneration even if cached
// Flag.dry-run: type=bool, usage=Show what would be analyzed without executing
// Flag.tui: type=bool, usage=Enable TUI console
// Flag.no-tui: type=bool, usage=Disable TUI console
// Flag.turbo: type=bool, usage=Enable turbo mode (+2 parallel workers)
// Usage: update ai-summary [module...]
package aisummary

import (
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/environment"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

var log = logging.C()

func init() {
	registry.Register(UpdateAISummary)
}

// UpdateAISummary is the entry point for the ai-summary command.
func UpdateAISummary() int {
	args := os.Args[3:] // Skip program name, "update", and "ai-summary"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printUsage()
		return 0
	}

	// Detect execution environment
	env := environment.Detect()

	// Parse flags
	config, err := parseConfig(args, env)
	if err != nil {
		log.Errorf("Error: %v", err)
		printUsage()
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		log.Errorf("Error: failed to load module contracts: %v", err)
		return 1
	}

	// Expand monikers to all modules if none specified
	monikers := config.Monikers
	if len(monikers) == 0 {
		for _, module := range moduleReport.Registry.All() {
			monikers = append(monikers, module.Moniker)
		}
	}

	// Validate specified monikers
	for _, moniker := range monikers {
		if _, exists := moduleReport.Registry.Get(moniker); !exists {
			log.Errorf("Error: module not found: %s", moniker)
			return 1
		}
	}

	// Create command config for framework
	cmdCfg := &cmdframework.CommandConfig{
		Type:           core.ActionAISummary,
		CommandPath:    "update ai-summary",
		OutputDir:      paths.OutAISummaryRelPath,
		Monikers:       monikers,
		SkipDeps:       true, // AI analysis doesn't need system deps
		SkipDepm:       true, // No module dependencies
		ForceRebuild:   config.SkipCache,
		MaxConcurrency: 2, // AI tasks are token-limited, not CPU/RAM-limited
		Turbo:          config.Turbo,
		DryRun:         config.DryRun,
		UseTUI:         config.UseTUI,
		TUIHeight:      6,
		DebugMode:      config.Debug,
	}

	// Store analysis type filter
	cmdCfg.AnalysisType = config.AnalysisType

	return RunAISummaryWithFramework(cmdCfg)
}

// Config holds configuration for the ai-summary command.
type Config struct {
	Monikers     []string
	AnalysisType string // Empty = all types, or specific type (dsl, specs, source, docs)
	Debug        bool
	SkipCache    bool
	DryRun       bool
	UseTUI       bool
	Turbo        bool
}

// parseConfig parses command line arguments into Config.
func parseConfig(args []string, env *environment.Env) (*Config, error) {
	config := &Config{
		UseTUI: env.IsLocalConsole, // TUI on by default for local
	}

	var positionalArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--debug" || arg == "-d":
			config.Debug = true

		case arg == "--skip-cache":
			config.SkipCache = true

		case arg == "--dry-run":
			config.DryRun = true

		case arg == "--tui":
			config.UseTUI = true

		case arg == "--no-tui":
			config.UseTUI = false

		case arg == "--turbo":
			config.Turbo = true

		case strings.HasPrefix(arg, "--type="):
			config.AnalysisType = strings.TrimPrefix(arg, "--type=")
			if !isValidAnalysisType(config.AnalysisType) {
				return nil, fmt.Errorf("invalid analysis type: %s (valid: dsl, specs, source, docs)", config.AnalysisType)
			}

		case arg == "--type" || arg == "-t":
			if i+1 < len(args) {
				config.AnalysisType = args[i+1]
				i++
				if !isValidAnalysisType(config.AnalysisType) {
					return nil, fmt.Errorf("invalid analysis type: %s (valid: dsl, specs, source, docs)", config.AnalysisType)
				}
			}

		case !strings.HasPrefix(arg, "-"):
			positionalArgs = append(positionalArgs, arg)

		default:
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	config.Monikers = positionalArgs
	return config, nil
}

// isValidAnalysisType checks if the analysis type is valid.
func isValidAnalysisType(t string) bool {
	switch t {
	case "", "dsl", "specs", "source", "docs":
		return true
	default:
		return false
	}
}

// parseDebugFlag checks for debug flag in args (compatibility with shared flags).
func parseDebugFlag(args []string) bool {
	return flags.ParseDebugFlag(args)
}

func printUsage() {
	log.Info("Generate AI-powered analysis summaries for modules")
	log.Info("")
	log.Info("Usage: update ai-summary [flags] [module...]")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  module...     Module monikers to analyze (analyzes all if none specified)")
	log.Info("")
	log.Info("Flags:")
	log.Info("  -d, --debug       Enable debug logging")
	log.Info("  -t, --type TYPE   Specific analysis type (dsl, specs, source, docs)")
	log.Info("  --skip-cache      Force regeneration even if cached")
	log.Info("  --dry-run         Show what would be analyzed without executing")
	log.Info("  --tui             Enable TUI console (default for local)")
	log.Info("  --no-tui          Disable TUI console")
	log.Info("  --turbo           Enable turbo mode (+2 parallel workers)")
	log.Info("  -h, --help        Show this help message")
	log.Info("")
	log.Info("Analysis Types:")
	log.Info("  dsl       Analyze Structurizr DSL architecture files")
	log.Info("  specs     Analyze Gherkin BDD specifications")
	log.Info("  source    Analyze source code (depends on dsl and specs)")
	log.Info("  docs      Analyze documentation files")
	log.Info("")
	log.Info("Output:")
	log.Info("  out/ai-summary/<module>/dsl-status.md")
	log.Info("  out/ai-summary/<module>/specs-status.md")
	log.Info("  out/ai-summary/<module>/source-status.md")
	log.Info("  out/ai-summary/<module>/docs-status.md")
	log.Info("")
	log.Info("Examples:")
	log.Info("  update ai-summary                   # Analyze all modules")
	log.Info("  update ai-summary eac           # Analyze single module")
	log.Info("  update ai-summary --type=dsl core   # Analyze only DSL for core")
}
